package main

import (
	"embed"
	"fmt"
	"github.com/EphyraSoftware/locus/auth"
	"github.com/EphyraSoftware/locus/coldmfa"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/cors"
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
	oryPublicUrl := os.Getenv("ORY_PUBLIC_URL")
	if oryPublicUrl == "" {
		oryPublicUrl = "http://127.0.0.1:4433"
	}
	config.Servers = ory.ServerConfigurations{{URL: oryPublicUrl}}

	oryClient := ory.NewAPIClient(config)

	log.SetLevel(log.LevelInfo)
	log.Infof("Ory client connected @ %s\n", oryClient.GetConfig().Servers[0].URL)

	engine := html.NewFileSystem(http.FS(public), ".html")
	app := fiber.New(fiber.Config{
		Views:       engine,
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
	})

	if len(os.Args) >= 2 && os.Args[1] == "dev" {
		log.Info("Running in dev mode")
		app.Use(cors.New(cors.Config{
			AllowOrigins:     "http://localhost:5173",
			AllowHeaders:     "Origin, Content-Type, Accept",
			AllowMethods:     "GET, POST, PUT, DELETE",
			AllowCredentials: true,
		}))
		devAuth(app)
	} else {
		log.Info("Running in production mode")
		oryAuthApp := auth.App{
			Router:  app.Group("/auth"),
			Ory:     oryClient,
			OryBase: oryPublicUrl,
		}
		oryAuthApp.Prepare(app)
	}

	databaseUrlPath := os.Getenv("DATABASE_URL_FILE")
	var databaseUrl string
	if databaseUrlPath == "" {
		databaseUrl = "postgres://coldmfa:coldmfa@localhost:5432/coldmfa?sslmode=disable"
	} else {
		databaseUrlBytes, err := os.ReadFile(databaseUrlPath)
		if err != nil {
			log.Fatal(err)
		}
		databaseUrl = string(databaseUrlBytes)
	}
	coldMfaApp := coldmfa.App{
		Router:      app.Group("/coldmfa"),
		DatabaseUrl: databaseUrl,
		Public:      public,
	}
	coldMfaApp.Prepare()

	app.Get("/", func(c *fiber.Ctx) error {
		log.Debug("Redirecting to coldmfa")
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

func devAuth(app *fiber.App) {
	app.Use(func(c *fiber.Ctx) error {
		var out map[string]interface{}
		if err := json.Unmarshal([]byte("{\"email\": \"tester@local.net\", \"name\": {\"username\": \"tester\"}}"), &out); err != nil {
			return err
		}
		c.Locals("session", &ory.Session{
			Identity: &ory.Identity{
				Id:     "tester",
				Traits: out,
			},
		})

		return c.Next()
	})
}
