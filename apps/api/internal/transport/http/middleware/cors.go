package middleware

import "net/http"

// CORS returns a permissive-but-scoped CORS middleware. Allows the
// configured app base URL (typically the SolidStart frontend) plus
// credentials so the refresh cookie can flow on /v1/auth/refresh.
//
// `origin` is the single allowed Origin value. For multi-origin setups
// (eg. staging + prod sharing the same api) extend to a set match.
func CORS(origin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			h.Set("Vary", "Origin")
			if r.Header.Get("Origin") == origin {
				h.Set("Access-Control-Allow-Origin", origin)
				h.Set("Access-Control-Allow-Credentials", "true")
				h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				h.Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
				h.Set("Access-Control-Max-Age", "600")
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
