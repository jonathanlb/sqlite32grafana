package routes

import (
	"encoding/json"

	"github.com/gofiber/fiber"
	"github.com/jonathanlb/sqlite32grafana/sqlite3"
)

type QueryPayload struct {
	Range         sqlite3.QueryRange
	RangeRaw      sqlite3.QueryRangeRaw
	Interval      string
	IntervalMs    int32
	Targets       []sqlite3.QueryTarget
	AdhocFilters  []sqlite3.QueryFilter
	Format        string
	MaxDataPoints int32
}

type Timeseries struct {
	Target     string      `json:"target"`
	DataPoints [][]float64 `json:"datapoints"`
}

func InstallQuery(app *fiber.App, tsm sqlite3.TimeSeriesManager) {
	app.Post("/query", func(c *fiber.Ctx) {
		var query QueryPayload
		body := []byte(c.Body())
		err := json.Unmarshal(body, &query)
		sugar.Debugw("route query", "err", err, "body", string(body), "query", query)
		if err == nil {
			err = validateQuery(&query)
		}
		if err != nil {
			c.SendStatus(400)
			c.SendString(err.Error())
			return
		}

		queryOpts := sqlite3.TimeSeriesQueryOpts{
			Interval:      query.Interval,
			MaxDataPoints: query.MaxDataPoints,
			Filters:       query.AdhocFilters,
		}

		result := []Timeseries{}
		for _, target := range query.Targets {
			// XXX switch on target.Type
			var series map[string][]sqlite3.DataPoint
			if err := tsm.GetTimeSeries(target.Target, &query.Range, &queryOpts, &series); err != nil {
				c.SendStatus(400)
				c.SendString(err.Error())
				return
			} // XXX it looks like response spec is one series per target?
			for _, data := range series {
				result = append(result, Timeseries{
					Target:     target.Target,
					DataPoints: datapointsToArray(data),
				})
			}
		}

		resultBytes, err := json.Marshal(result)
		if err != nil {
			c.SendStatus(400)
			c.SendString(err.Error())
			return
		}
		c.Set("Content-Type", "application/json")
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
