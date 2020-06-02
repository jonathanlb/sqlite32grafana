package routes

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gofiber/fiber"
	"github.com/jonathanlb/sqlite32grafana/cli"
	"github.com/jonathanlb/sqlite32grafana/sqlite3"
)

type searchTarget struct {
	Target string `json:"target"`
}

// InstallSearch sets up the search end point for simple-json-datasource to
// wire up column hints. We will use tag-keys + time column for now.
func InstallSearch(app *fiber.App, route cli.RouteConfig, tsm sqlite3.TimeSeriesManager) {
	endPoint := fmt.Sprintf("%s/%s/%s/search", route.DBAlias, route.Table, route.TimeColumn)
	app.Post(endPoint, func(c *fiber.Ctx) {
		body := []byte(c.Body())
		var target string

		if len(c.Body()) > 0 {
			var targetJSON searchTarget
			err := json.Unmarshal(body, &targetJSON)
			if err != nil {
				send400(c, err)
				return
			}
			target = strings.ToLower(targetJSON.Target)
		} else {
			target = ""
		}

		var tagKeys []sqlite3.TagKey
		tsm.GetTagKeys(route.Table, &tagKeys)
		result := []string{}

		addTagKey := func(tag string) {
			lowerTag := strings.ToLower(tag)
			if strings.Contains(lowerTag, target) {
				result = append(result, lowerTag)
			}
		}
		for _, i := range tagKeys {
			addTagKey(strings.ToLower(i.Text))
		}
		addTagKey(route.TimeColumn)
		send200(c, result)
	})
}
