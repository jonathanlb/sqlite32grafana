package cli

import (
	"reflect"
	"strings"
	"testing"
)

func Test_ParseArgs(t *testing.T) {
	expectedRoute := RouteConfig{DBAlias: "db", DBFile: "db.sqlite3", Table: "a", TimeColumn: "ts"}
	args := strings.Split("-port 4000 -tab a -time ts -db db.sqlite3 -a db", " ")
	config, err := Parse(args)

	if err != nil {
		t.Fatalf(`unexpected error "%v"`, err)
	}
	if config.Port != 4000 {
		t.Fatalf(`expected port 4000 but got %d`, config.Port)
	}

	if len(config.Routes) != 1 || !reflect.DeepEqual(expectedRoute, config.Routes[0]) {
		t.Fatalf(`expected table list "a b", but got "%v"`, config.Routes)
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

	if err == nil || !strings.HasPrefix(err.Error(), "each -db option requires a -tab ") {
		t.Fatalf(`expected table option required but got error "%v"`, err)
	}
}

func Test_RequiresTime(t *testing.T) {
	args := strings.Split("-port 4000 -db db.sqlite3 -tab table", " ")
	_, err := Parse(args)

	if err == nil || !strings.HasPrefix(err.Error(), "each -tab option requires a -time ") {
		t.Fatalf(`expected time column option required but got error "%v"`, err)
	}
}

func Test_RequiresMatchingAlias(t *testing.T) {
	args := strings.Split("-port 4000 -db db.sqlite3 -tab table -time t -db db.sqlite3 -tab table -time t -a db", " ")
	_, err := Parse(args)

	if err == nil || !strings.HasPrefix(err.Error(), "either all db files must have alias, or none") {
		t.Fatalf(`either all db files must have alias, or none, but got error "%v"`, err)
	}
}
