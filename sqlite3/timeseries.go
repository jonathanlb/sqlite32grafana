package sqlite3

import (
	"database/sql"
	"fmt"
	"reflect"
	"regexp"

	"strings"
	"time"

	"github.com/jonathanlb/sqlite32grafana/timecodex"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type sqliteTimeSeriesManager struct {
	db     *sql.DB
	tables *set
}

var logger, _ = zap.NewDevelopment()

// var logger, _ = zap.NewProduction()

// defer logger.Sync()
var sugar = logger.Sugar()

var integerSqlTypes = NewSet("int", "integer", "tinyint")

func (this *sqliteTimeSeriesManager) GetTimeSeries(target string, fromTo *QueryRange, opts *TimeSeriesQueryOpts, dest *map[string][]DataPoint) error {
	tableName, timeColumn, valueColumn, tagColumns := this.target2tokens(target)
	if tableName == "" || timeColumn == "" || valueColumn == "" {
		return errors.Errorf(`malformed target "%s"`, target)
	}

	var colBuilder strings.Builder
	colBuilder.WriteString(fmt.Sprintf("%s, %s", timeColumn, valueColumn))
	for _, i := range tagColumns {
		colBuilder.WriteString(", ")
		colBuilder.WriteString(i)
	}
	fromTime, err := this.formatUserTimeForQuery(tableName, timeColumn, fromTo.From)
	if err != nil {
		return errors.Wrap(err, "get from time for timeseries")
	}
	toTime, err := this.formatUserTimeForQuery(tableName, timeColumn, fromTo.To)
	if err != nil {
		return errors.Wrap(err, "get to time for timeseries")
	}

	timeReader := this.getTimeToMillis(tableName, timeColumn) // XXX memoize?

	var queryBuilder strings.Builder
	queryBuilder.WriteString(fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s >= ? AND %s < ? ORDER BY %s",
		colBuilder.String(), tableName, timeColumn, timeColumn, timeColumn))

	if opts != nil && opts.MaxDataPoints > 0 {
		queryBuilder.WriteString(fmt.Sprintf(" LIMIT %d", opts.MaxDataPoints))
	}

	query := queryBuilder.String()
	sugar.Debugw("timeseries query",
		"query", query,
		"from", fromTime,
		"to", toTime)

	rows, err := this.db.Query(query, fromTime, toTime)
	if err != nil {
		return errors.Wrap(err, "bad query for timeseries")
	}
	rowCount := 0
	result := make(map[string][]DataPoint)
	var values []interface{}
	tag := valueColumn
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
			tagBuilder.WriteString(*values[2].(*string))
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

func (this *sqliteTimeSeriesManager) GetTagValues(tableName string, key string, dest *[]string) error {
	return nil
}

func New(dbFileName string, tables []string) (TimeSeriesManager, error) {
	db, err := sql.Open("sqlite3", dbFileName)
	if err != nil {
		return nil, err
	}
	return newFromDb(db, tables), nil
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
func (this *sqliteTimeSeriesManager) formatUserTimeForQuery(tableName string, timeColumn string, timeStr string) (interface{}, error) {
	columnType := this.getColumnType(tableName, timeColumn)
	switch strings.ToLower(columnType) {
	case "int":
		t, err := timecodex.StringToTime(timeStr)
		if err != nil {
			return nil, err
		}
		unitGuess := this.guessTimeMetric(tableName, timeColumn, t)
		return unitGuess, nil
	case "datetime":
		return timeStr, nil
	case "text":
		return timeStr, nil
	default:
		return nil, errors.Errorf("unknown time type %s for time column %s in table %s", columnType, timeColumn, tableName)
	}
}

func (this *sqliteTimeSeriesManager) getColumnType(tableName string, columnName string) string {
	var schema []TagKey
	if err := this.getSchema(tableName, &schema); err != nil {
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
func (this *sqliteTimeSeriesManager) getSchema(tableName string, dest *[]TagKey) error {
	query := fmt.Sprintf("PRAGMA table_info(%s)", tableName)
	sugar.Debugw("schema", "query", query)
	schema, err := this.db.Query(query)
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
func (this *sqliteTimeSeriesManager) getTimeToMillis(tableName string, timeColumn string) func(input interface{}) (int64, error) {
	columnType := this.getColumnType(tableName, timeColumn)
	switch strings.ToLower(columnType) {
	case "int":
		return this.columnToMillis(tableName, timeColumn)
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
func (this *sqliteTimeSeriesManager) guessTimeMetric(tableName string, timeColumn string, t time.Time) int64 {
	scale, s := this.guessTimeScalar(tableName, timeColumn)
	if s {
		return t.Unix() * 1000 / scale
	} else {
		return t.Unix() * 1000 * scale
	}
}

// Return a value to scale (multiply) numeric values from the table/column to
// arrive at epoch millis.
func (this *sqliteTimeSeriesManager) guessTimeScalar(tableName string, timeColumn string) (int64, bool) {
	query := fmt.Sprintf("SELECT %s FROM %s ORDER BY %s LIMIT 1", timeColumn, tableName, timeColumn)
	row := this.db.QueryRow(query)
	var t int64
	row.Scan(&t)
	return timecodex.NumberToScalar(t)
}

// Build a function to translate numeric time values from the table/column
// to epoch millis.
func (this *sqliteTimeSeriesManager) columnToMillis(tableName string, timeColumn string) func(input interface{}) (int64, error) {
	scale, s := this.guessTimeScalar(tableName, timeColumn)

	if s {
		return func(input interface{}) (int64, error) {
			millis, ok := input.(*int64)
			if ok {
				return *millis * scale, nil
			}
			return 0, errors.Errorf("cannot cast %s %+v to millis", reflect.TypeOf(input), input)
		}
	} else {
		return func(input interface{}) (int64, error) {
			millis, ok := input.(*int64)
			if ok {
				return *millis / scale, nil
			}
			return 0, errors.Errorf("cannot cast %s %+v to millis", reflect.TypeOf(input), input)
		}
	}
}

// Parse the target as "table timeColumn valueColumn [tagColumn]*"
func (this *sqliteTimeSeriesManager) target2tokens(target string) (string, string, string, []string) {
	tokens := strings.Split(target, " ")
	var tableName string
	if len(tokens) > 0 {
		tableName = tokens[0]
		if !this.tables.Contains(tableName) {
			return "", "", "", []string{}
		}
	}
	var timeColumn string
	if len(tokens) > 1 {
		timeColumn = tokens[1]
	}
	var valueColumn string
	if len(tokens) > 2 {
		valueColumn = tokens[2]
	}
	var tagColumns []string
	if len(tokens) > 3 {
		tagColumns = tokens[3:]
	}
	return tableName, timeColumn, valueColumn, tagColumns
}

func newFromDb(db *sql.DB, tables []string) TimeSeriesManager {
	tableMap := NewSet(tables...)
	return &sqliteTimeSeriesManager{db: db, tables: tableMap}
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
