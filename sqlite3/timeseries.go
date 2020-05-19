package sqlite3

import (
	"database/sql"
	"fmt"
	"reflect"
	"regexp"

	"strings"
	"time"

	"github.com/jonathanlb/sqlite32grafana/cli"
	"github.com/jonathanlb/sqlite32grafana/timecodex"
	_ "github.com/mattn/go-sqlite3" // db driver
	"github.com/pkg/errors"
)

type sqliteTimeSeriesManager struct {
	db         *sql.DB
	table      string
	timeColumn string
}

var sugar = cli.Logger()

var integerSQLTypes = NewSet("int", "integer", "tinyint")

func (seriesMan *sqliteTimeSeriesManager) GetTimeSeries(target string, fromTo *QueryRange, opts *TimeSeriesQueryOpts, dest *map[string][]DataPoint) error {
	fromTime, err := seriesMan.formatUserTimeForQuery(seriesMan.table, seriesMan.timeColumn, fromTo.From)
	if err != nil {
		return errors.Wrap(err, "get from time for timeseries")
	}
	toTime, err := seriesMan.formatUserTimeForQuery(seriesMan.table, seriesMan.timeColumn, fromTo.To)
	if err != nil {
		return errors.Wrap(err, "get to time for timeseries")
	}

	timeReader := seriesMan.getTimeToMillis(seriesMan.table, seriesMan.timeColumn) // XXX memoize?

	query, valueColumn, tagColumns := seriesMan.buildQuery(target, opts)
	sugar.Debugw("timeseries query",
		"query", query,
		"from", fromTime,
		"to", toTime)
	if valueColumn == "" {
		return errors.Errorf(`malformed target "%s"`, target)
	}

	rows, err := seriesMan.db.Query(query, fromTime, toTime)
	if err != nil {
		return errors.Wrap(err, "bad query for timeseries")
	}
	rowCount := 0
	result := make(map[string][]DataPoint)
	var values []interface{}
	tag := valueColumn // default value
	for rows.Next() {
		rowCount++
		if values == nil {
			values, err = getScanDest(rows)
			if err != nil {
				return err
			}
		}

		if err := rows.Scan(values...); err != nil {
			return errors.Errorf("Cannot scan row: %v", err)
		}

		timeMillis, err := timeReader(values[0])
		if err != nil {
			return err
		}

		value, err := valueReader(values[1])
		if err != nil {
			return err
		}

		if len(tagColumns) > 0 {
			var tagBuilder strings.Builder
			tagBuilder.WriteString(*values[2].(*string)) // XXX cast can break server
			for _, i := range values[3:] {
				tagBuilder.WriteString(" ")
				tagBuilder.WriteString(*i.(*string))
			}
			tag = tagBuilder.String()
		}
		newPoint := DataPoint{Time: timeMillis, Value: value}
		result[tag] = append(result[tag], newPoint)
	}
	*dest = result
	sugar.Debugw("timeseries completed", "#rows", rowCount)
	return nil
}

func (seriesMan *sqliteTimeSeriesManager) GetTagValues(tableName string, key string, dest *[]string) error {
	return nil
}

// New builds a new timeseries manager backed by the DB file and table with indexed time column.
func New(dbFileName string, table string, timeColumn string) (TimeSeriesManager, error) {
	db, err := sql.Open("sqlite3", dbFileName)
	if err != nil {
		return nil, err
	}
	//check presence of table and timeColumn
	tsm := sqliteTimeSeriesManager{db: db, table: table, timeColumn: timeColumn}
	var schema []TagKey
	if err := tsm.getSchema(table, &schema); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("cannot get schema for table %s", table))
	}
	if len(schema) > 0 {
		for _, col := range schema {
			if strings.EqualFold(col.Text, timeColumn) {
				// TODO: check valid column type?
				return &tsm, nil
			}
		}
	}
	return nil, errors.Wrap(err, fmt.Sprintf("cannot find time column %s in table with schema %+v", timeColumn, schema))
}

func (seriesMan *sqliteTimeSeriesManager) buildQuery(target string, opts *TimeSeriesQueryOpts) (string, string, []string) {
	valueColumn, tagColumns, selectExpr, groupExpr := seriesMan.parseTarget(target)
	var queryBuilder strings.Builder

	var orderBy string
	if groupExpr == "" {
		orderBy = seriesMan.timeColumn
	} else {
		orderBy = strings.Replace(groupExpr, " GROUP BY ", "", 1)
	}

	queryBuilder.WriteString(fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s >= ? AND %s < ?%s ORDER BY %s",
		selectExpr, seriesMan.table, seriesMan.timeColumn, seriesMan.timeColumn, groupExpr, orderBy))

	if opts != nil && opts.MaxDataPoints > 0 {
		queryBuilder.WriteString(fmt.Sprintf(" LIMIT %d", opts.MaxDataPoints))
	}

	return queryBuilder.String(), valueColumn, tagColumns
}

