package routes

import (
	"github.com/gofiber/fiber"
	"github.com/jonathanlb/sqlite32grafana/sqlite3"
	"go.uber.org/zap"
)

var logger, _ = zap.NewDevelopment()

// var logger, _ = zap.NewProduction()
var sugar = logger.Sugar()

func InstallAllRoutes(app *fiber.App, tsm sqlite3.TimeSeriesManager) {
	InstallTestConnection(app)
	InstallSearch(app)
	InstallQuery(app, tsm)
	InstallAnnotations(app)
	InstallTagKeys(app)
	InstallTagValues(app)
}
