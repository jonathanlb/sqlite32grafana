package sqlite3

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func Test_GetTagKeysFailsOnInvalidTableName(t *testing.T) {
	db, _ := sql.Open("sqlite3", ":memory")
	tsm := newFromDb(db, []string{})
	var keys []TagKey
	if err := tsm.GetTagKeys("nonExistentTable", &keys); err == nil {
		t.Fatalf("Did not fail to get keys from non-existant table: %v", keys)
	}
}

func Test_GetTagKeys(t *testing.T) {
	db, _ := sql.Open("sqlite3", ":memory")
	db.Exec("CREATE TABLE tsTab (x INT, tag TEXT, t INT)")
	tsm := newFromDb(db, []string{"tsTab"})
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
	db, _ := sql.Open("sqlite3", ":memory")
	db.Exec("CREATE TABLE tsTab (x INT, tag TEXT, t INT)")
	tsm := newFromDb(db, []string{"tsTab"})
	var keys []TagKey
	if err := tsm.GetTagKeys("tsTab", &keys); err != nil {
		t.Fatalf("Failed to query keys: %v", err)
	}
	if keys == nil || len(keys) != 3 {
		t.Fatalf("Failed to infer all 3 keys, got %v", keys)
	}
}

func Test_DetectUnknownColumnType(t *testing.T) {
	if colType := sql2grafanaType("thingy"); "???" != colType {
		t.Fatalf(`sql2grafana must return "???" for unknown, got "%s"`, colType)
	}
}