func dateTimeToMillis(input interface{}) (int64, error) {
	dateStr, ok := input.(*string)
	if !ok {
		return 0,
			errors.Errorf("cannot cast time column of type %v to *string", reflect.TypeOf(input))
	}

	ts, err := timecodex.StringToTime(*dateStr)
	if err != nil {
		return 0, errors.Errorf("Cannot parse time: %v", err)
	}
	millis := int64(ts.UnixNano() / 100000)
	return millis, nil
}

// Convert user-supplied time string to one comparable to the stated type
// of the column.
func (seriesMan *sqliteTimeSeriesManager) formatUserTimeForQuery(tableName string, timeColumn string, timeStr string) (interface{}, error) {
	columnType := seriesMan.getColumnType(tableName, timeColumn)
	switch strings.ToLower(columnType) {
	case "int":
		t, err := timecodex.StringToTime(timeStr)
		if err != nil {
			return nil, err
		}
		unitGuess := seriesMan.guessTimeMetric(tableName, timeColumn, t)
		return unitGuess, nil
	case "datetime":
		return timeStr, nil
	case "text":
		return timeStr, nil
	default:
		return nil, errors.Errorf("unknown time type %s for time column %s in table %s", columnType, timeColumn, tableName)
	}
}

func (seriesMan *sqliteTimeSeriesManager) getColumnType(tableName string, columnName string) string {
	var schema []TagKey
	if err := seriesMan.getSchema(tableName, &schema); err != nil {
		sugar.Panicf("cannot find column type for %s in table %s: %v", columnName, tableName, err)
	}
	for _, tag := range schema {
		if strings.EqualFold(columnName, tag.Text) {
			return tag.Type
		}
	}
	return ""
}

// Get scan value destinations ala https://github.com/golang/go/blob/master/src/database/sql/sql_test.go
// ColumnTypes not available until after Rows.Next() called https://github.com/mattn/go-sqlite3/issues/682
func getScanDest(rows *sql.Rows) ([]interface{}, error) {
	tt, err := rows.ColumnTypes()
	if err != nil {
		return nil, errors.Errorf("cannot infer column types from row: %v", err)
	}
	types := make([]reflect.Type, len(tt))
	for i, tp := range tt {
		st := tp.ScanType()
		if st == nil {
			return nil, errors.Errorf("cannot infer column type for %q", tp.Name())
		}
		types[i] = st
	}
	values := make([]interface{}, len(tt))
	for i := range values {
		values[i] = reflect.New(types[i]).Interface()
	}
	return values, nil
}

// Build a slice of column names and reported types.  Note that Sqlite3 does
// not validate that stored column values correspond to the stated type.
func (seriesMan *sqliteTimeSeriesManager) getSchema(tableName string, dest *[]TagKey) error {
	query := fmt.Sprintf("PRAGMA table_info(%s)", tableName)
	sugar.Debugw("schema", "query", query)
	schema, err := seriesMan.db.Query(query)
	if err != nil {
		return err
	}
	defer schema.Close()
	for schema.Next() {
		var cid, name, ctype, notNull, defaultVal, pk string
		schema.Scan(&cid, &name, &ctype, &notNull, &defaultVal, &pk)
		*dest = append(*dest, TagKey{ctype, name})
	}

	if len(*dest) == 0 {
		return errors.Errorf(`nonExistentTable: "%s"`, tableName)
	}

	return nil
}

// Determine which function to use to read a time value from the table/column
// and translate to epoch millis for Grafana.
func (seriesMan *sqliteTimeSeriesManager) getTimeToMillis(tableName string, timeColumn string) func(input interface{}) (int64, error) {
	columnType := seriesMan.getColumnType(tableName, timeColumn)
	switch strings.ToLower(columnType) {
	case "int":
		return seriesMan.columnToMillis(tableName, timeColumn)
	case "datetime":
		return dateTimeToMillis
	case "text":
		return dateTimeToMillis
	default:
		sugar.Panicf("unknown time type %s for time column %s in table %s", columnType, timeColumn, tableName)
		return nil
	}
}

