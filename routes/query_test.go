package routes

import (
	"testing"

	"github.com/gofiber/fiber"
)

func Test_FailEmptyTimeseries(t *testing.T) {
	app := fiber.New(&fiber.Settings{})
	InstallQuery(app)
	resp, err := postResponse(app, "/query", "")

	checkStatus(t, "get-empty-timeseries", 400, resp, err)
}

func Test_FailBadJsonTimeseries(t *testing.T) {
	app := fiber.New(&fiber.Settings{})
	InstallQuery(app)
	queryStr := `{"range": {`
	resp, err := postResponse(app, "/query", queryStr)

	checkStatus(t, "get-empty-timeseries", 400, resp, err)
}

func Test_GetTimeseries(t *testing.T) {
	app := fiber.New(&fiber.Settings{})
	InstallQuery(app)

	queryStr := `{
    "range": {
      "from": "2020-03-16", "to": "2020-05-01"
    },
    "interval": "1h",
    "intervalMs": 3600000,
    "targets": [],
    "adhocFilters": [],
    "format": "json",
    "maxDataPoints": 1023
  }`
	resp, err := postResponse(app, "/query", queryStr)

	check200(t, "get-timeseries", resp, err)
	// checkBody(t, "test-connection", "ok", resp)
}
