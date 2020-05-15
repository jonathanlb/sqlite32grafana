package routes

import (
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
