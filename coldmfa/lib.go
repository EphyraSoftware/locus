package coldmfa

import (
	"database/sql"
	"embed"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"log"
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

	_, err = sql.Open("postgres", a.DatabaseUrl)
	if err != nil {
		log.Fatal(err)
	}

	a.Router.Use(filesystem.New(filesystem.Config{
		Root:       http.FS(public),
		PathPrefix: "public",
		Index:      "index.html",
	}))

	api := a.Router.Group("/api")

	api.Get("/groups", func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})
}
