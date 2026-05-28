package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/apudiu/quranprism/api/internal/transport/http/httperr"
)

// rateLimitScript is a tiny Lua atom:
//   - INCR key
//   - if the post-incr count is 1 (first hit), set the window TTL
//   - return the post-incr count
//
// Atomic so a race between two requests can't lose the TTL set.
var rateLimitScript = redis.NewScript(`
local n = redis.call('INCR', KEYS[1])
if n == 1 then
  redis.call('PEXPIRE', KEYS[1], ARGV[1])
end
return n
`)

// KeyFunc derives the bucket key from the request. Returning "" skips
// the limit entirely (e.g. for an internal-only header). Most callers
// use IPKeyFunc; auth handlers compose it with EmailKeyFunc.
type KeyFunc func(r *http.Request) string

// IPKeyFunc keys the bucket by the client IP, honouring X-Forwarded-For
// when present (the first hop is the client). When neither is parseable
// it falls back to a wildcard "anon" so traffic isn't dropped silently.
func IPKeyFunc(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// First comma-separated address is the originating client.
		for i := 0; i < len(xff); i++ {
			if xff[i] == ',' {
				return xff[:i]
			}
		}
		return xff
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "anon"
	}
	return host
}

// RateLimit returns a middleware that allows `limit` requests per
// `window` per (scope, key). The Redis key namespace is shared with
// other application caches but prefixed so it can't collide.
//
// Logs the per-request count at debug level; exceeded buckets log at
// warn and return 429.
func RateLimit(rdb *redis.Client, log *slog.Logger, scope string, limit int, window time.Duration, key KeyFunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			k := key(r)
			if k == "" {
				next.ServeHTTP(w, r)
				return
			}
			redisKey := fmt.Sprintf("ratelimit:%s:%s", scope, k)
			ctx, cancel := context.WithTimeout(r.Context(), 100*time.Millisecond)
			defer cancel()

			res, err := rateLimitScript.Run(ctx, rdb, []string{redisKey}, window.Milliseconds()).Int64()
			if err != nil {
				// Failing open is the right call: a Redis blip should not
				// take down auth. Log loudly so the on-call sees it.
				log.Warn("rate limit redis err — failing open", "scope", scope, "key", k, "err", err)
				next.ServeHTTP(w, r)
				return
			}
			if int(res) > limit {
				log.Warn("rate limit exceeded", "scope", scope, "key", k, "count", res, "limit", limit)
				httperr.Write(w, log, httperr.TooManyRequests())
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
