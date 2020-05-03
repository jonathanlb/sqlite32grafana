package routes

import (
	"github.com/gofiber/fiber"
)

func InstallTestConnection(app *fiber.App) {
	app.Get("/", func(c *fiber.Ctx) {
		c.Send("ok")
	})
}
