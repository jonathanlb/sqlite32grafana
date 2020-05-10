package sqlite3

import (
	"database/sql"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

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

func (this *sqliteTimeSeriesManager) GetTimeSeries(target string, from string, to string, dest *map[string][]DataPoint) error {
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
	fromTime, err := this.getTime(tableName, timeColumn, from)
	if err != nil {
		return errors.Wrap(err, "get from time for timeseries")
	}
	toTime, err := this.getTime(tableName, timeColumn, to)
	if err != nil {
		return errors.Wrap(err, "get to time for timeseries")
	}

	timeReader := this.getTimeReader(tableName, timeColumn) // XXX memoize?
	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s >= ? AND %s < ? ORDER BY %s",
		colBuilder.String(), tableName, timeColumn, timeColumn, timeColumn)
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

	ts, err := parseTime(*dateStr)
	if err != nil {
		return 0, errors.Errorf("Cannot parse time: %v", err)
	}
	millis := int64(ts.UnixNano() / 100000)
	return millis, nil
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

// Convert user-supplied time string to one comparable to the stated type of the column
func (this *sqliteTimeSeriesManager) getTime(tableName string, timeColumn string, timeStr string) (interface{}, error) {
	columnType := this.getColumnType(tableName, timeColumn)
	switch strings.ToLower(columnType) {
	case "int":
		return millisToMillis(timeStr)
	case "datetime":
		return timeStr, nil
	case "text":
		return timeStr, nil
	default:
		return nil, errors.Errorf("unknown time type %s for time column %s in table %s", columnType, timeColumn, tableName)
	}
}

func (this *sqliteTimeSeriesManager) getTimeReader(tableName string, timeColumn string) func(input interface{}) (int64, error) {
	columnType := this.getColumnType(tableName, timeColumn)
	switch strings.ToLower(columnType) {
	case "int":
		return millisToMillis
	case "datetime":
		return dateTimeToMillis
	case "text":
		return dateTimeToMillis
	default:
		sugar.Panicf("unknown time type %s for time column %s in table %s", columnType, timeColumn, tableName)
		return nil
	}
}

func millisToMillis(input interface{}) (int64, error) {
	millis, ok := input.(*int64)
	if ok {
		return *millis, nil
	}

	timeStr, ok := input.(string)
	if !ok {
		return 0,
			errors.Errorf("cannot cast time column of type %v to int64 or string", reflect.TypeOf(input))
	}
	t, err := parseTime(timeStr)
	if err != nil {
		return 0, err
	}
	return t.UnixNano() / 1000000, err
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

func parseTime(timeStr string) (time.Time, error) {
	result, err := time.Parse(time.RFC3339, timeStr)
	if err == nil {
		return result, err
	}

	if yyyymmdd.MatchString(timeStr) {
		return time.Parse(time.RFC3339, timeStr+"T0:00:00Z")
	}

	if intStr.MatchString(timeStr) {
		epochMillis, err := strconv.Atoi(timeStr)
		if err == nil {
			s := (int64)(epochMillis / 1000)
			ns := (int64)((epochMillis % 1000) * 1000000)
			t := time.Unix(s, ns)
			return t, nil
		}
	}
	return time.Unix(0, 0), errors.Errorf(`cannot parse datetime "%s"`, timeStr)
}

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
