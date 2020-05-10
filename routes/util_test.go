package routes

import (
	"bytes"
	"database/sql"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gofiber/fiber"
	"github.com/jonathanlb/sqlite32grafana/sqlite3"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

func check200(t *testing.T, testName string, resp *http.Response, err error) {
	checkStatus(t, testName, 200, resp, err)
}

func checkBody(t *testing.T, testName string, expected string, resp *http.Response) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("%+v", errors.Errorf(`unexpected error reading %s body: "%v"`, testName, err))
	}
	strResp := string(body)
	if strResp != expected {
		t.Fatalf("%+v", errors.Errorf(`unexpected %s response: "%s", expected "%s"`, testName, strResp, expected))
	}
}

func checkStatus(t *testing.T, testName string, expectStatus int, resp *http.Response, err error) {
	if err != nil {
		t.Fatalf("%+v", errors.Errorf(`unexpected %s error: '%v'`, testName, err))
	}
	if resp.StatusCode != expectStatus {
		body, _ := ioutil.ReadAll(resp.Body)
		t.Fatalf("%+v", errors.Errorf(`unexpected %s code: %d, expected %d ("%s")`, testName, resp.StatusCode, expectStatus, body))
	}
}

func createTimeSeriesManager(dbFileName string) sqlite3.TimeSeriesManager {
	db, err := sql.Open("sqlite3", dbFileName)
	if err != nil {
		sugar.Fatalf("cannot open sqlite at %s: %v", dbFileName, err)
	}

	queries := []string{
		"CREATE TABLE series (x INT, tag TEXT, t DATETIME)",
		"CREATE INDEX idx_series_t ON series(t)",
		"INSERT INTO series (t, x, tag) VALUES ('2020-04-01', 100, 'a')",
		"INSERT INTO series (t, x, tag) VALUES ('2020-04-02', 200, 'b')",
		"INSERT INTO series (t, x, tag) VALUES ('2020-04-03', 300, 'a')",
		"INSERT INTO series (t, x, tag) VALUES ('2020-04-04', 400, 'b')",
	}
	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			sugar.Fatalf(`cannot execute sqlite query "%s": %v`, q, err)
		}
	}
	db.Close()

	tsm, err := sqlite3.New(dbFileName, []string{"series"})
	if err != nil {
		sugar.Fatalf(`cannot create time series manager: %v`, err)
	}
	return tsm
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

func tempFileName(t *testing.T) string {
	f, err := ioutil.TempFile("", "sqlite32grafana-routes-test-")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}
