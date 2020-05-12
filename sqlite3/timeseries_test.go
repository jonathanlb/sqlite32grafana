package sqlite3

import (
	"database/sql"
	"io/ioutil"
	"os"
	"reflect"
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
		t.Fatalf(`Cannot create file-backed db at %s: "%+v"`, dbFileName, err)
	}
	db.Exec("CREATE TABLE tsTab (x INT, tag TEXT, t INT)")
	db.Close()

	_, err = New(dbFileName, []string{})
	if err != nil {
		t.Fatalf(`Cannot create test db "%+v"`, err)
	}
}

func Test_formatUserTimeForQuery(t *testing.T) {
	db := createDbWithTable(t)
	tsm := newFromDb(db, []string{"tsTab"}).(*sqliteTimeSeriesManager)
	ts, err := tsm.formatUserTimeForQuery("tsTab", "ts", "2020-05-01")
	if err != nil {
		t.Fatalf(`cannot parse time "2020-05-01" for int time: %+v`, err)
	}
	tt := reflect.TypeOf(ts).String()
	if tt != "int64" {
		t.Fatalf(`expected "2020-05-01" for int time to be of type int64 not %s`, tt)
	}
	if ts.(int64) < 1588204800 || ts.(int64) > 1588377600 {
		t.Fatalf(`cannot parse time "2020-05-01" for int time: %+v, expected greater than 1588204800`, ts)
	}

	ts, err = tsm.formatUserTimeForQuery("tsTab", "ts", "1")
	if err != nil {
		t.Fatalf(`cannot parse time "1" for text time: %+v`, err)
	}
	tt = reflect.TypeOf(ts).String()
	if tt != "int64" {
		t.Fatalf(`expected "1" for integer-like string time to be of type int64 not %s`, tt)
	}
	if ts.(int64) > 86400 {
		t.Fatalf(`cannot parse time "1" for string time: %+v`, ts)
	}

	ts, err = tsm.formatUserTimeForQuery("tsTab", "tag", "2020-05-01")
	if err != nil {
		t.Fatalf(`cannot parse time "2020-05-01" for text time: %+v`, err)
	}
	tt = reflect.TypeOf(ts).String()
	if tt != "string" {
		t.Fatalf(`expected "2020-05-01" for string time to be of type string not %s`, tt)
	}
	if ts.(string) != "2020-05-01" {
		t.Fatalf(`cannot parse time "2020-05-01" for string time: %+v`, ts)
	}

	ts, err = tsm.formatUserTimeForQuery("tsTab", "dt", "2020-05-01")
	tt = reflect.TypeOf(ts).String()
	if tt != "string" {
		t.Fatalf(`expected "2020-05-01" for string time to be of type datetime not %s`, tt)
	}
	if ts.(string) != "2020-05-01" {
		t.Fatalf(`cannot parse time "2020-05-01" for string time: %+v`, ts)
	}
}

func Test_GetTimeSeries(t *testing.T) {
	db := createDbWithTable(t)
	tsm := newFromDb(db, []string{"tsTab"})
	var ts map[string][]DataPoint
	err := tsm.GetTimeSeries("tsTab ts x tag", "0", "10", &ts)
	if err != nil {
		t.Fatalf(`Unexpected error querying timeseries "%+v"`, err)
	}
	if ts == nil || len(ts) != 2 ||
		len(ts["a"]) != 2 || len(ts["b"]) != 2 ||
		ts["a"][0] != (DataPoint{Time: 1000, Value: 100.}) ||
		ts["a"][1] != (DataPoint{Time: 3000, Value: 300.}) ||
		ts["b"][0] != (DataPoint{Time: 2000, Value: 200.}) ||
		ts["b"][1] != (DataPoint{Time: 4000, Value: 400.}) {
		t.Fatalf(`Unexpected timeseries response "%+v"`, ts)
	}
}

func Test_GetTimeSeriesParsingRange(t *testing.T) {
	db := createDbWithTable(t)
	tsm := newFromDb(db, []string{"tsTab"})
	var ts map[string][]DataPoint
	err := tsm.GetTimeSeries("tsTab ts x tag", "1969-01-01", "1971-12-31", &ts)
	if err != nil {
		t.Fatalf(`Unexpected error querying timeseries with datetime "%+v"`, err)
	}
	if ts == nil || len(ts) != 2 ||
		len(ts["a"]) != 2 || len(ts["b"]) != 2 ||
		ts["a"][0] != (DataPoint{Time: 1000, Value: 100.}) ||
		ts["a"][1] != (DataPoint{Time: 3000, Value: 300.}) ||
		ts["b"][0] != (DataPoint{Time: 2000, Value: 200.}) ||
		ts["b"][1] != (DataPoint{Time: 4000, Value: 400.}) {
		t.Fatalf(`Unexpected timeseries response from datetime range "%+v"`, ts)
	}
}

