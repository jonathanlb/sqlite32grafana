package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber"
	"github.com/gofiber/logger"
	"github.com/jonathanlb/sqlite32grafana/cli"
	"github.com/jonathanlb/sqlite32grafana/routes"
	"github.com/jonathanlb/sqlite32grafana/sqlite3"
)

func main() {
	config, err := cli.Parse(os.Args[1:])
	if err != nil {
		log.Fatal(err.Error())
	}
	tsm, err := sqlite3.New(config.DBFile, config.Tables)
	if err != nil {
		log.Fatalf("cannot open db: %v", err)
	}

	var app = fiber.New()
	app.Use(logger.New())
	routes.InstallAllRoutes(app, tsm)
	app.Listen(config.Port)
}
