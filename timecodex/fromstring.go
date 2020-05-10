package timecodex

import (
	"regexp"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

var isYyyymmdd = regexp.MustCompile(`^[0-9]{2,4}[/\- ][0-9]{1,2}[/\- ][0-9]{1,2}$`)

// Interpret a string as a time.
// Possibilities are:
//  - RFC3339
//  - YYYYY-MM-DD (slash, hyphen, or space separators)
//  - Integer treated as seconds or milliseconds from January 1, 1970 UTC
func StringToTime(dateTimeStr string) (time.Time, error) {
	result, err := time.Parse(time.RFC3339, dateTimeStr)
	if err == nil {
		return result, err
	}

	if isYyyymmdd.MatchString(dateTimeStr) {
		return time.Parse(time.RFC3339, dateTimeStr+"T0:00:00Z")
	}

	epochs, err := strconv.Atoi(dateTimeStr)
	if err != nil {
		return NumberToTime(int64(epochs)), nil
	}

	return time.Unix(0, 0), errors.Errorf(`cannot parse datetime "%s"`, dateTimeStr)
}
