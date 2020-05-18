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

// Set up the search end point for simple-json-datasource.... which seems
// to wire up column hints.... tag-keys?
func InstallSearch(app *fiber.App, route cli.RouteConfig, tsm sqlite3.TimeSeriesManager) {
	endPoint := fmt.Sprintf("%s/%s/%s/search", route.DBAlias, route.Table, route.TimeColumn)
	app.Post(endPoint, func(c *fiber.Ctx) {
		body := []byte(c.Body())
		var target searchTarget

		if len(c.Body()) > 0 {
			err := json.Unmarshal(body, &target)
			if err != nil {
				send400(c, err)
				return
			}
			target.Target = strings.ToLower(target.Target)
		} else {
			target = searchTarget{""}
		}

		var tagKeys []sqlite3.TagKey
		tsm.GetTagKeys(route.Table, &tagKeys)
		result := []string{}
		for _, i := range tagKeys {
			resultText := strings.ToLower(i.Text)
			if strings.Contains(resultText, target.Target) {
				result = append(result, resultText)
			}
		}
		send200(c, result)
	})
}
