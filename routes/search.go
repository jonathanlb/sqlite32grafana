package routes

import (
	"log"

	"github.com/gofiber/fiber"
)

func InstallSearch(app *fiber.App) {
	app.Get("/search", func(c *fiber.Ctx) {
		log.Print(c.OriginalURL())
	})
}
