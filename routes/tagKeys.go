package routes

import (
	"log"

	"github.com/gofiber/fiber"
)

func InstallTagKeys(app *fiber.App) {
	app.Post("/tag-keys", func(c *fiber.Ctx) {
		log.Print(c.OriginalURL())
	})
}
