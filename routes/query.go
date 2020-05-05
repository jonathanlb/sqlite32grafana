package routes

import (
	"encoding/json"

	"github.com/gofiber/fiber"
	"github.com/jonathanlb/sqlite32grafana/sqlite3"
)

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

type QueryPayload struct {
	Range         QueryRange
	RangeRaw      QueryRangeRaw
	Interval      string
	IntervalMs    int32
	Targets       []QueryTarget
	AdhocFilters  []QueryFilter
	Format        string
	MaxDataPoints int32
}

type Timeseries struct {
	Target     string
	DataPoints [][]float64
}

func InstallQuery(app *fiber.App, tsm sqlite3.TimeSeriesManager) {
	app.Post("/query", func(c *fiber.Ctx) {
		var query QueryPayload
		body := []byte(c.Body())
		err := json.Unmarshal(body, &query)
		if err == nil {
			err = validateQuery(&query)
		}
		if err != nil {
			c.SendStatus(400)
			c.SendString(err.Error())
			return
		}

		result := []Timeseries{}
		from, to := query.Range.From, query.Range.To
		for _, target := range query.Targets {
			var series map[string][]sqlite3.DataPoint
			if err := tsm.GetTimeSeries(target.Target, from, to, &series); err != nil {
				c.SendStatus(400)
				c.SendString(err.Error())
				return
			}
			for key, data := range series {
				result = append(result, Timeseries{
					Target:     key,
					DataPoints: datapointsToArray(data),
				})
			}
			// TODO: Filter and limit points
		}

		resultBytes, err := json.Marshal(result)
		if err != nil {
			c.SendStatus(400)
			c.SendString(err.Error())
			return
		}
		c.Send(resultBytes)
		c.SendStatus(200)
	})
}

func datapointsToArray(pts []sqlite3.DataPoint) [][]float64 {
	arr := make([][]float64, len(pts))
	for i, p := range pts {
		arr[i] = make([]float64, 2)
		arr[i][0] = p.Value
		arr[i][1] = float64(p.Time)
	}
	return arr
}

func validateQuery(query *QueryPayload) error {
	// XXX TODO
	return nil
}
