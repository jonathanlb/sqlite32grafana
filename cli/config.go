package cli

import (
	"errors"
	"flag"
)

type Config struct {
	DBFile string
	Tables []string
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

func Parse(args []string) (Config, error) {
	var config Config
	fs := flag.NewFlagSet("sqlite2grafana", flag.ContinueOnError)
	var tables arrayFlags
	fs.StringVar(&config.DBFile, "db", "", "Sqlite3 backing file")
	fs.Var(&tables, "tab", "Tables to serve")
	fs.IntVar(&config.Port, "port", 4200, "Port serving requests")
	fs.Parse(args)
	config.Tables = tables

	if config.DBFile == "" {
		return config, errors.New("-db <file-name> option required")
	}
	if len(config.Tables) <= 0 {
		return config, errors.New("at least one -tab <file-name> option required")
	}

	return config, nil
}
