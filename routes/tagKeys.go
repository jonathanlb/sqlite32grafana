package routes

import (
	"github.com/gofiber/fiber"
)

func InstallTagKeys(app *fiber.App) {
	app.Post("/tag-keys", func(c *fiber.Ctx) {
		sugar.Info(c.OriginalURL())
	})
}
