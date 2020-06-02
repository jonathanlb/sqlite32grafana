package cli

import (
	"errors"
	"flag"
	"os"

	"go.uber.org/zap"
)

// RouteConfig stores SQLite table information to expose to ReST for
// for simple-json-datasource access.
type RouteConfig struct {
	DBAlias    string
	DBFile     string
	Table      string
	TimeColumn string
}

// Config stores application startup options.
type Config struct {
	Routes []RouteConfig
	Port   int
}

type arrayFlags []string

func (i *arrayFlags) String() string {
	return "multiple string arguments"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var logger *zap.Logger

// Logger creates a zap sugared logger, using environment variables
// for configuration.
func Logger() *zap.SugaredLogger {
	if logger == nil {
		debugEnv := os.Getenv("DEBUG")
		if len(debugEnv) > 0 {
			logger, _ = zap.NewDevelopment()
		} else {
			logger, _ = zap.NewProduction()
		}
	}
	return logger.Sugar()
}

// Parse command-line arguments to configure the Sqlite to Grafana interface.
func Parse(args []string) (Config, error) {
	var config Config
	fs := flag.NewFlagSet("sqlite2grafana", flag.ContinueOnError)
	var files, filesAlia, tables, columns arrayFlags
	fs.Var(&files, "db", "Sqlite3 backing file")
	fs.Var(&filesAlia, "a", "File endpoint alias")
	fs.Var(&tables, "tab", "Table to serve")
	fs.Var(&columns, "time", "Time column")
	fs.IntVar(&config.Port, "port", 4200, "Port serving requests")
	fs.Parse(args)

	if len(files) <= 0 {
		return config, errors.New("-db <file-name> option required")
	}
	if len(filesAlia) != 0 && len(filesAlia) != len(files) {
		return config, errors.New("either all db files must have alias, or none")
	}
	if len(tables) != len(files) {
		return config, errors.New("each -db option requires a -tab <table-name> option")
	}
	if len(columns) != len(tables) {
		return config, errors.New("each -tab option requires a -time <time-column> option")
	}

	for i, f := range files {
		route := RouteConfig{DBFile: f, Table: tables[i], TimeColumn: columns[i]}
		if len(filesAlia) == 0 {
			route.DBAlias = route.DBFile
		} else {
			route.DBAlias = filesAlia[i]
		}
		config.Routes = append(config.Routes, route)
	}

	return config, nil
}
