package routes

import (
	"log"

	"github.com/gofiber/fiber"
)

func InstallAnnotations(app *fiber.App) {
	app.Post("/annotations", func(c *fiber.Ctx) {
		log.Print(c.OriginalURL())
	})
}
