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

	var app = fiber.New()
	app.Use(logger.New())
	for _, route := range config.Routes {
		tsm, err := sqlite3.New(route.DBFile, route.Table, route.TimeColumn)
		if err != nil {
			log.Fatalf("cannot open db for route %+v: %+v", route, err)
		}
		routes.InstallAllRoutes(app, route, tsm)
	}

	if err := app.Listen(config.Port); err != nil {
		log.Fatalf("cannot listen on port %d: %+v", config.Port, err)
	}
}
