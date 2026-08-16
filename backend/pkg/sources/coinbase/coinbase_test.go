package coinbase

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"candle-engine/pkg/sources"
)

// newStubExchange serves /products/{pair}/candles the way Coinbase does: each
// response holds the candles inside the requested [start, end] window, newest
// first, capped at 300 entries.
func newStubExchange(t *testing.T, first time.Time, count, granularity int) *httptest.Server {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		start, err := time.Parse("2006-01-02T15:04:05Z", q.Get("start"))
		if err != nil {
			t.Errorf("bad start param %q: %v", q.Get("start"), err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		end, err := time.Parse("2006-01-02T15:04:05Z", q.Get("end"))
		if err != nil {
			t.Errorf("bad end param %q: %v", q.Get("end"), err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var out [][]float64
		// Walk newest → oldest, as the real API does.
		for i := count - 1; i >= 0; i-- {
			ts := first.Add(time.Duration(i*granularity) * time.Second)
			if ts.Before(start) || ts.After(end) {
				continue
			}
			price := float64(1000 + i)
			// [time, low, high, open, close, volume]
			out = append(out, []float64{float64(ts.Unix()), price - 1, price + 1, price, price, 10})
			if len(out) == 300 {
				break
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(out); err != nil {
			t.Errorf("encode response: %v", err)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

// TestGetData_MultiSegmentIsChronological guards the ordering bug behind the
// chart rendering candles from outside the selected range: a range wider than
// one 300-candle page is fetched in several segments, and each segment arrives
// newest-first. Reversing only the concatenated result put the segments
// themselves in descending order, so the returned slice was not sorted.
func TestGetData_MultiSegmentIsChronological(t *testing.T) {
	const granularity = 60 // "1m"
	const count = 700      // three segments of 300/300/100

	first := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	srv := newStubExchange(t, first, count, granularity)

	src := &CoinbaseSource{Name: "Coinbase", BaseURL: srv.URL}

	timeRange := sources.TimeRange{
		Start: first,
		End:   first.Add(time.Duration((count-1)*granularity) * time.Second),
	}

	entries, err := src.GetData("BTC-USDT", "1m", timeRange)
	if err != nil {
		t.Fatalf("GetData failed: %v", err)
	}
	if len(entries) != count {
		t.Fatalf("expected %d entries, got %d", count, len(entries))
	}

	for i, e := range entries {
		want := first.Add(time.Duration(i*granularity) * time.Second)
		if !e.Timestamp.Equal(want) {
			t.Fatalf("entry %d: expected timestamp %v, got %v", i, want, e.Timestamp)
		}
	}
}
