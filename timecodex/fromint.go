package timecodex

import (
	"time"
)

var secondsThreshold int64 = 32000000000
var nanoThreshold int64 = secondsThreshold * 1000

func NumberToScalar(epoch int64) (int64, bool) {
	if epoch < secondsThreshold && epoch >= -secondsThreshold {
		return 1000, true
	}
	if epoch > nanoThreshold || epoch <= -nanoThreshold {
		return 1000000, false
	}
	return 1, true
}

// Attempt to interpret an integral value as a timestamp.
// - Absolute values less than 3.2e10 will be treated as epoch seconds, yielding
// times about a millenia from the epoch.
// - Absolute values greater than 3.2e13 will be treated as nanoseconds from the
// epoch.
// - Values in between are treated as milliseconds since the epoch.
func NumberToTime(epoch int64) time.Time {
	if epoch < secondsThreshold && epoch >= -secondsThreshold {
		return time.Unix(epoch, 0)
	}
	if epoch > nanoThreshold || epoch <= -nanoThreshold {
		return time.Unix(0, epoch)
	}
	return time.Unix(0, epoch*1000000)
}
