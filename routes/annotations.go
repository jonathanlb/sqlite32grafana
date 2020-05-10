package routes

import (
	"github.com/gofiber/fiber"
)

func InstallAnnotations(app *fiber.App) {
	app.Post("/annotations", func(c *fiber.Ctx) {
		sugar.Info(c.OriginalURL())
	})
}
