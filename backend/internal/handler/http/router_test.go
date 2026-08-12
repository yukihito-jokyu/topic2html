package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yukihito-jokyu/topic2html/backend/internal/domain"
	httphandler "github.com/yukihito-jokyu/topic2html/backend/internal/handler/http"
)

type readinessChecker struct {
	result domain.Readiness
}

func (checker readinessChecker) Check(context.Context) domain.Readiness {
	return checker.result
}

func TestNewRouterReturnsReadinessResult(t *testing.T) {
	t.Parallel()

	router := httphandler.NewRouter(readinessChecker{result: domain.Readiness{Ready: true}})
	request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/readyz", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	if body := recorder.Body.String(); body != "{\"ready\":true}" {
		t.Fatalf("body = %q, want ready JSON", body)
	}
}
