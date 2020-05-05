package cli

import (
	"reflect"
	"strings"
	"testing"
)

func Test_ParseArgs(t *testing.T) {
	args := strings.Split("-port 4000 -tab a -tab b -db db.sqlite3", " ")
	config, err := Parse(args)

	if err != nil {
		t.Fatalf(`unexpected error "%v"`, err)
	}
	expectedTables := []string{"a", "b"}
	if !reflect.DeepEqual(config.Tables, expectedTables) {
		t.Fatalf(`expected table list "a b", but got "%v"`, config.Tables)
	}

	if config.DBFile != "db.sqlite3" {
		t.Fatalf(`expected db file "db.sqlite3", but got "%s"`, config.DBFile)
	}
}

func Test_RequiresDB(t *testing.T) {
	args := strings.Split("-port 4000 -tab a -tab b", " ")
	_, err := Parse(args)

	if err == nil || !strings.HasPrefix(err.Error(), "-db ") {
		t.Fatalf(`expected db option required but got error "%v"`, err)
	}
}

func Test_RequiresTable(t *testing.T) {
	args := strings.Split("-port 4000 -db db.sqlite3", " ")
	_, err := Parse(args)

	if err == nil || !strings.HasPrefix(err.Error(), "at least one -tab ") {
		t.Fatalf(`expected table option required but got error "%v"`, err)
	}
}
