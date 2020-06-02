package sqlite3

// DataPoint is a time-scalar tuple for reporting observations back to Grafana.
type DataPoint struct {
	Time  int64
	Value float64
}

// TagKey represents a column name and the declared type of column values.
type TagKey struct {
	Type string `json:"type"`
	Text string ` json:"text"`
}

// QueryRangeRaw stores the grafana time query range as entered by user.
// The range values typically have relative values, such as "now-10d".
type QueryRangeRaw struct {
	From string
	To   string
}

// QueryRange represents a time period requested by Grafana.
type QueryRange struct {
	From string
	To   string
	Raw  QueryRangeRaw
}

// QueryTarget represents a timeseries requested by Grafana or subsequence
// returned to Grafana of a requested timeseries where each observation
// is tagged by the Target field value.
type QueryTarget struct {
	Target string
	RefID  string
	Type   string
}

// QueryFilter stores a query limiter requested by Grafana.
type QueryFilter struct {
	Key      string
	Operator string
	Value    string
}

// TimeSeriesQueryOpts holds options for a query.  Currently, only the
// MaxDataPoints field is respected by sqlite32grafana.
type TimeSeriesQueryOpts struct {
	Interval      string
	MaxDataPoints int32
	Filters       []QueryFilter
}

// TimeSeriesManager exposes calls available to ReST end points to query
// SQLite table columns to Grafana.
type TimeSeriesManager interface {
	GetTimeSeries(target string, fromTo *QueryRange, opts *TimeSeriesQueryOpts, dest *map[string][]DataPoint) error
	GetTagKeys(tableName string, dest *[]TagKey) error
	GetTagValues(tableName string, key string, dest *[]string) error
}
