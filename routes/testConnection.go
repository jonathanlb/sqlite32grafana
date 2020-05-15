package routes

import (
	"fmt"

	"github.com/jonathanlb/sqlite32grafana/cli"

	"github.com/gofiber/fiber"
)

func InstallTestConnection(app *fiber.App, route cli.RouteConfig) {
	endPoint := fmt.Sprintf("%s/%s/%s/", route.DBAlias, route.Table, route.TimeColumn)
	app.Get(endPoint, func(c *fiber.Ctx) {
		c.Send("ok")
	})
}
