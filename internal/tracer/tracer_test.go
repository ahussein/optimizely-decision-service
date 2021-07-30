// +build unit

package tracer_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ahussein/optimizely-decision-service/internal/tracer"
	"github.com/stretchr/testify/assert"
)

func TestTracingMiddleware(t *testing.T) {
	server := httptest.NewServer(tracer.TracingMiddleware("myspan")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})))
	defer server.Close()
	res, err := http.Get(server.URL)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, res.StatusCode)
}
