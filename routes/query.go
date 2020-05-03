package routes

import (
	"encoding/json"

	"github.com/gofiber/fiber"
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
	DataPoints [][]int64
}

func InstallQuery(app *fiber.App) {
	app.Post("/query", func(c *fiber.Ctx) {
		var query QueryPayload
		err := json.Unmarshal([]byte(c.Body()), &query)
		if err == nil {
			err = validateQuery(&query)
		}
		if err != nil {
			c.SendStatus(400)
			c.SendString(err.Error())
			return
		}

		// Iterate over targets to get timeseries
		// Filter and limit points

	})
}

func validateQuery(query *QueryPayload) error {
	// XXX TODO
	return nil
}
