package ginadapter

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yukihito-jokyu/topic2html/backend/observability"
	usecaseauth "github.com/yukihito-jokyu/topic2html/backend/usecase/auth"
)

type sessionRouterService struct {
	bootstrapOutput usecaseauth.SessionBootstrapOutput
	bootstrapErr    error
	logoutDecision  usecaseauth.GuardDecision
	logoutErr       error
	logoutInput     usecaseauth.SessionInput
}

func (s *sessionRouterService) Bootstrap(context.Context, string) (usecaseauth.SessionBootstrapOutput, error) {
	return s.bootstrapOutput, s.bootstrapErr
}
func (*sessionRouterService) AuthorizeRead(context.Context, string) (usecaseauth.GuardDecision, error) {
	return usecaseauth.GuardUnauthenticated, nil
}
func (s *sessionRouterService) Logout(_ context.Context, input usecaseauth.SessionInput) (usecaseauth.GuardDecision, error) {
	s.logoutInput = input

	return s.logoutDecision, s.logoutErr
}
func (*sessionRouterService) RunAuthorizedAdminStateChange(context.Context, usecaseauth.SessionInput, usecaseauth.AdminStateChangeOperation) (usecaseauth.GuardDecision, error) {
	return usecaseauth.GuardUnauthenticated, nil
}

func TestSessionHandlerBootstrap(t *testing.T) {
	for _, tt := range []struct {
		name       string
		service    sessionRouterService
		withCookie bool
		wantStatus int
		wantBody   string
		wantDelete bool
	}{
		{
			name: "authenticated",
			service: sessionRouterService{
				bootstrapOutput: usecaseauth.SessionBootstrapOutput{
					Authenticated: true,
					CSRFToken:     "csrf",
				},
			},
			withCookie: true,
			wantStatus: http.StatusOK,
			wantBody:   `"csrf_token":"csrf"`,
		},
		{
			name:       "anonymous invalid cookie",
			service:    sessionRouterService{},
			withCookie: true,
			wantStatus: http.StatusOK,
			wantBody:   `"authenticated":false`,
			wantDelete: true,
		},
		{
			name:       "anonymous without cookie",
			service:    sessionRouterService{},
			wantStatus: http.StatusOK,
			wantBody:   `"authenticated":false`,
		},
		{
			name: "unavailable",
			service: sessionRouterService{
				bootstrapErr: errors.New("database"),
			},
			withCookie: true,
			wantStatus: http.StatusServiceUnavailable,
			wantBody:   `"authentication_unavailable"`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/admin/auth/session", nil)
			if tt.withCookie {
				request.AddCookie(&http.Cookie{Name: adminSessionCookie, Value: "session"})
			}
			response := httptest.NewRecorder()
			NewRouter(noOpOAuthService{}, &tt.service, observability.NewDiscardLogger()).ServeHTTP(response, request)
			if response.Code != tt.wantStatus || !strings.Contains(response.Body.String(), tt.wantBody) || response.Header().Get("Cache-Control") != "no-store" {
				t.Fatalf("status/body/cache=%d/%q/%q", response.Code, response.Body.String(), response.Header().Get("Cache-Control"))
			}
			if tt.wantDelete != strings.Contains(response.Header().Get("Set-Cookie"), "Max-Age=0") {
				t.Fatalf("Set-Cookie=%q", response.Header().Get("Set-Cookie"))
			}
		})
	}
}

func TestSessionHandlerLogout(t *testing.T) {
	for _, tt := range []struct {
		name       string
		service    sessionRouterService
		wantStatus int
		wantBody   string
		wantDelete bool
	}{
		{
			name: "success",
			service: sessionRouterService{
				logoutDecision: usecaseauth.GuardAllowed,
			},
			wantStatus: http.StatusOK,
			wantBody:   `"authenticated":false`,
			wantDelete: true,
		},
		{
			name: "forbidden",
			service: sessionRouterService{
				logoutDecision: usecaseauth.GuardForbidden,
			},
			wantStatus: http.StatusForbidden,
			wantBody:   `"forbidden"`,
		},
		{
			name: "unavailable",
			service: sessionRouterService{
				logoutErr: errors.New("database"),
			},
			wantStatus: http.StatusServiceUnavailable,
			wantBody:   `"authentication_unavailable"`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/admin/auth/logout", nil)
			request.Header.Set("Origin", "https://admin.example.test")
			request.Header.Set("X-CSRF-Token", "csrf")
			request.AddCookie(&http.Cookie{Name: adminSessionCookie, Value: "session"})
			response := httptest.NewRecorder()
			NewRouter(noOpOAuthService{}, &tt.service, observability.NewDiscardLogger()).ServeHTTP(response, request)
			if response.Code != tt.wantStatus || !strings.Contains(response.Body.String(), tt.wantBody) || response.Header().Get("Cache-Control") != "no-store" {
				t.Fatalf("status/body/cache=%d/%q/%q", response.Code, response.Body.String(), response.Header().Get("Cache-Control"))
			}
			if tt.wantDelete != strings.Contains(response.Header().Get("Set-Cookie"), "Max-Age=0") {
				t.Fatalf("Set-Cookie=%q", response.Header().Get("Set-Cookie"))
			}
			if tt.service.logoutInput.CSRFToken != "csrf" || tt.service.logoutInput.SessionReference != "session" {
				t.Fatalf("logout input=%+v", tt.service.logoutInput)
			}
		})
	}
}
