package coldmfa

import "C"
import (
	"database/sql"
	"embed"
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

//go:embed public/*
var public embed.FS

type App struct {
	Router      fiber.Router
	DatabaseUrl string
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
		Root:       http.FS(public),
		PathPrefix: "public",
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
		rows, err := db.Query("SELECT id, name FROM code_group")
		if err != nil {
			log.Errorf("failed to query groups: %s", err.Error())
			return c.SendStatus(http.StatusInternalServerError)
		}

		out := make([]CodeGroup, 0)
		for rows.Next() {
			var group CodeGroup
			err = rows.Scan(&group.Id, &group.Name)
			if err != nil {
				log.Errorf("failed to scan group: %s", err.Error())
				return c.SendStatus(http.StatusInternalServerError)
			}
			out = append(out, group)
		}

		return c.Status(200).JSON(out)
	})

	api.Get("/groups/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		if id == "" {
			return c.SendStatus(http.StatusBadRequest)
		}

		var group CodeGroup
		err := db.QueryRow("SELECT id, name FROM code_group WHERE id = $1", id).Scan(&group.Id, &group.Name)
		if err != nil {
			log.Errorf("failed to query group: %s", err.Error())
			return c.SendStatus(http.StatusInternalServerError)
		}

		return c.Status(200).JSON(group)
	})

	api.Post("/groups", func(c *fiber.Ctx) error {
		codeGroup := new(CodeGroup)
		if err := c.BodyParser(codeGroup); err != nil {
			return c.SendStatus(http.StatusBadRequest)
		}

		groupId, err := gonanoid.New()
		if err != nil {
			log.Errorf("failed to generate group id: %s", err.Error())
			return c.SendStatus(http.StatusInternalServerError)
		}

		_, err = db.Exec("INSERT INTO code_group (group_id, name) VALUES ($1, $2)", groupId, codeGroup.Name)

		return nil
	})
}
