package sqlite3

type DataPoint struct {
	Time  int64
	Value float64
}

type TagKey struct {
	Type string
	Text string
}

type QueryRangeRaw struct {
	From string
	To   string
}

type QueryRange struct {
	From string
	To   string
	Raw  QueryRangeRaw
}

type QueryTarget struct {
	Target string
	RefId  string
	Type   string
}

type QueryFilter struct {
	Key      string
	Operator string
	Value    string
}

type TimeSeriesQueryOpts struct {
	Interval      string
	MaxDataPoints int32
	Filters       []QueryFilter
}

type TimeSeriesManager interface {
	GetTimeSeries(target string, fromTo *QueryRange, opts *TimeSeriesQueryOpts, dest *map[string][]DataPoint) error
	GetTagKeys(tableName string, dest *[]TagKey) error
	GetTagValues(tableName string, key string, dest *[]string) error
}
