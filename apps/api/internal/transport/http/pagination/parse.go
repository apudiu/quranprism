// Package pagination is the repo-wide ?limit=&offset= convention for
// list endpoints. Defaults limit=20, max 100. Negative or non-int values
// return a typed BadRequest so the handler can pass it to httperr.Write
// directly.
//
// Response envelope used by callers (compose with response.JSON):
//
//	response.JSON(w, 200, map[string]any{
//	    "items":  items,
//	    "total":  total,
//	    "limit":  limit,
//	    "offset": offset,
//	})
package pagination

import (
	"net/http"
	"strconv"

	"github.com/apudiu/quranprism/api/internal/transport/http/httperr"
)

const (
	DefaultLimit = 20
	MaxLimit     = 100
)

// Parse reads `?limit=` and `?offset=` from the request and returns
// safe values. Missing query params get defaults; oversized limits are
// clamped to MaxLimit; negative or non-numeric values return a
// *httperr.E with status 400.
func Parse(r *http.Request) (limit, offset int, err error) {
	limit = DefaultLimit
	offset = 0

	if s := r.URL.Query().Get("limit"); s != "" {
		v, perr := strconv.Atoi(s)
		if perr != nil || v < 1 {
			return 0, 0, httperr.BadRequest("invalid limit")
		}
		if v > MaxLimit {
			v = MaxLimit
		}
		limit = v
	}

	if s := r.URL.Query().Get("offset"); s != "" {
		v, perr := strconv.Atoi(s)
		if perr != nil || v < 0 {
			return 0, 0, httperr.BadRequest("invalid offset")
		}
		offset = v
	}

	return limit, offset, nil
}
