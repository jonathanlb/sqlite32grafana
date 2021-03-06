package routes

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/gofiber/fiber"
	"github.com/jonathanlb/sqlite32grafana/cli"
)

func Test_FailEmptyTimeseries(t *testing.T) {
	app := fiber.New(&fiber.Settings{})
	dbFileName := tempFileName(t)
	defer func() {
		os.Remove(dbFileName)
	}()
	tsm := createTimeSeriesManager(dbFileName)
	route := cli.RouteConfig{DBAlias: "db", Table: "tab", TimeColumn: "t"}
	InstallQuery(app, route, tsm)
	resp, err := postResponse(app, "/db/tab/t/query", "")

	checkStatus(t, "query-empty-timeseries", 400, resp, err)
}

func Test_FailBadJsonTimeseries(t *testing.T) {
	app := fiber.New(&fiber.Settings{})
	dbFileName := tempFileName(t)
	defer func() {
		os.Remove(dbFileName)
	}()
	tsm := createTimeSeriesManager(dbFileName)
	route := cli.RouteConfig{DBAlias: "db", Table: "tab", TimeColumn: "t"}
	InstallQuery(app, route, tsm)
	queryStr := `{"range": {`
	resp, err := postResponse(app, "/db/tab/t/query", queryStr)

	checkStatus(t, "query-badjson-timeseries", 400, resp, err)
}

func Test_GetTimeseries(t *testing.T) {
	app := fiber.New(&fiber.Settings{})
	dbFileName := tempFileName(t)
	defer func() {
		os.Remove(dbFileName)
	}()

	tsm := createTimeSeriesManager(dbFileName)
	route := cli.RouteConfig{DBAlias: "db", Table: "tab", TimeColumn: "t"}
	InstallQuery(app, route, tsm)

	queryStr := `{
    "range": {
      "from": "2020-03-16", "to": "2020-05-01"
    },
    "interval": "1h",
    "intervalMs": 3600000,
    "targets": [{ "target": "x tag", "refId": "A", "type": "timeserie" }],
    "adhocFilters": [],
    "format": "json",
    "maxDataPoints": 1023
  }`
	resp, err := postResponse(app, "/db/tab/t/query", queryStr)

	check200(t, "query-timeseries", resp, err)
	body, _ := ioutil.ReadAll(resp.Body)
	var timeseries []Timeseries
	if err := json.Unmarshal(body, &timeseries); err != nil {
		t.Fatalf("failed to read timeseries response: %v", err)
	}
	sugar.Debugf("response %+v", timeseries)
	if len(timeseries) != 2 {
		t.Fatalf("read %d series, expected 2", len(timeseries))
	}
}
