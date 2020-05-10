package routes

import (
	"github.com/gofiber/fiber"
)

func InstallSearch(app *fiber.App) {
	app.Get("/search", func(c *fiber.Ctx) {
		sugar.Info(c.OriginalURL())
	})
}
