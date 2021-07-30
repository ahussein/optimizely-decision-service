package tracer

import (
	"fmt"
	"net/http"

	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/trace"
)

// TracingMiddleware adds tracing to each request
func TracingMiddleware(spanName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handler := &ochttp.Handler{
				Handler: withAdditionalRequestAttributes(next),
				FormatSpanName: func(r *http.Request) string {
					return fmt.Sprintf(" %s", spanName)
				},
			}
			handler.ServeHTTP(w, r)
		})
	}
}

func withAdditionalRequestAttributes(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		span := trace.FromContext(r.Context())
		requestID := r.Header.Get("X-Request-ID")
		if requestID != "" {
			span.AddAttributes(trace.StringAttribute("request.id", requestID))
		}
		next.ServeHTTP(w, r)
	})
}
