package routes

import (
	"github.com/gofiber/fiber"
)

func InstallAllRoutes(app *fiber.App) {
	InstallTestConnection(app)
	InstallSearch(app)
	InstallQuery(app)
	InstallAnnotations(app)
	InstallTagKeys(app)
	InstallTagValues(app)
}
