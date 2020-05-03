package routes

import (
	"testing"

	"github.com/gofiber/fiber"
)

func Test_TestConnection(t *testing.T) {
	app := fiber.New(&fiber.Settings{})
	InstallTestConnection(app)
	resp, err := getResponse(app, "/")
	defer resp.Body.Close()

	check200(t, "test-connection", resp, err)
	checkBody(t, "test-connection", "ok", resp)
}
