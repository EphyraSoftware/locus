package main

import (
	"fmt"
	"github.com/coldmfa/v2/auth"
	"github.com/coldmfa/v2/coldmfa"
	"github.com/gofiber/fiber/v2"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	ory "github.com/ory/client-go"
	"log"
)

func main() {
	config := ory.NewConfiguration()
	oryPort := 4433
	oryBase := fmt.Sprintf("http://127.0.0.1:%d/", oryPort)
	config.Servers = ory.ServerConfigurations{{URL: oryBase}}

	oryClient := ory.NewAPIClient(config)

	log.Printf("Ory client connected %s\n", oryClient.GetConfig().Servers)

	app := fiber.New()

	coldMfaRouter := app.Group("/coldmfa", func(c *fiber.Ctx) error {
		log.Println("ColdMFA middleware")

		sessionCookie := c.Cookies("ory_kratos_session")
		log.Println("Session cookie", sessionCookie)

		return c.Next()
	})

	coldMfaApp := coldmfa.App{
		Router:      coldMfaRouter,
		DatabaseUrl: "postgres://coldmfa:coldmfa@localhost:5432/coldmfa?sslmode=disable",
	}
	coldMfaApp.Prepare()

	go func() {
		err := auth.StartApp(oryClient, oryBase)
		if err != nil {
			log.Fatal(err)
		}
	}()

	err := app.Listen("[::]:3000")
	if err != nil {
		log.Fatal(err)
	}
}
