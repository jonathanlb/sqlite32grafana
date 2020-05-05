package sqlite3

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
