package sqlite3

import (
	"database/sql"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func Test_CreateInMemory(t *testing.T) {
	db, _ := sql.Open("sqlite3", ":memory:")
	newFromDb(db, []string{})
}

func Test_CreateFromFileNames(t *testing.T) {
	dbFileName := tempFileName(t)
	configFileName := tempFileName(t)
	defer func() {
		for _, f := range []string{dbFileName, configFileName} {
			os.Remove(f)
		}
	}()

	db, err := sql.Open("sqlite3", dbFileName)
	if err != nil {
		t.Fatalf(`Cannot create file-backed db at %s: "%v"`, dbFileName, err)
	}
	db.Exec("CREATE TABLE tsTab (x INT, tag TEXT, t INT)")
	db.Close()

	_, err = New(dbFileName, []string{})
	if err != nil {
		t.Fatalf(`Cannot create test db "%v"`, err)
	}
}

func Test_GetTimeSeries(t *testing.T) {
	db := createDbWithTable(t)
	tsm := newFromDb(db, []string{"tsTab"})
	var ts map[string][]DataPoint
	err := tsm.GetTimeSeries("tsTab t x tag", "0", "10", &ts)
	if err != nil {
		t.Fatalf(`Unexpected error querying timeseries "%v"`, err)
	}
	if ts == nil || len(ts) != 2 ||
		len(ts["a"]) != 2 || len(ts["b"]) != 2 ||
		ts["a"][0] != (DataPoint{Time: 1, Value: 100.}) ||
		ts["a"][1] != (DataPoint{Time: 3, Value: 300.}) ||
		ts["b"][0] != (DataPoint{Time: 2, Value: 200.}) ||
		ts["b"][1] != (DataPoint{Time: 4, Value: 400.}) {
		t.Fatalf(`Unexpected timeseries response "%v"`, ts)
	}
}

func Test_GetTimeSeriesWithTimeRange(t *testing.T) {
	db := createDbWithTable(t)
	tsm := newFromDb(db, []string{"tsTab"})
	var ts map[string][]DataPoint
	err := tsm.GetTimeSeries("tsTab t x", "2", "4", &ts)
	if err != nil {
		t.Fatalf(`Unexpected error querying timeseries "%v"`, err)
	}
	if ts == nil || len(ts) != 1 ||
		ts["x"][0] != (DataPoint{Time: 2, Value: 200.}) ||
		ts["x"][1] != (DataPoint{Time: 3, Value: 300.}) {
		t.Fatalf(`Unexpected timeseries response "%v"`, ts)
	}
}

func Test_GetTimeSeriesWithDatetimeIndex(t *testing.T) {
	db, _ := sql.Open("sqlite3", ":memory:")
	db.Exec("CREATE TABLE tsTab (x INT, tag TEXT, t DATETIME)")
	db.Exec("CREATE INDEX idx_tsTab_t ON tsTab(t)")
	db.Exec("INSERT INTO tsTab (t, x) VALUES ('2020-04-01', 100)")
	db.Exec("INSERT INTO tsTab (t, x) VALUES ('2020-04-02', 200)")
	db.Exec("INSERT INTO tsTab (t, x) VALUES ('2020-04-03', 300)")
	db.Exec("INSERT INTO tsTab (t, x) VALUES ('2020-04-04', 400)")

	tsm := newFromDb(db, []string{"tsTab"})
	var ts map[string][]DataPoint
	err := tsm.GetTimeSeries("tsTab t x", "2020-04-02", "2020-04-04", &ts)
	if err != nil {
		t.Fatalf(`Unexpected error querying timeseries "%v"`, err)
	}
	if ts == nil || len(ts) != 1 ||
		ts["x"][0].Value != 200 ||
		ts["x"][1].Value != 300 {
		t.Fatalf(`Unexpected timeseries response "%v"`, ts)
	}
}

func Test_GetTimeSeriesFailsOnMissingTable(t *testing.T) {
	db, _ := sql.Open("sqlite3", ":memory:")
	tsm := newFromDb(db, []string{})
	var ts map[string][]DataPoint
	err := tsm.GetTimeSeries("tsTab x", "0", "10", &ts)
	if err == nil || !strings.HasPrefix(err.Error(), "malformed target") {
		t.Fatalf(`Unexpected error querying missing table "%v"`, err)
	}
}

func Test_GetTimeSeriesFailsOnUnspecifiedTimeColumn(t *testing.T) {
	db := createDbWithTable(t)
	tsm := newFromDb(db, []string{"tsTab"})
	var ts map[string][]DataPoint
	err := tsm.GetTimeSeries("tsTab", "0", "10", &ts)
	if err == nil || !strings.HasPrefix(err.Error(), "malformed target") {
		t.Fatalf(`Unexpected error querying missing time column "%v"`, err)
	}
}

func Test_GetTimeSeriesFailsOnUnspecifiedValueColumn(t *testing.T) {
	db := createDbWithTable(t)
	tsm := newFromDb(db, []string{"tsTab"})
	var ts map[string][]DataPoint
	err := tsm.GetTimeSeries("tsTabt ", "0", "10", &ts)
	if err == nil || !strings.HasPrefix(err.Error(), "malformed target") {
		t.Fatalf(`Unexpected error querying missing table "%v"`, err)
	}
}

func Test_target2tokens(t *testing.T) {
	db := createDbWithTable(t)
	tsm := newFromDb(db, []string{"tsTab"}).(*sqliteTimeSeriesManager)
	table, time, value, tags := tsm.target2tokens("tsTab t x")
	if table != "tsTab" || time != "t" || value != "x" || len(tags) != 0 {
		t.Fatalf(
			`Expected target2tokens to be "tsTab, t, x, []", but got "%s, %s, %s, %v"`,
			table, time, value, tags)
	}
}

func Test_target2tokensFailsOnMissingTable(t *testing.T) {
	db := createDbWithTable(t)
	tsm := newFromDb(db, []string{"tsTab"}).(*sqliteTimeSeriesManager)
	table, time, value, tags := tsm.target2tokens("foo")
	if table != "" || time != "" || value != "" || len(tags) != 0 {
		t.Fatalf(
			`Expected failure of target2tokens, but got "%s, %s, %s, %v"`,
			table, time, value, tags)
	}
}

func Test_target2tokensFailsOnMissingTableWithTarget(t *testing.T) {
	db := createDbWithTable(t)
	tsm := newFromDb(db, []string{"tsTab"}).(*sqliteTimeSeriesManager)
	table, time, value, tags := tsm.target2tokens("foo t x")
	if table != "" || time != "" || value != "" || len(tags) != 0 {
		t.Fatalf(
			`Expected failure of target2tokens, but got "%s, %s, %s, %v"`,
			table, time, value, tags)
	}
}

func Test_target2tokensOnOverspecified(t *testing.T) {
	db := createDbWithTable(t)
	tsm := newFromDb(db, []string{"tsTab"}).(*sqliteTimeSeriesManager)
	table, time, value, tags := tsm.target2tokens("tsTab t x tag ...")
	if table != "tsTab" || time != "t" || value != "x" || len(tags) != 2 {
		t.Fatalf(
			`Expected target2tokens "tsTab, t, x, tag, ...", but got "%s, %s, %s, %v"`,
			table, time, value, tags)
	}
}

func createDbWithTable(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal("Cannot create in-memory sqlite DB")
	}

	queries := []string{
		"CREATE TABLE tsTab (x INT, tag TEXT, t INT)",
		"CREATE INDEX idx_tsTab_t ON tsTab(t)",
		"INSERT INTO tsTab (t, x, tag) VALUES (1, 100, 'a')",
		"INSERT INTO tsTab (t, x, tag) VALUES (2, 200, 'b')",
		"INSERT INTO tsTab (t, x, tag) VALUES (3, 300, 'a')",
		"INSERT INTO tsTab (t, x, tag) VALUES (4, 400, 'b')",
	}
	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			t.Fatalf(`cannot issue query "%s" for test: %v`, q, err)
		}
	}
	return db
}

func tempFileName(t *testing.T) string {
	f, err := ioutil.TempFile("", "sqlite32grafana-timeseries-test-")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}
