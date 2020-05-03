Leverage [simple-json-datasource](https://github.com/grafana/simple-json-datasource) to allow Grafana to pull data from Sqlite3 database files.

## Startup
```
go run main <sqlite.db files>
```

## Query
The target name parameter should be formed as follows
```
table-file time-column value-column [key-column]*
```
