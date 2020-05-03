package routes

import (
	"log"

	"github.com/gofiber/fiber"
)

func InstallTagValues(app *fiber.App) {
	app.Post("/tag-values", func(c *fiber.Ctx) {
		log.Print(c.OriginalURL())
	})
}
