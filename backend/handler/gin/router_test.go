package ginadapter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewRouter(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		path string
		want int
	}{
		{name: "health", path: "/health", want: http.StatusNoContent},
		{name: "unknown", path: "/unknown", want: http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			NewRouter().ServeHTTP(response, httptest.NewRequestWithContext(context.Background(), http.MethodGet, tt.path, nil))
			if response.Code != tt.want {
				t.Errorf("status = %d, want %d", response.Code, tt.want)
			}
		})
	}
}
