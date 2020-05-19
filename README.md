# sqlite32grafana

Monitor and explore the contents of your
[Sqlite3](https://www.sqlite.org/index.html) database using
[Grafana](https://github.com/grafana/grafana) and
[simple-json-datasource](https://github.com/grafana/simple-json-datasource)
 to allow Grafana to pull data from Sqlite3 database files.

## Startup
```
go run main -port <port-number> \
  [-db file-name.sqlite3 -tab table-name -time time-column [-a db-alias] ]*
```

sqlite32grafana will fire up a server to listen for timeseries requests.
(Table queries are not yet implemented.)

On your Grafana server,

- Install [simple-json-datasource](https://github.com/grafana/simple-json-datasource).
- Create a new datasource on from your Grafana web interface.
 - Select "SimpleJson" under the "Others" section near the bottom of the list.
 - Enter an URL under HTTP that looks like
```
http://your-host:port/db-file-or-alias/table-name/time-column
```
 using the the command-line arguments matching the tuples used upon sqlite32grafana start up above.  The `-a` option, alias, is useful/required for hiding slashes and other unpleasant characters in the DB file.
 - Clicking "Save & Test" should send a liveness check to your sqlite32grafana instance, or alert you of a mistake.
- You'll need a separate datasource for every time column you'll query.

### The Time Column

A time column can be either a scalar value, `DATETIME`, or `TEXT` column.
For scalars, sqlite32grafana will infer either epoch seconds, milliseconds, or
nanoseconds based upon the smallest value used in the column.

## Query

When you build a query, select your datasource; the "timeserie" option, and a
column name in table you'll be querying for time series values.
The drop-down hint will auto-populate with the table column names, from which
you can select one.

You can apply options to the query by appending them to the column name.
Use a space to delimit an option from the value column and adjacent options.

The entry will be used to build a SQL query in the formatUserTimeForQuery:
```
SELECT time-exp, value-exp FROM table WHERE time-range [interval-group]
```
(ignoring order and row-limit modifiers).

### Summarization
Any place that you use a column name in a query, you can also use
[an aggregate function](https://www.sqlite.org/lang_aggfunc.html) on
the column, such as `count()` or `sum()`, useful during intervalization.

### Intervalize
You can intervalize your results using the `t()` option to transform the time
column and group by the results.
The argument to `t()` is a SQL clause to transform the time column, which you
can alias as `?`

For example, if you have a time column named `created` representing epoch
seconds, you can count the number of rows every six hours with the target query
```
count(created) t(21600*(?/21600))
```
which translates to a querying
```
SELECT 21600*(created/21600), count(created) FROM notes WHERE created >= ? AND created < ? GROUP BY 21600*(created/21600) ORDER BY 21600*(created/21600)
```
We'll implement duration-based rounding intervalization to allow terser queries.

## Debugging

sqlite32grafana uses the `DEBUG` environment variable to turn on development
debugging.  Any non-empty value will trigger it at the moment....

## TODO
- Implement multiple group-by options.
- Implement tag-values.
- Clean up time intervalization to allow duration, e.g. `i(10s)`
- Fix bug
```
DEBUG   sqlite3/timeseries.go:55 timeseries
            query  {"query": "SELECT created, author, author FROM notes WHERE created >= ? AND created < ? ORDER BY created LIMIT 1250", "from": 1587308754, "to": 1589900754}
panic: interface conversion: interface {} is *int64, not *string
goroutine 20 [running]:
github.com/jonathanlb/sqlite32grafana/sqlite3.(*sqliteTimeSeriesManager).GetTimeSeries(0xc0002966f0, 0xc00010f840, 0xd, 0xc000134000, 0xc00011c2d0, 0xc000112058, 0x0, 0xc0003e9868)
        github.com/jonathanlb/sqlite32grafana/sqlite3/timeseries.go:84 +0x1388
github.com/jonathanlb/sqlite32grafana/routes.InstallQuery.func1(0xc000120000)
        github.com/jonathanlb/sqlite32grafana/routes/query.go:53 +0x6d3
```
