package routes

import (
	"testing"

	"github.com/gofiber/fiber"
	"github.com/jonathanlb/sqlite32grafana/cli"
)

func Test_TestConnection(t *testing.T) {
	app := fiber.New(&fiber.Settings{})
	route := cli.RouteConfig{DBAlias: "db", Table: "tab", TimeColumn: "t"}
	InstallTestConnection(app, route)
	resp, err := getResponse(app, "/db/tab/t")
	defer resp.Body.Close()

	check200(t, "test-connection", resp, err)
	checkBody(t, "test-connection", "ok", resp)
}
