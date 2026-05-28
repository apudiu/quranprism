package pagination

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/apudiu/quranprism/api/internal/transport/http/httperr"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		wantLimit int
		wantOff   int
		wantErr   bool
	}{
		{"defaults", "", 20, 0, false},
		{"limit_set", "limit=50", 50, 0, false},
		{"offset_set", "offset=15", 20, 15, false},
		{"limit_clamped", "limit=500", 100, 0, false},
		{"limit_at_max", "limit=100", 100, 0, false},
		{"limit_zero_rejected", "limit=0", 0, 0, true},
		{"limit_negative", "limit=-1", 0, 0, true},
		{"limit_nonint", "limit=abc", 0, 0, true},
		{"offset_negative", "offset=-1", 0, 0, true},
		{"offset_nonint", "offset=xyz", 0, 0, true},
		{"both", "limit=10&offset=20", 10, 20, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/?"+tc.query, nil)
			l, o, err := Parse(r)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				var e *httperr.E
				if !errors.As(err, &e) || e.Status != 400 {
					t.Fatalf("expected 400 httperr.E, got %#v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if l != tc.wantLimit {
				t.Errorf("limit = %d, want %d", l, tc.wantLimit)
			}
			if o != tc.wantOff {
				t.Errorf("offset = %d, want %d", o, tc.wantOff)
			}
		})
	}
}
