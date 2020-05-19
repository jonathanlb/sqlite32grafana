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
 using the the command-line arguments matching the tuples used upon sqlite32grafana start up above.  The `-a` option, alias, is useful/required for avoiding slashes and other unpleasant characters in the DB file from appearing in the REST endpoint.
 - Clicking "Save & Test" should send a liveness check to your sqlite32grafana instance, or alert you of a mistake.
- You'll need a separate datasource for every time column you'll query.

### The Time Column

A time column can be either a scalar value, `DATETIME`, or `TEXT` column.
For scalars, sqlite32grafana will infer either epoch seconds, milliseconds, or
nanoseconds based upon the smallest value used in the column.

## Query

When you build a query, select your datasource; the "timeserie" option, and
enter a column name from table you'll be querying for time series values.
The drop-down hint will auto-populate with the table column names, from which
you can select one.

You can apply tags or options to the query by appending them to the column name.
Use a space to delimit an option from the value column and adjacent options.
The expected format/order of elements in the query is
```
value-column [tag-column] [interval-option]
```

The entry will be used to build a SQL query in the form:
```
SELECT time-expr, value-expr, [tag-column] FROM table WHERE time-range [interval-group]
```
(ignoring order and row-limit modifiers).

### Tags
You can signal Grafana to plot multiple time series from the same table using
a another column to name (tag) the series.
For example, if you have a table
```
CREATE TABLE patientTemperature (ts INT, patient TEXT, tempF REAL)
```
The query target `tempF patient` will plot one temperature series for every
unique value of `patient` encountered.

### Summarization
Any place that you use a column name in a query, you can also use
[an aggregate function](https://www.sqlite.org/lang_aggfunc.html) on
the column, such as `count()` or `sum()`, useful during intervalization, next.

### Intervalize
You can intervalize your results using the `t()` option to transform the time
column and group by the results.
The argument to `t()` is a SQL clause to transform the time column, which you
can alias as `?`

For example, if the time column `ts` represents epoch
seconds, you can plot the count the number of rows every six hours with
the target query
```
count(ts) t(21600*(?/21600))
```
which translates to a querying
```
SELECT 21600*(ts/21600), count(ts) FROM notes WHERE ts >= ? AND ts < ? GROUP BY 21600*(ts/21600) ORDER BY 21600*(ts/21600)
```
We'll implement duration-based rounding intervalization to allow terser queries
for a future issue.

Currently, we ignore the `__interval` and `__interval_ms` options on the
Grafana request.

## Debugging

sqlite32grafana uses the `DEBUG` environment variable to turn on development
debugging.  Any non-empty value will trigger it at the moment....

## TODO
- Fix the tag values displayed in the legend.
- Implement multiple group-by options.
- Implement tag-values.
- Add intervalization aliases to allow duration, e.g. `i(10s)` to intervalize
every 10 seconds.