// Take a time value and guess a numeric value useable for the specified
// timeColumn.
func (seriesMan *sqliteTimeSeriesManager) guessTimeMetric(tableName string, timeColumn string, t time.Time) int64 {
	scale, s := seriesMan.guessTimeScalar(tableName, timeColumn)
	if s {
		return t.Unix() * 1000 / scale
	}
	return t.Unix() * 1000 * scale
}

// Return a value to scale (multiply) numeric values from the table/column to
// arrive at epoch millis.
func (seriesMan *sqliteTimeSeriesManager) guessTimeScalar(tableName string, timeColumn string) (int64, bool) {
	query := fmt.Sprintf("SELECT %s FROM %s ORDER BY %s LIMIT 1", timeColumn, tableName, timeColumn)
	row := seriesMan.db.QueryRow(query)
	var t int64
	row.Scan(&t)
	return timecodex.NumberToScalar(t)
}

// Build a function to translate numeric time values from the table/column
// to epoch millis.
func (seriesMan *sqliteTimeSeriesManager) columnToMillis(tableName string, timeColumn string) func(input interface{}) (int64, error) {
	scale, s := seriesMan.guessTimeScalar(tableName, timeColumn)

	errFunc := func(input interface{}) (int64, error) {
		return 0, errors.Errorf("cannot cast %s %+v to millis", reflect.TypeOf(input), input)
	}

	if s {
		return func(input interface{}) (int64, error) {
			millis, ok := input.(*int64)
			if ok {
				return *millis * scale, nil
			}
			return errFunc(nil)
		}
	}
	return errFunc
}

// Break up the target into the value column, tag columns, selected columns,
// and group-by expression if necessary
func (seriesMan *sqliteTimeSeriesManager) parseTarget(target string) (string, []string, string, string) {
	valueColumn, tagColumns := seriesMan.target2tokens(target)
	var colBuilder strings.Builder
	var groupBy string

	// Scan tagOptions for a token of the form "t(...)"
	// signalling time-intervalization.  If one is present, return it stripped
	// of the t() and substituting "?" for the time column name; the the remaining
	// time options.  Otherwise, return the empty string and the original options.
	getTimeExpr := func(tagOptions []string) (string, []string) {
		for i, tag := range tagOptions {
			if strings.HasPrefix(tag, "t(") && strings.HasSuffix(tag, ")") {
				remainTags := append(tagOptions[:i], tagOptions[i+1:]...)
				timeF := strings.ReplaceAll(tag[2:len(tag)-1], "?", seriesMan.timeColumn)
				return timeF, remainTags
			}
		}
		return "", tagOptions
	}

	timeExp, tagColumns := getTimeExpr(tagColumns)
	if timeExp == "" {
		colBuilder.WriteString(fmt.Sprintf("%s, %s", seriesMan.timeColumn, valueColumn))
	} else {
		colBuilder.WriteString(fmt.Sprintf("%s, %s", timeExp, valueColumn))
		groupBy = fmt.Sprintf(" GROUP BY %s", timeExp)
	}

	for _, i := range tagColumns {
		colBuilder.WriteString(", ")
		colBuilder.WriteString(i)
	}

	return valueColumn, tagColumns, colBuilder.String(), groupBy
}

// Parse the target as "valueColumn [tagOptions]*"
func (seriesMan *sqliteTimeSeriesManager) target2tokens(target string) (string, []string) {
	tokens := strings.Fields(target)
	var valueColumn string
	if len(tokens) > 0 {
		valueColumn = tokens[0]
	}
	var tagOptions []string
	if len(tokens) > 1 {
		tagOptions = tokens[1:]
	}
	return valueColumn, tagOptions
}

var yyyymmdd = regexp.MustCompile(`^[0-9]{2,4}[/\- ][0-9]{1,2}[/\- ][0-9]{1,2}$`)
var intStr = regexp.MustCompile(`^[0-9]+$`)

func sql2grafanaType(sqlType string) string {
	switch strings.ToLower(sqlType) {
	case "int":
		return "number"
	case "text":
		return "string"
	default:
		sugar.Debugw("Unknown sql column",
			"type", sqlType)
		return "???"
	}
}

func valueReader(value interface{}) (float64, error) {
	x, ok := value.(*float64)
	if !ok {
		i, ok := value.(*int64)
		if !ok {
			return 0.0, errors.Errorf(`cannot coerce value "%v" to float64`, value)
		}
		f := float64(*i)
		return f, nil
	}
	return *x, nil
}
