package main

import (
	"github.com/gofiber/fiber"
	"github.com/gofiber/logger"
	"github.com/jonathanlb/sqlite32grafana/routes"
)

func main() {
	var app = fiber.New()
  app.Use(logger.New())
	routes.InstallAllRoutes(app)
	app.Listen(3939)
}
