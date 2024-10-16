package coldmfa

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"github.com/EphyraSoftware/locus/auth"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"image/jpeg"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

//go:embed migrations/*
var migrations embed.FS

type App struct {
	Router      fiber.Router
	DatabaseUrl string
	Public      embed.FS
}

func (a *App) Prepare() {
	source, err := iofs.New(migrations, "migrations")
	if err != nil {
		log.Fatalf("failed to prepare migration source: %s", err.Error())
	}
	m, err := migrate.NewWithSourceInstance("io/fs", source, a.DatabaseUrl)
	if err != nil {
		log.Fatalf("failed to configure migration: %s", err.Error())
	}
	err = m.Migrate(1)
	if err != nil && err.Error() != "no change" {
		log.Fatalf("failed to run migration: %s", err.Error())
	}

	db, err := sql.Open("postgres", a.DatabaseUrl)
	if err != nil {
		log.Fatal(err)
	}

	a.Router.Use(filesystem.New(filesystem.Config{
		Root:       http.FS(a.Public),
		PathPrefix: "public/coldmfa",
		Index:      "index.html",
	}))

	api := a.Router.Group("/api")

	api.Get("/user", func(c *fiber.Ctx) error {
		sessionUser := auth.SessionUser(c)
		if sessionUser == "" {
			return c.SendStatus(http.StatusUnauthorized)
		}

		return c.Status(200).JSON(fiber.Map{
			"user": sessionUser,
		})
	})

	api.Get("/groups", func(c *fiber.Ctx) error {
		sessionId := auth.SessionId(c)
		if sessionId == "" {
			return c.SendStatus(http.StatusUnauthorized)
		}

		rows, err := db.Query("SELECT group_id, name FROM code_group where owner_id = $1", sessionId)
		if err != nil {
			log.Errorf("failed to query groups: %s", err.Error())
			return c.Status(http.StatusInternalServerError).JSON(ApiError{Error: "database error"})
		}

		out := make([]CodeGroup, 0)
		for rows.Next() {
			var group CodeGroup
			err = rows.Scan(&group.GroupId, &group.Name)
			if err != nil {
				log.Errorf("failed to scan group: %s", err.Error())
				return c.Status(http.StatusInternalServerError).JSON(ApiError{Error: "database error"})
			}
			out = append(out, group)
		}

		return c.Status(200).JSON(out)
	})

	api.Get("/groups/:id", func(c *fiber.Ctx) error {
		sessionId := auth.SessionId(c)
		if sessionId == "" {
			return c.SendStatus(http.StatusUnauthorized)
		}

		groupId := c.Params("id")
		if groupId == "" {
			return c.Status(http.StatusBadRequest).JSON(ApiError{Error: "missing id"})
		}

		codeGroup, err := readCodeGroup(c.Context(), db, sessionId, groupId)
		if err != nil {
			log.Errorf("failed to read group: %s", err.Error())
			return c.Status(http.StatusNotFound).JSON(ApiError{Error: "group not found"})
		}

		rows, err := db.QueryContext(c.Context(), "select code_id, name, preferred_name, created_at, deleted, deleted_at from code where code_group_id = (select id from code_group where owner_id = $1 and group_id = $2)", sessionId, groupId)
		if err != nil {
			log.Errorf("failed to query codes: %s", err.Error())
			return c.Status(http.StatusInternalServerError).JSON(ApiError{Error: "database error"})
		}

		codeGroup.Codes = make([]CodeSummary, 0)
		for rows.Next() {
			var code CodeSummary
			err = rows.Scan(&code.CodeId, &code.Name, &code.PreferredName, &code.CreatedAt, &code.Deleted, &code.DeletedAt)
			if err != nil {
				log.Errorf("failed to scan code: %s", err.Error())
				return c.Status(http.StatusInternalServerError).JSON(ApiError{Error: "database error"})
			}
			codeGroup.Codes = append(codeGroup.Codes, code)
		}

		return c.Status(200).JSON(codeGroup)
	})

	api.Post("/groups", func(c *fiber.Ctx) error {
		sessionId := auth.SessionId(c)
		if sessionId == "" {
			return c.SendStatus(http.StatusUnauthorized)
		}

		codeGroup := new(CodeGroup)
		if err := c.BodyParser(codeGroup); err != nil {
			return c.Status(http.StatusBadRequest).JSON(ApiError{Error: "invalid request"})
		}

		if len(codeGroup.Name) < 3 {
			return c.Status(http.StatusBadRequest).JSON(ApiError{Error: "name too short"})
		}

		groupId, err := gonanoid.New()
		if err != nil {
			log.Errorf("failed to generate group id: %s", err.Error())
			return c.Status(http.StatusInternalServerError).JSON(ApiError{Error: "internal error"})
		}

		_, err = db.Exec("insert into code_group (owner_id, group_id, name) values ($1, $2, $3)", sessionId, groupId, codeGroup.Name)
		if err != nil {
			log.Errorf("failed to insert group: %s", err.Error())
			return c.Status(http.StatusInternalServerError).JSON(ApiError{Error: "database error"})
		}

		createdCodeGroup, err := readCodeGroup(c.Context(), db, sessionId, groupId)
		if err != nil {
			log.Errorf("failed to read group: %s", err.Error())
			return c.Status(http.StatusInternalServerError).JSON(ApiError{Error: "group not found"})
		}

		return c.Status(http.StatusCreated).JSON(createdCodeGroup)
	})

	api.Post("/groups/:groupId/codes", func(c *fiber.Ctx) error {
		sessionId := auth.SessionId(c)
		if sessionId == "" {
			return c.SendStatus(http.StatusUnauthorized)
		}

		groupId := c.Params("groupId")
		if groupId == "" {
			return c.Status(http.StatusBadRequest).JSON(ApiError{Error: "missing groupId"})
		}

		createCode := new(CreateCode)
		if err := c.BodyParser(createCode); err != nil {
			return c.Status(http.StatusBadRequest).JSON(ApiError{Error: "invalid request"})
		}

		otpUrl, err := url.Parse(createCode.Original)
		if err != nil {
			log.Errorf("failed to parse otp url: %s", err.Error())
			return c.Status(http.StatusBadRequest).JSON(ApiError{Error: "invalid request"})
		}
		parts := strings.Split(otpUrl.Path, "/")

		firstPart := ""
		for _, part := range parts {
			if part != "" {
				firstPart = part
				break
			}
		}
		name := firstPart

		secret := otpUrl.Query().Get("secret")
		if secret == "" {
			log.Errorf("missing secret in otp url")
			return c.Status(http.StatusBadRequest).JSON(ApiError{Error: "invalid request"})
		}

		_, err = totp.GenerateCode(secret, time.Now())
		if err != nil {
			log.Errorf("failed to generate code: %s", err.Error())
			return c.Status(http.StatusBadRequest).JSON(ApiError{Error: "invalid or unsupported opt provided"})
		}

		codeId, err := gonanoid.New()
		if err != nil {
			log.Errorf("failed to generate code id: %s", err.Error())
			return c.Status(http.StatusInternalServerError).JSON(ApiError{Error: "internal error"})
		}

		_, err = db.Exec("insert into code (code_group_id, code_id, original, name) values ((select id from code_group where owner_id = $1 and group_id = $2), $3, $4, $5)", sessionId, groupId, codeId, createCode.Original, name)
		if err != nil {
			log.Errorf("failed to insert code: %s", err.Error())
			return c.Status(http.StatusInternalServerError).JSON(ApiError{Error: "database error"})
		}

		createdCode, err := readCodeSummary(db, sessionId, groupId, codeId)
		if err != nil {
			log.Errorf("failed to read code: %s", err.Error())
			return c.Status(http.StatusInternalServerError).JSON(ApiError{Error: "code not found"})
		}

		return c.Status(http.StatusCreated).JSON(createdCode)
	})

	api.Get("/groups/:groupId/codes/:codeId", func(c *fiber.Ctx) error {
		sessionId := auth.SessionId(c)
		if sessionId == "" {
			return c.SendStatus(http.StatusUnauthorized)
		}

		groupId := c.Params("groupId")
		if groupId == "" {
			return c.Status(http.StatusBadRequest).JSON(ApiError{Error: "missing groupId"})
		}

		codeId := c.Params("codeId")
		if codeId == "" {
			return c.Status(http.StatusBadRequest).JSON(ApiError{Error: "missing codeId"})
		}

		row := db.QueryRow("select original from code where code_group_id = (select id from code_group where owner_id = $1 and group_id = $2) and code_id = $3", sessionId, groupId, codeId)
		if row == nil {
			return c.Status(http.StatusNotFound).JSON(ApiError{Error: "code not found"})
		}

		var original string
		err := row.Scan(&original)
		if err != nil {
			log.Errorf("failed to scan code: %s", err.Error())
			return c.Status(http.StatusInternalServerError).JSON(ApiError{Error: "database error"})
		}

		otpConfig, err := extractOtpAuthUrl(original)
		if err != nil {
			log.Errorf("failed to extract otp config: %s", err.Error())
			return c.Status(http.StatusInternalServerError).JSON(ApiError{Error: "internal error"})
		}

		opts, err := otpConfig.toOpts()
		if err != nil {
			log.Errorf("failed to convert otp config: %s", err.Error())
			return c.Status(http.StatusInternalServerError).JSON(ApiError{Error: "internal error"})
		}
		now := time.Now()
		passcodeNow, err := totp.GenerateCodeCustom(otpConfig.Secret, now, *opts)
		if err != nil {
			log.Errorf("failed to generate code: %s", err.Error())
			return c.Status(http.StatusInternalServerError).JSON(ApiError{Error: "internal error"})
		}

		later := now.Add(time.Duration(opts.Period) * time.Second)
		passcodeLater, err := totp.GenerateCodeCustom(otpConfig.Secret, later, *opts)
		if err != nil {
			log.Errorf("failed to generate code: %s", err.Error())
			return c.Status(http.StatusInternalServerError).JSON(ApiError{Error: "internal error"})
		}

		return c.Status(200).JSON(PasscodeResponse{
			Passcode:     passcodeNow,
			NextPasscode: passcodeLater,
			ServerTime:   now.Unix(),
			Period:       opts.Period,
		})
	})

	api.Put("/groups/:groupId/codes/:codeId", func(c *fiber.Ctx) error {
		sessionId := auth.SessionId(c)
		if sessionId == "" {
			return c.SendStatus(http.StatusUnauthorized)
		}

		groupId := c.Params("groupId")
		if groupId == "" {
			return c.Status(http.StatusBadRequest).JSON(ApiError{Error: "missing groupId"})
		}

		codeId := c.Params("codeId")
		if codeId == "" {
			return c.Status(http.StatusBadRequest).JSON(ApiError{Error: "missing codeId"})
		}

		codeSummary := new(CodeSummary)
		if err := c.BodyParser(codeSummary); err != nil {
			return c.Status(http.StatusBadRequest).JSON(ApiError{Error: "invalid request"})
		}

		var result sql.Result
		if codeSummary.PreferredName != nil && strings.TrimSpace(*codeSummary.PreferredName) == "" {
			result, err = db.ExecContext(c.Context(), "update code set preferred_name = NULL where deleted = false and code_group_id = (select id from code_group where owner_id = $1 and group_id = $2) and code_id = $3", sessionId, groupId, codeId)
		} else {
			setName := strings.TrimSpace(*codeSummary.PreferredName)
			result, err = db.ExecContext(c.Context(), "update code set preferred_name = $4 where deleted = false and code_group_id = (select id from code_group where owner_id = $1 and group_id = $2) and code_id = $3", sessionId, groupId, codeId, setName)
		}

		if err != nil {
			log.Errorf("failed to update code: %s", err.Error())
			return c.Status(http.StatusInternalServerError).JSON(ApiError{Error: "database error"})
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			log.Errorf("failed to update code: %s", err.Error())
			return c.Status(http.StatusInternalServerError).JSON(ApiError{Error: "database error"})
		}

		if rowsAffected == 0 {
			return c.Status(http.StatusNotFound).JSON(ApiError{Error: "code not found"})
		}

		return c.SendStatus(http.StatusNoContent)
	})

	api.Post("/groups/:groupId/codes/:codeId/move", func(c *fiber.Ctx) error {
		sessionId := auth.SessionId(c)
		if sessionId == "" {
			return c.SendStatus(http.StatusUnauthorized)
		}

		currentGroupId := c.Params("groupId")
		if currentGroupId == "" {
			return c.Status(http.StatusBadRequest).JSON(ApiError{Error: "missing groupId"})
		}

		codeId := c.Params("codeId")
		if codeId == "" {
			return c.Status(http.StatusBadRequest).JSON(ApiError{Error: "missing codeId"})
		}

		moveCodeRequest := new(MoveCodeRequest)
		if err := c.BodyParser(moveCodeRequest); err != nil {
			return c.Status(http.StatusBadRequest).JSON(ApiError{Error: "invalid request"})
		}

		if moveCodeRequest.ToGroupId == "" {
			return c.Status(http.StatusBadRequest).JSON(ApiError{Error: "missing toGroupId"})
		}

		result, err := db.ExecContext(c.Context(), "update code set code_group_id = (select id from code_group where owner_id = $1 and group_id = $2) where deleted = false and code_group_id = (select id from code_group where owner_id = $1 and group_id = $3) and code_id = $4", sessionId, moveCodeRequest.ToGroupId, currentGroupId, codeId)

		if err != nil {
			log.Errorf("failed to move code: %s", err.Error())
			return c.Status(http.StatusInternalServerError).JSON(ApiError{Error: "database error"})
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			log.Errorf("failed to update code: %s", err.Error())
			return c.Status(http.StatusInternalServerError).JSON(ApiError{Error: "database error"})
		}

		if rowsAffected == 0 {
			return c.Status(http.StatusNotFound).JSON(ApiError{Error: "code or target group not found"})
		}

		return c.SendStatus(http.StatusNoContent)
	})

	api.Delete("/groups/:groupId/codes/:codeId", func(c *fiber.Ctx) error {
		sessionId := auth.SessionId(c)
		if sessionId == "" {
			return c.SendStatus(http.StatusUnauthorized)
		}

		groupId := c.Params("groupId")
		if groupId == "" {
			return c.Status(http.StatusBadRequest).JSON(ApiError{Error: "missing groupId"})
		}

		codeId := c.Params("codeId")
		if codeId == "" {
			return c.Status(http.StatusBadRequest).JSON(ApiError{Error: "missing codeId"})
		}

		result, err := db.ExecContext(c.Context(), "update code set deleted = true, deleted_at = now() where deleted = false and code_group_id = (select id from code_group where owner_id = $1 and group_id = $2) and code_id = $3", sessionId, groupId, codeId)
		if err != nil {
			log.Errorf("failed to delete code: %s", err.Error())
			return c.Status(http.StatusInternalServerError).JSON(ApiError{Error: "database error"})
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			log.Errorf("failed to delete code: %s", err.Error())
			return c.Status(http.StatusInternalServerError).JSON(ApiError{Error: "database error"})
		}

		if rowsAffected == 0 {
			return c.Status(http.StatusNotFound).JSON(ApiError{Error: "code not found"})
		}

		return c.SendStatus(http.StatusNoContent)
	})

	api.Get("/groups/:groupId/codes/:codeId/qr", func(c *fiber.Ctx) error {
		sessionId := auth.SessionId(c)
		if sessionId == "" {
			return c.SendStatus(http.StatusUnauthorized)
		}

		groupId := c.Params("groupId")
		if groupId == "" {
			return c.Status(http.StatusBadRequest).JSON(ApiError{Error: "missing groupId"})
		}

		codeId := c.Params("codeId")
		if codeId == "" {
			return c.Status(http.StatusBadRequest).JSON(ApiError{Error: "missing codeId"})
		}

		row := db.QueryRow("select original from code where code_group_id = (select id from code_group where owner_id = $1 and group_id = $2) and code_id = $3", sessionId, groupId, codeId)
		if row == nil {
			return c.Status(http.StatusNotFound).JSON(ApiError{Error: "code not found"})
		}

		var original string
		err := row.Scan(&original)
		if err != nil {
			log.Errorf("failed to scan code: %s", err.Error())
			return c.Status(http.StatusInternalServerError).JSON(ApiError{Error: "database error"})
		}

		key, err := otp.NewKeyFromURL(original)
		if err != nil {
			log.Errorf("failed to parse key: %s", err.Error())
			return c.Status(http.StatusInternalServerError).JSON(ApiError{Error: "internal error"})
		}

		image, err := key.Image(250, 250)
		if err != nil {
			log.Errorf("failed to generate qr: %s", err.Error())
			return c.Status(http.StatusInternalServerError).JSON(ApiError{Error: "internal error"})
		}

		reader, writer := io.Pipe()
		go func() {
			defer func(writer *io.PipeWriter) {
				err := writer.Close()
				if err != nil {
					log.Errorf("failed to close pipe: %s", err.Error())
				}
			}(writer)
			err = jpeg.Encode(writer, image, nil)
			if err != nil {
				log.Errorf("failed to encode qr: %s", err.Error())
			}
		}()

		c.GetRespHeaders()["Content-Type"] = []string{"image/jpeg"}
		return c.Status(200).SendStream(reader)
	})

	api.Post("/backups", func(c *fiber.Ctx) error {
		sessionId := auth.SessionId(c)
		if sessionId == "" {
			return c.SendStatus(http.StatusUnauthorized)
		}

		backupRequest := new(BackupRequest)
		if err := c.BodyParser(backupRequest); err != nil {
			return c.Status(http.StatusBadRequest).JSON(ApiError{Error: "invalid request"})
		}

		rows, err := db.Query("select code_group.name, code.original, code.name as code_name, code.preferred_name, code.created_at, code.deleted, code.deleted_at from code_group left join code on code.code_group_id = code_group.id where owner_id = $1", sessionId)
		if err != nil {
			log.Errorf("failed to query backups: %s", err.Error())
			return c.Status(http.StatusInternalServerError).JSON(ApiError{Error: "database error"})
		}

		backupItems := make([]BackupItem, 0)
		for rows.Next() {
			var item BackupItem
			err = rows.Scan(&item.GroupName, &item.Original, &item.CodeName, &item.PreferredName, &item.CreatedAt, &item.Deleted, &item.DeletedAt)
			if err != nil {
				log.Errorf("failed to scan backup item: %s", err.Error())
				return c.Status(http.StatusInternalServerError).JSON(ApiError{Error: "database error"})
			}
			backupItems = append(backupItems, item)
		}

		var backupContent []string
		for _, item := range backupItems {
			it, err := json.Marshal(item)
			if err != nil {
				log.Errorf("failed to marshal backup item: %s", err.Error())
				return c.Status(http.StatusInternalServerError).JSON(ApiError{Error: "internal error"})
			}
			backupContent = append(backupContent, string(it))
		}

		encrypted, err := EncryptMfaCodeBackupItems(backupContent, backupRequest.Password)
		if err != nil {
			log.Errorf("failed to encrypt backup: %s", err.Error())
			return c.Status(http.StatusInternalServerError).JSON(ApiError{Error: "internal error"})
		}

		// As a last step before returning the encrypted backup, record the backup time in the database
		_, err = db.Exec("insert into last_backup (owner_id, backup_at) values ($1, now()) on conflict on constraint owner_id_unique do update set backup_at = now()", sessionId)
		if err != nil {
			log.Errorf("failed to record backup: %s", err.Error())
			return c.Status(http.StatusInternalServerError).JSON(ApiError{Error: "database error"})
		}

		return c.Status(http.StatusOK).Send(encrypted)
	})

	api.Put("/backups", func(c *fiber.Ctx) error {
		sessionId := auth.SessionId(c)
		if sessionId == "" {
			return c.SendStatus(http.StatusUnauthorized)
		}

		restoreBackupRequest := new(RestoreBackupRequest)
		if err := c.BodyParser(restoreBackupRequest); err != nil {
			log.Info("failed to parse body")
			return c.Status(http.StatusBadRequest).JSON(ApiError{Error: "invalid request"})
		}

		decrypted, err := DecryptMfaCodeBackupItems(restoreBackupRequest.BackupContent, restoreBackupRequest.Password)
		if err != nil {
			log.Errorf("failed to decrypt backup: %s", err.Error())
			return c.Status(http.StatusBadRequest).JSON(ApiError{Error: "invalid backup"})
		}

		tx, err := db.BeginTx(c.Context(), nil)
		if err != nil {
			log.Errorf("failed to start transaction: %s", err.Error())
			return c.Status(http.StatusInternalServerError).JSON(ApiError{Error: "database error"})
		}

		defer func(tx *sql.Tx) {
			err := tx.Rollback()
			if err != nil && !errors.Is(err, sql.ErrTxDone) {
				log.Errorf("failed to rollback transaction: %s", err.Error())
			}
		}(tx)

		for _, item := range decrypted {
			var backupItem BackupItem
			err := json.Unmarshal([]byte(item), &backupItem)
			if err != nil {
				log.Errorf("failed to unmarshal backup item: %s", err.Error())
				return c.Status(http.StatusInternalServerError).JSON(ApiError{Error: "internal error"})
			}

			groupId, err := gonanoid.New()
			if err != nil {
				log.Errorf("failed to generate group id: %s", err.Error())
				return c.Status(http.StatusInternalServerError).JSON(ApiError{Error: "internal error"})
			}

			_, err = tx.ExecContext(c.Context(), "insert into code_group (owner_id, group_id, name) values ($1, $2, $3) on conflict on constraint owner_id_name_unique do nothing", sessionId, groupId, backupItem.GroupName)
			if err != nil {
				log.Errorf("failed to insert group: %s", err.Error())
				return c.Status(http.StatusInternalServerError).JSON(ApiError{Error: "database error"})
			}

			row := tx.QueryRowContext(c.Context(), "select id from code_group where owner_id = $1 and name = $2", sessionId, backupItem.GroupName)

			var groupDatabaseId int
			err = row.Scan(&groupDatabaseId)
			if err != nil {
				log.Errorf("failed to insert or read group: %s", err.Error())
				return c.Status(http.StatusInternalServerError).JSON(ApiError{Error: "database error"})
			}

			if backupItem.CodeName != nil {
				codeId, err := gonanoid.New()
				if err != nil {
					log.Errorf("failed to generate code id: %s", err.Error())
					return c.Status(http.StatusInternalServerError).JSON(ApiError{Error: "internal error"})
				}

				_, err = tx.ExecContext(c.Context(), "insert into code (code_group_id, code_id, original, name, preferred_name, created_at, deleted, deleted_at) values ($1, $2, $3, $4, $5, $6, $7, $8) on conflict on constraint code_group_id_original_unique do nothing", groupDatabaseId, codeId, backupItem.Original, backupItem.CodeName, backupItem.PreferredName, backupItem.CreatedAt, backupItem.Deleted, backupItem.DeletedAt)
				if err != nil {
					log.Errorf("failed to insert code: %s", err.Error())
					return c.Status(http.StatusInternalServerError).JSON(ApiError{Error: "database error"})
				}
			}
		}

		err = tx.Commit()
		if err != nil {
			log.Errorf("failed to commit transaction: %s", err.Error())
			return c.Status(http.StatusInternalServerError).JSON(ApiError{Error: "database error"})
		}

		return c.SendStatus(http.StatusOK)
	})

	api.Get("/backups/warning", func(c *fiber.Ctx) error {
		sessionId := auth.SessionId(c)
		if sessionId == "" {
			return c.SendStatus(http.StatusUnauthorized)
		}

		lastBackupRow := db.QueryRow("select last_backup.backup_at from last_backup where last_backup.owner_id = $1", sessionId)
		countRow := db.QueryRow("select count(code.id) from last_backup left join code_group on code_group.owner_id = last_backup.owner_id left join code on code.code_group_id = code_group.id where last_backup.owner_id = $1 and code.created_at > last_backup.backup_at group by last_backup.owner_id, last_backup.backup_at", sessionId)

		var warning BackupWarning
		err := lastBackupRow.Scan(&warning.LastBackupAt)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Errorf("failed to read warning: %s", err.Error())
			return c.Status(http.StatusInternalServerError).JSON(ApiError{Error: "database error"})
		}

		err = countRow.Scan(&warning.NumberNotBackedUp)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				warning.NumberNotBackedUp = 0
			} else {
				log.Errorf("failed to read warning: %s", err.Error())
				return c.Status(http.StatusInternalServerError).JSON(ApiError{Error: "database error"})
			}
		}

		return c.Status(http.StatusOK).JSON(warning)
	})
}

