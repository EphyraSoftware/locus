package main

import (
	"embed"
	"fmt"
	"github.com/EphyraSoftware/locus/auth"
	"github.com/EphyraSoftware/locus/coldmfa"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/template/html/v2"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	ory "github.com/ory/client-go"
	"net/http"
	"os"
)

//go:embed public/*
var public embed.FS

func main() {
	config := ory.NewConfiguration()
	oryPort := 4433
	oryBase := fmt.Sprintf("http://127.0.0.1:%d/", oryPort)
	config.Servers = ory.ServerConfigurations{{URL: oryBase}}

	oryClient := ory.NewAPIClient(config)

	log.SetLevel(log.LevelInfo)
	log.Infof("Ory client connected @ %s\n", oryClient.GetConfig().Servers[0].URL)

	engine := html.NewFileSystem(http.FS(public), ".html")
	app := fiber.New(fiber.Config{
		Views:       engine,
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
	})

	oryAuthApp := auth.App{
		Router:  app.Group("/auth"),
		Ory:     oryClient,
		OryBase: oryBase,
	}
	oryAuthApp.Prepare(app)

	coldMfaApp := coldmfa.App{
		Router:      app.Group("/coldmfa"),
		DatabaseUrl: "postgres://coldmfa:coldmfa@localhost:5432/coldmfa?sslmode=disable",
	}
	coldMfaApp.Prepare()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Redirect("/coldmfa")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	err := app.Listen(fmt.Sprintf("[::]:%s", port))
	if err != nil {
		log.Fatal(err)
	}
}
