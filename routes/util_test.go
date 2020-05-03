package routes

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gofiber/fiber"
)

func check200(t *testing.T, testName string, resp *http.Response, err error) {
	checkStatus(t, testName, 200, resp, err)
}

func checkBody(t *testing.T, testName string, expected string, resp *http.Response) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf(`unexpected error reading %s body: "%v"`, testName, err)
	}
	strResp := string(body)
	if strResp != expected {
		t.Fatalf(`unexpected %s response: "%s", expected "%s"`, testName, strResp, expected)
	}
}

func checkStatus(t *testing.T, testName string, expectStatus int, resp *http.Response, err error) {
	if err != nil {
		t.Fatalf(`unexpected %s error: '%v'`, testName, err)
	}
	if resp.StatusCode != expectStatus {
		t.Fatalf(`unexpected %s code: %d, expected %d`, testName, resp.StatusCode, expectStatus)
	}
}

func getResponse(app *fiber.App, url string, m ...string) (*http.Response, error) {
	req, _ := http.NewRequest("GET", url, nil)
	return app.Test(req)
}

func postResponse(app *fiber.App, url string, payload string) (*http.Response, error) {
	bs := []byte(payload)
	req := httptest.NewRequest("POST", url, bytes.NewBuffer(bs))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Length", strconv.Itoa(len(bs)))
	return app.Test(req)
}