func Test_GetTimeSeriesWithTimeRange(t *testing.T) {
	db := createDbWithTable(t)
	tsm := newFromDb(db, []string{"tsTab"})
	var ts map[string][]DataPoint
	err := tsm.GetTimeSeries("tsTab ts x", "2", "4", &ts)
	if err != nil {
		t.Fatalf(`Unexpected error querying timeseries "%+v"`, err)
	}
	if ts == nil || len(ts) != 1 ||
		ts["x"][0] != (DataPoint{Time: 2000, Value: 200.}) ||
		ts["x"][1] != (DataPoint{Time: 3000, Value: 300.}) {
		t.Fatalf(`Unexpected timeseries response "%+v"`, ts)
	}
}

func Test_GetTimeSeriesWithDatetimeIndex(t *testing.T) {
	db := createDbWithTable(t)
	tsm := newFromDb(db, []string{"tsTab"})
	var ts map[string][]DataPoint
	err := tsm.GetTimeSeries("tsTab dt x", "2020-04-02", "2020-04-04", &ts)
	if err != nil {
		t.Fatalf(`Unexpected error querying timeseries "%+v"`, err)
	}
	if ts == nil || len(ts) != 1 ||
		ts["x"][0].Value != 200 ||
		ts["x"][1].Value != 300 {
		t.Fatalf(`Unexpected timeseries response "%+v"`, ts)
	}
}

func Test_GetTimeSeriesFailsOnMissingTable(t *testing.T) {
	db, _ := sql.Open("sqlite3", ":memory:")
	tsm := newFromDb(db, []string{})
	var ts map[string][]DataPoint
	err := tsm.GetTimeSeries("tsTab x", "0", "10", &ts)
	if err == nil || !strings.HasPrefix(err.Error(), "malformed target") {
		t.Fatalf(`Unexpected error querying missing table "%+v"`, err)
	}
}

func Test_GetTimeSeriesFailsOnUnspecifiedTimeColumn(t *testing.T) {
	db := createDbWithTable(t)
	tsm := newFromDb(db, []string{"tsTab"})
	var ts map[string][]DataPoint
	err := tsm.GetTimeSeries("tsTab", "0", "10", &ts)
	if err == nil || !strings.HasPrefix(err.Error(), "malformed target") {
		t.Fatalf(`Unexpected error querying missing time column "%+v"`, err)
	}
}

func Test_GetTimeSeriesFailsOnUnspecifiedValueColumn(t *testing.T) {
	db := createDbWithTable(t)
	tsm := newFromDb(db, []string{"tsTab"})
	var ts map[string][]DataPoint
	err := tsm.GetTimeSeries("tsTabt ", "0", "10", &ts)
	if err == nil || !strings.HasPrefix(err.Error(), "malformed target") {
		t.Fatalf(`Unexpected error querying missing table "%+v"`, err)
	}
}

func Test_guessTimeScalar(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal("Cannot create in-memory sqlite DB")
	}

	queries := []string{
		"CREATE TABLE tsTab (seconds INT, millis int, nanos INT, dt DATETIME)",
		"INSERT INTO tsTab (seconds, millis, nanos, dt) VALUES (1585742400, 1585742400000, 1585742400000000000, '2020-04-01 12:00')",
		"INSERT INTO tsTab (seconds, millis, nanos, dt) VALUES (1585828800, 1585828800000, 1585828800000000000, '2020-04-02 12:00')",
	}
	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			t.Fatalf(`cannot issue query "%s" for test: %+v`, q, err)
		}
	}
	tsm := newFromDb(db, []string{"tsTab"}).(*sqliteTimeSeriesManager)
	scalar, sign := tsm.guessTimeScalar("tsTab", "seconds")
	if scalar != 1000 || !sign {
		t.Fatalf(`expected scalar, sign of 1000, true, but got %d, %t`, scalar, sign)
	}

	scalar, sign = tsm.guessTimeScalar("tsTab", "millis")
	if scalar != 1 {
		t.Fatalf(`expected scalar, sign of 1, true, but got %d, %t`, scalar, sign)
	}

	scalar, sign = tsm.guessTimeScalar("tsTab", "nanos")
	if scalar != 1000000 || sign {
		t.Fatalf(`expected scalar, sign of 1000000, false, but got %d, %t`, scalar, sign)
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
		"CREATE TABLE tsTab (x INT, tag TEXT, ts INT, dt DATETIME)",
		"CREATE INDEX idx_tsTab_ts ON tsTab(ts)",
		"INSERT INTO tsTab (ts, x, tag, dt) VALUES (1, 100, 'a', '2020-04-01')",
		"INSERT INTO tsTab (ts, x, tag, dt) VALUES (2, 200, 'b', '2020-04-02')",
		"INSERT INTO tsTab (ts, x, tag, dt) VALUES (3, 300, 'a', '2020-04-03')",
		"INSERT INTO tsTab (ts, x, tag, dt) VALUES (4, 400, 'b', '2020-04-04')",
	}
	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			t.Fatalf(`cannot issue query "%s" for test: %+v`, q, err)
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