func readCodeGroup(context context.Context, db *sql.DB, ownerId string, groupId string) (*CodeGroup, error) {
	row := db.QueryRowContext(context, "select group_id, name from code_group where owner_id = $1 and group_id = $2", ownerId, groupId)
	if row == nil {
		return nil, fmt.Errorf("group not found")
	}

	var group CodeGroup
	err := row.Scan(&group.GroupId, &group.Name)
	return &group, err
}

func readCodeSummary(db *sql.DB, ownerId string, groupId, codeId string) (*CodeSummary, error) {
	row := db.QueryRow("select code_id, name, preferred_name, created_at, deleted, deleted_at from code where code_group_id = (select id from code_group where owner_id = $1 and group_id = $2) and code_id = $3", ownerId, groupId, codeId)
	if row == nil {
		return nil, fmt.Errorf("code not found")
	}

	var code CodeSummary
	err := row.Scan(&code.CodeId, &code.Name, &code.PreferredName, &code.CreatedAt, &code.Deleted, &code.DeletedAt)
	return &code, err
}

type OtpConfig struct {
	Type      string
	Label     string
	Secret    string
	Issuer    *string
	Algorithm *string
	Digits    *int
	Counter   *int
	Period    *uint
}

func extractOtpAuthUrl(raw string) (*OtpConfig, error) {
	otpUrl, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse otp url: %s", err.Error())
	}

	if otpUrl.Scheme != "otpauth" {
		return nil, fmt.Errorf("invalid otp url scheme")
	}

	typ := otpUrl.Host
	if typ != "totp" && typ != "hotp" {
		return nil, fmt.Errorf("unsupported otp type")
	}

	path := otpUrl.Path
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}
	parts := strings.Split(otpUrl.Path, "/")

	if len(parts) < 1 {
		return nil, fmt.Errorf("invalid otp url path")
	}

	label := parts[1]

	secret := otpUrl.Query().Get("secret")
	if secret == "" {
		return nil, fmt.Errorf("missing secret in otp url")
	}

	out := OtpConfig{
		Type:   typ,
		Label:  label,
		Secret: secret,
	}

	issuer := otpUrl.Query().Get("issuer")
	if issuer != "" {
		out.Issuer = &issuer
	}

	algorithm := otpUrl.Query().Get("algorithm")
	if algorithm != "" {
		out.Algorithm = &algorithm
	}

	digits := otpUrl.Query().Get("digits")
	if digits != "" {
		num, err := strconv.Atoi(digits)
		if err != nil {
			return nil, fmt.Errorf("invalid digits")
		}
		out.Digits = &num
	}

	counter := otpUrl.Query().Get("counter")
	if counter != "" {
		num, err := strconv.Atoi(counter)
		if err != nil {
			return nil, fmt.Errorf("invalid counter")
		}
		out.Counter = &num
	} else if typ == "hotp" {
		return nil, fmt.Errorf("missing counter in hotp url")
	}

	period := otpUrl.Query().Get("period")
	if typ == "totp" && period != "" {
		num, err := strconv.Atoi(period)
		if err != nil || num < 0 {
			return nil, fmt.Errorf("invalid period")
		}
		unsignedNum := uint(num)
		out.Period = &unsignedNum
	}

	return &out, nil
}

func (cfg *OtpConfig) toOpts() (*totp.ValidateOpts, error) {
	opts := totp.ValidateOpts{}

	if cfg.Period != nil {
		opts.Period = *cfg.Period
	} else {
		// Documented default, but might not match what is expected by the totp library if
		// that code is updated and this isn't...
		opts.Period = 30
	}

	if cfg.Digits != nil {
		switch *cfg.Digits {
		case 6:
			opts.Digits = otp.DigitsSix
		case 8:
			opts.Digits = otp.DigitsEight
		default:
			return nil, fmt.Errorf("invalid digits")
		}
	}

	if cfg.Algorithm != nil {
		switch *cfg.Algorithm {
		case "SHA1":
			opts.Algorithm = otp.AlgorithmSHA1
		case "SHA256":
			opts.Algorithm = otp.AlgorithmSHA256
		case "SHA512":
			opts.Algorithm = otp.AlgorithmSHA512
		case "MD5":
			opts.Algorithm = otp.AlgorithmMD5
		default:
			return nil, fmt.Errorf("invalid algorithm")
		}
	}

	return &opts, nil
}
