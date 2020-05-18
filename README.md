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

## Query

When you build a query, select your datasource; the "timeserie" option, and a
column in table you'll be querying.
The drop-down hint will auto-populate with the table column names, from which
you can select one.

At the moment, you can enter any text you wish, which will blindly be used in
the SELECT statement, but this feature/bug will be disabled shortly, replaced
with a group-by feature....

## The Time Column

A time column can be either a scalar value, `DATETIME`, or `TEXT` column.
For scalars, sqlite32grafana will infer either epoch seconds, milliseconds, or
nanoseconds based upon the smallest value used in the column.

## Debugging

sqlite32grafana uses the `DEBUG` environment variable to turn on development
debugging.  Any non-empty value will trigger it at the moment....

## TODO
- Install raw-query option.
- Implement group-by option to regular queries.
- Implement tag-values.
- Summarize intervals.
