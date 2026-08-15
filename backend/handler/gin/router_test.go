package ginadapter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yukihito-jokyu/topic2html/backend/observability"
	usecaseauth "github.com/yukihito-jokyu/topic2html/backend/usecase/auth"
)

type noOpOAuthService struct{}

func (noOpOAuthService) Start(context.Context, usecaseauth.StartInput) (usecaseauth.StartOutput, error) {
	return usecaseauth.StartOutput{}, nil
}

func (noOpOAuthService) Callback(context.Context, usecaseauth.CallbackInput) (usecaseauth.CallbackOutput, error) {
	return usecaseauth.CallbackOutput{}, nil
}

func TestNewRouter(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		path string
		want int
	}{
		{
			name: "health",
			path: "/health",
			want: http.StatusNoContent,
		},
		{
			name: "unknown",
			path: "/unknown",
			want: http.StatusNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			NewRouter(noOpOAuthService{}, observability.NewDiscardLogger()).ServeHTTP(response, httptest.NewRequestWithContext(context.Background(), http.MethodGet, tt.path, nil))
			if response.Code != tt.want {
				t.Errorf("status = %d, want %d", response.Code, tt.want)
			}
			if response.Header().Get("Cache-Control") != "" {
				t.Error("public route received no-store")
			}
		})
	}
}
