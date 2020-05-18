package routes

import (
	"fmt"

	"github.com/gofiber/fiber"
	"github.com/jonathanlb/sqlite32grafana/cli"
	"github.com/jonathanlb/sqlite32grafana/sqlite3"
)

func InstallTagKeys(app *fiber.App, route cli.RouteConfig, tsm sqlite3.TimeSeriesManager) {
	endPoint := fmt.Sprintf("%s/%s/%s/tag-keys", route.DBAlias, route.Table, route.TimeColumn)
	app.Post(endPoint, func(c *fiber.Ctx) {
		sugar.Info(c.OriginalURL())
		var tagKeys []sqlite3.TagKey
		tsm.GetTagKeys(route.Table, &tagKeys) // check the error? how could we fix it?
		send200(c, tagKeys)
	})
}
