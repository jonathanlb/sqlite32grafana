package timecodex

import (
	"testing"
	"time"
)

func Test_NumberToTime(t *testing.T) {
	ts := NumberToTime(60)
	seconds := ts.Unix()
	if seconds != 60 {
		t.Fatalf("expected 60 epoch seconds from date, got %+v", ts)
	}

	now := time.Now()
	expectedS := now.Unix()
	ts = NumberToTime(expectedS)
	seconds = ts.Unix()
	if seconds != expectedS {
		t.Fatalf("expected %d epoch seconds from now, got %+v", expectedS, ts)
	}

	expectedMillis := now.Unix() * 1000
	ts = NumberToTime(expectedMillis)
	millis := ts.Unix() * 1000
	if millis != expectedMillis {
		t.Fatalf("expected %d epoch millis from now, got %d (time %+v)",
			expectedMillis, millis, ts)
	}

	expectedNanos := now.UnixNano()
	ts = NumberToTime(expectedNanos)
	nanos := ts.UnixNano()
	if nanos != expectedNanos {
		t.Fatalf("expected %d epoch nanos from now, got %d (time %+v)",
			expectedNanos, nanos, ts)
	}
}
