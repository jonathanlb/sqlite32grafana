package routes

import (
	"encoding/json"

	"github.com/gofiber/fiber"
	"github.com/jonathanlb/sqlite32grafana/cli"
	"github.com/jonathanlb/sqlite32grafana/sqlite3"
)

var sugar = cli.Logger()

func InstallAllRoutes(app *fiber.App, route cli.RouteConfig, tsm sqlite3.TimeSeriesManager) {
	InstallTestConnection(app, route)
	InstallSearch(app, route, tsm)
	InstallQuery(app, route, tsm)
	InstallAnnotations(app, route, tsm)
	InstallTagKeys(app, route, tsm)
	InstallTagValues(app, route, tsm)
}

func send200(c *fiber.Ctx, result interface{}) {
	resultBytes, err := json.Marshal(result)
	if err != nil {
		send400(c, err)
		return
	}
	c.Set("Content-Type", "application/json")
	c.Send(resultBytes)
	c.SendStatus(200)
}

func send400(c *fiber.Ctx, err error) {
	c.SendStatus(400)
	c.SendString(err.Error())
}
