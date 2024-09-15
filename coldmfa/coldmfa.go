package coldmfa

import "C"
import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"github.com/EphyraSoftware/locus/auth"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/pquerna/otp/totp"
	"net/http"
	"net/url"
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

		rows, err := db.QueryContext(c.Context(), "select code_id, name, preferred_name from code where code_group_id = (select id from code_group where owner_id = $1 and group_id = $2)", sessionId, groupId)
		if err != nil {
			log.Errorf("failed to query codes: %s", err.Error())
			return c.Status(http.StatusInternalServerError).JSON(ApiError{Error: "database error"})
		}

		codeGroup.Codes = make([]CodeSummary, 0)
		for rows.Next() {
			var code CodeSummary
			err = rows.Scan(&code.CodeId, &code.Name, &code.PreferredName)
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
	row := db.QueryRow("select code_id, name, preferred_name from code where code_group_id = (select id from code_group where owner_id = $1 and group_id = $2) and code_id = $3", ownerId, groupId, codeId)
	if row == nil {
		return nil, fmt.Errorf("code not found")
	}

	var code CodeSummary
	err := row.Scan(&code.CodeId, &code.Name, &code.PreferredName)
	return &code, err
}
