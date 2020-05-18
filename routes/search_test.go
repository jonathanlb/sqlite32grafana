package routes

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/gofiber/fiber"
	"github.com/jonathanlb/sqlite32grafana/cli"
)

func Test_EmptySearch(t *testing.T) {
	app := fiber.New(&fiber.Settings{})
	dbFileName := tempFileName(t)
	defer func() {
		os.Remove(dbFileName)
	}()

	tsm := createTimeSeriesManager(dbFileName)
	route := cli.RouteConfig{DBAlias: "db", Table: "tab", TimeColumn: "t"}
	InstallSearch(app, route, tsm)

	resp, err := postResponse(app, "/db/tab/t/search", "")
	check200(t, "search-empty", resp, err)
	body, _ := ioutil.ReadAll(resp.Body)
	var searchResults []string
	if err := json.Unmarshal(body, &searchResults); err != nil {
		t.Fatalf("failed to read search results response: %v", err)
	}
	expected := []string{"x", "tag"}
	if !reflect.DeepEqual(expected, searchResults) {
		t.Fatalf(`expected search result "%+v", but got "%+v"`, expected, searchResults)
	}

	resp, err = postResponse(app, "/db/tab/t/search", "not json")
	checkStatus(t, "search-garbage", 400, resp, err)
}

func Test_SubStringSearch(t *testing.T) {
	app := fiber.New(&fiber.Settings{})
	dbFileName := tempFileName(t)
	defer func() {
		os.Remove(dbFileName)
	}()

	tsm := createTimeSeriesManager(dbFileName)
	route := cli.RouteConfig{DBAlias: "db", Table: "tab", TimeColumn: "t"}
	InstallSearch(app, route, tsm)

	resp, err := postResponse(app, "/db/tab/t/search", `{"target":"ag"}`)
	check200(t, "search-substring", resp, err)
	body, _ := ioutil.ReadAll(resp.Body)
	var searchResults []string
	if err := json.Unmarshal(body, &searchResults); err != nil {
		t.Fatalf("failed to read search results response: %v", err)
	}
	expected := []string{"tag"}
	if !reflect.DeepEqual(expected, searchResults) {
		t.Fatalf(`expected substring search result "%+v", but got "%+v"`, expected, searchResults)
	}
}
