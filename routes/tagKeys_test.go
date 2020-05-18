package routes

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/gofiber/fiber"
	"github.com/jonathanlb/sqlite32grafana/cli"
	"github.com/jonathanlb/sqlite32grafana/sqlite3"
)

func Test_EmptyTagKeys(t *testing.T) {
	app := fiber.New(&fiber.Settings{})
	dbFileName := tempFileName(t)
	defer func() {
		os.Remove(dbFileName)
	}()

	tsm := createTimeSeriesManager(dbFileName)
	route := cli.RouteConfig{DBAlias: "db", Table: "tab", TimeColumn: "t"}
	InstallTagKeys(app, route, tsm)

	resp, err := postResponse(app, "/db/tab/t/tag-keys", "{}")
	check200(t, "tag-keys-empty", resp, err)
	body, _ := ioutil.ReadAll(resp.Body)
	var tagKeysResults []sqlite3.TagKey
	if err := json.Unmarshal(body, &tagKeysResults); err != nil {
		t.Fatalf("failed to read tag keys results response: %v", err)
	}
	expected := []sqlite3.TagKey{
		sqlite3.TagKey{Type: "number", Text: "x"},
		sqlite3.TagKey{Type: "string", Text: "tag"},
	}
	if !reflect.DeepEqual(expected, tagKeysResults) {
		t.Fatalf(`expected tag keys result "%+v", but got "%+v"`, expected, tagKeysResults)
	}
}
