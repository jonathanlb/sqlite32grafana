package sqlite3

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type DataPoint struct {
	Time  int64
	Value float64
}

type TagKey struct {
	Type string
	Text string
}

type TimeSeriesManager interface {
	GetTimeSeries(target string, from string, to string, dest *map[string][]DataPoint) error
	GetTagKeys(tableName string, dest *[]TagKey) error
	GetTagValues(tableName string, key string, dest *[]string) error
}

type sqliteTimeSeriesManager struct {
	db     *sql.DB
	tables *set
}

var integerSqlTypes = NewSet("int", "integer", "tinyint")

func (this *sqliteTimeSeriesManager) GetTimeSeries(target string, from string, to string, dest *map[string][]DataPoint) error {
	tableName, timeColumn, valueColumn, tagColumns := this.target2tokens(target)
	if tableName == "" || timeColumn == "" || valueColumn == "" {
		return errors.New(fmt.Sprintf(`malformed target "%s"`, target))
	}

	var colBuilder strings.Builder
	colBuilder.WriteString(fmt.Sprintf("%s, %s", timeColumn, valueColumn))
	for _, i := range tagColumns {
		colBuilder.WriteString(", ")
		colBuilder.WriteString(i)
	}
	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s >= ? AND %s < ? ORDER BY %s",
		colBuilder.String(), tableName, timeColumn, timeColumn, timeColumn)
	log.Println(query)

	rows, err := this.db.Query(query, from, to)
	if err != nil {
		return err
	}
	result := make(map[string][]DataPoint)
	var values []interface{}
	tag := valueColumn
	for rows.Next() {
		if values == nil {
			values, err = getScanDest(rows)
			if err != nil {
				return err
			}
		}

		if err := rows.Scan(values...); err != nil {
			return errors.New(fmt.Sprintf("Cannot scan row: %v", err))
		}

		timeMillis, ok := values[0].(*int64) // XXX could be string date
		if !ok {
			ts, err := time.Parse(time.RFC3339, *values[0].(*string))
			if err != nil {
				return errors.New(fmt.Sprintf("Cannot parse time: %v", err))
			}
			millis := int64(ts.UnixNano() / 100000)
			timeMillis = &millis
		}
		value, ok := values[1].(*float64)
		if !ok {
			v := float64(*values[1].(*int64))
			value = &v
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
		newPoint := DataPoint{Time: *timeMillis, Value: *value}
		result[tag] = append(result[tag], newPoint)
	}
	*dest = result
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

// Get scan value destinations ala https://github.com/golang/go/blob/master/src/database/sql/sql_test.go
// ColumnTypes not available until after Rows.Next() called https://github.com/mattn/go-sqlite3/issues/682
func getScanDest(rows *sql.Rows) ([]interface{}, error) {
	tt, err := rows.ColumnTypes()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Cannot infer column types from row: %v", err))
	}
	types := make([]reflect.Type, len(tt))
	for i, tp := range tt {
		st := tp.ScanType()
		if st == nil {
			return nil, errors.New(fmt.Sprintf("Cannot infer column type for %q", tp.Name()))
		}
		types[i] = st
	}
	values := make([]interface{}, len(tt))
	for i := range values {
		values[i] = reflect.New(types[i]).Interface()
	}
	return values, nil
}

func (this *sqliteTimeSeriesManager) getSchema(tableName string, dest *[]TagKey) error {
	query := fmt.Sprintf("PRAGMA table_info(%s)", tableName)
	log.Printf("get Schema %s", query)
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
		return errors.New(fmt.Sprintf(`nonExistentTable: "%s"`, tableName))
	}

	return nil
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

func sql2grafanaType(sqlType string) string {
	switch strings.ToLower(sqlType) {
	case "int":
		return "number"
	case "text":
		return "string"
	default:
		log.Printf("Unknown sql column type: %s", sqlType)
		return "???"
	}
}
