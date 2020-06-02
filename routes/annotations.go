package routes

import (
	"fmt"

	"github.com/gofiber/fiber"
	"github.com/jonathanlb/sqlite32grafana/cli"
	"github.com/jonathanlb/sqlite32grafana/sqlite3"
)

// InstallAnnotations is a stub to set up the ReST end point for Grafana
// access metadata around SQLite tables.  It's currently inactive.
func InstallAnnotations(app *fiber.App, route cli.RouteConfig, tsm sqlite3.TimeSeriesManager) {
	endPoint := fmt.Sprintf("%s/%s/%s/annotations", route.DBAlias, route.Table, route.TimeColumn)
	app.Post(endPoint, func(c *fiber.Ctx) {
		sugar.Info(c.OriginalURL())
	})
}
