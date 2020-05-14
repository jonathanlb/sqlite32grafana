package routes

import (
	"github.com/gofiber/fiber"
)

func InstallSearch(app *fiber.App) {
	app.Post("/search", func(c *fiber.Ctx) {
		body := []byte(c.Body())
		sugar.Debugw("route search", "body", string(body))
		sugar.Info(c.OriginalURL())
	})
}
