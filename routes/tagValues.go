package routes

import (
	"github.com/gofiber/fiber"
)

func InstallTagValues(app *fiber.App) {
	app.Post("/tag-values", func(c *fiber.Ctx) {
		sugar.Info(c.OriginalURL())
	})
}
