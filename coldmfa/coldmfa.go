package coldmfa

import "C"
import (
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
	"net/http"
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

		rows, err := db.Query("SELECT id, name FROM code_group where owner_id = $1", sessionId)
		if err != nil {
			log.Errorf("failed to query groups: %s", err.Error())
			return c.Status(http.StatusInternalServerError).JSON(ApiError{Error: "database error"})
		}

		out := make([]CodeGroup, 0)
		for rows.Next() {
			var group CodeGroup
			err = rows.Scan(&group.Id, &group.Name)
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

		id := c.Params("id")
		if id == "" {
			return c.Status(http.StatusBadRequest).JSON(ApiError{Error: "missing id"})
		}

		codeGroup, err := readCodeGroup(db, sessionId, id)
		if err != nil {
			log.Errorf("failed to read group: %s", err.Error())
			return c.Status(http.StatusNotFound).JSON(ApiError{Error: "group not found"})
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

		createdCodeGroup, err := readCodeGroup(db, sessionId, groupId)
		if err != nil {
			log.Errorf("failed to read group: %s", err.Error())
			return c.Status(http.StatusInternalServerError).JSON(ApiError{Error: "group not found"})
		}

		return c.Status(http.StatusCreated).JSON(createdCodeGroup)
	})
}

func readCodeGroup(db *sql.DB, ownerId string, groupId string) (*CodeGroup, error) {
	row := db.QueryRow("select group_id, name from code_group where owner_id = $1 and group_id = $2", ownerId, groupId)
	if row == nil {
		return nil, fmt.Errorf("group not found")
	}

	var group CodeGroup
	err := row.Scan(&group.Id, &group.Name)
	return &group, err
}
