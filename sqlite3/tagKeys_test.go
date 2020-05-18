package sqlite3

import (
	"database/sql"
	"reflect"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func Test_GetTagKeysFailsOnInvalidTableName(t *testing.T) {
	db, _ := sql.Open("sqlite3", ":memory:")
	tsm := sqliteTimeSeriesManager{db: db, table: "tsTab", timeColumn: "t"}
	var keys []TagKey
	if err := tsm.GetTagKeys("nonExistentTable", &keys); err == nil {
		t.Fatalf("Did not fail to get keys from non-existant table: %v", keys)
	}
}

func Test_GetTagKeys(t *testing.T) {
	db, _ := sql.Open("sqlite3", ":memory:")
	db.Exec("CREATE TABLE tsTab (x INT, tag TEXT, t INT)")
	tsm := sqliteTimeSeriesManager{db: db, table: "tsTab", timeColumn: "t"}
	var keys []TagKey
	if err := tsm.GetTagKeys("tsTab t x", &keys); err != nil {
		t.Fatalf("Failed to query keys: %v", err)
	}
	expectedKey := TagKey{Type: "string", Text: "tag"}
	if keys == nil || len(keys) != 1 || keys[0] != expectedKey {
		t.Fatalf(`Failed to infer keys expected "[%v]", got "%v"`, expectedKey, keys)
	}
}

func Test_GetTagKeysRaw(t *testing.T) {
	db, _ := sql.Open("sqlite3", ":memory:")
	db.Exec("CREATE TABLE tsTab (x INT, tag TEXT, t INT)")
	tsm := sqliteTimeSeriesManager{db: db, table: "tsTab", timeColumn: "t"}
	var keys []TagKey
	if err := tsm.GetTagKeys("tsTab", &keys); err != nil {
		t.Fatalf("Failed to query keys: %v", err)
	}
	expectedKeys := []TagKey{
		TagKey{"number", "x"},
		TagKey{"string", "tag"},
	}
	if !reflect.DeepEqual(expectedKeys, keys) {
		t.Fatalf(`expected keys "%+v", but got "%+v"`, expectedKeys, keys)
	}
}

func Test_DetectUnknownColumnType(t *testing.T) {
	if colType := sql2grafanaType("thingy"); "???" != colType {
		t.Fatalf(`sql2grafana must return "???" for unknown, got "%s"`, colType)
	}
}
