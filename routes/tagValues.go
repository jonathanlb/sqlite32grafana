package routes

import (
	"fmt"

	"github.com/gofiber/fiber"
	"github.com/jonathanlb/sqlite32grafana/cli"
	"github.com/jonathanlb/sqlite32grafana/sqlite3"
)

// InstallTagValues is an inactive placeholder to publish values in columns
// labeling timeseries observations.
func InstallTagValues(app *fiber.App, route cli.RouteConfig, tsm sqlite3.TimeSeriesManager) {
	endPoint := fmt.Sprintf("%s/%s/%s/tag-values", route.DBAlias, route.Table, route.TimeColumn)
	app.Post(endPoint, func(c *fiber.Ctx) {
		sugar.Info(c.OriginalURL())
	})
}
