package ginadapter

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/yukihito-jokyu/topic2html/backend/observability"
	usecaseauth "github.com/yukihito-jokyu/topic2html/backend/usecase/auth"
)

type routerOAuthService struct {
	startOutput   usecaseauth.StartOutput
	startErr      error
	callbackOut   usecaseauth.CallbackOutput
	callbackErr   error
	startInput    usecaseauth.StartInput
	callbackInput usecaseauth.CallbackInput
}

func (s *routerOAuthService) Start(_ context.Context, input usecaseauth.StartInput) (usecaseauth.StartOutput, error) {
	s.startInput = input

	return s.startOutput, s.startErr
}

func (s *routerOAuthService) Callback(_ context.Context, input usecaseauth.CallbackInput) (usecaseauth.CallbackOutput, error) {
	s.callbackInput = input

	return s.callbackOut, s.callbackErr
}

func TestOAuthStartHTTPContract(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		method      string
		content     string
		body        string
		origin      []string
		service     *routerOAuthService
		wantStatus  int
		wantPath    string
		wantNoStore bool
	}{
		{
			name:    "success",
			method:  http.MethodPost,
			content: "application/x-www-form-urlencoded",
			body:    "return_path=%2Fadmin",
			origin:  []string{"https://admin.example.test"},
			service: &routerOAuthService{
				startOutput: usecaseauth.StartOutput{
					TransactionReference: "tx",
					AuthorizationURL:     "https://accounts.google.test/authorize",
				},
			},
			wantStatus:  http.StatusSeeOther,
			wantPath:    "https://accounts.google.test/authorize",
			wantNoStore: true,
		},
		{
			name:        "wrong method",
			method:      http.MethodGet,
			content:     "application/x-www-form-urlencoded",
			origin:      []string{"https://admin.example.test"},
			service:     &routerOAuthService{},
			wantStatus:  http.StatusMethodNotAllowed,
			wantNoStore: false,
		},
		{
			name:        "wrong content type",
			method:      http.MethodPost,
			content:     "application/json",
			origin:      []string{"https://admin.example.test"},
			service:     &routerOAuthService{},
			wantStatus:  http.StatusSeeOther,
			wantPath:    failureRedirect,
			wantNoStore: true,
		},
		{
			name:        "malformed content type",
			method:      http.MethodPost,
			content:     "%%%",
			origin:      []string{"https://admin.example.test"},
			service:     &routerOAuthService{},
			wantStatus:  http.StatusSeeOther,
			wantPath:    failureRedirect,
			wantNoStore: true,
		},
		{
			name:        "malformed form",
			method:      http.MethodPost,
			content:     "application/x-www-form-urlencoded",
			body:        "return_path=%zz",
			origin:      []string{"https://admin.example.test"},
			service:     &routerOAuthService{},
			wantStatus:  http.StatusSeeOther,
			wantPath:    failureRedirect,
			wantNoStore: true,
		},
		{
			name:        "unknown field",
			method:      http.MethodPost,
			content:     "application/x-www-form-urlencoded",
			body:        "unexpected=x",
			origin:      []string{"https://admin.example.test"},
			service:     &routerOAuthService{},
			wantStatus:  http.StatusSeeOther,
			wantPath:    failureRedirect,
			wantNoStore: true,
		},
		{
			name:        "duplicate field",
			method:      http.MethodPost,
			content:     "application/x-www-form-urlencoded",
			body:        "return_path=%2Fadmin&return_path=%2Fadmin",
			origin:      []string{"https://admin.example.test"},
			service:     &routerOAuthService{},
			wantStatus:  http.StatusSeeOther,
			wantPath:    failureRedirect,
			wantNoStore: true,
		},
		{
			name:        "empty field",
			method:      http.MethodPost,
			content:     "application/x-www-form-urlencoded",
			body:        "return_path=",
			origin:      []string{"https://admin.example.test"},
			service:     &routerOAuthService{},
			wantStatus:  http.StatusSeeOther,
			wantPath:    failureRedirect,
			wantNoStore: true,
		},
		{
			name:    "service failure",
			method:  http.MethodPost,
			content: "application/x-www-form-urlencoded",
			origin:  []string{"https://admin.example.test"},
			service: &routerOAuthService{
				startErr: errors.New("rejected"),
			},
			wantStatus:  http.StatusSeeOther,
			wantPath:    failureRedirect,
			wantNoStore: true,
		},
		{
			name:    "empty transaction",
			method:  http.MethodPost,
			content: "application/x-www-form-urlencoded",
			origin:  []string{"https://admin.example.test"},
			service: &routerOAuthService{
				startOutput: usecaseauth.StartOutput{
					AuthorizationURL: "https://accounts.google.test/authorize",
				},
			},
			wantStatus:  http.StatusSeeOther,
			wantPath:    failureRedirect,
			wantNoStore: true,
		},
		{
			name:    "empty authorization URL",
			method:  http.MethodPost,
			content: "application/x-www-form-urlencoded",
			origin:  []string{"https://admin.example.test"},
			service: &routerOAuthService{
				startOutput: usecaseauth.StartOutput{
					TransactionReference: "tx",
				},
			},
			wantStatus:  http.StatusSeeOther,
			wantPath:    failureRedirect,
			wantNoStore: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequestWithContext(context.Background(), tt.method, "/admin/auth/google/start", strings.NewReader(tt.body))
			request.Header.Set("Content-Type", tt.content)
			for _, origin := range tt.origin {
				request.Header.Add("Origin", origin)
			}
			if tt.name == "success" {
				request.AddCookie(&http.Cookie{
					Name:     oauthTransactionCookie,
					Value:    "old",
					Secure:   true,
					HttpOnly: true,
					SameSite: http.SameSiteLaxMode,
				})
			}
			response := httptest.NewRecorder()
			NewRouter(tt.service, noOpAdminSessionService{}, observability.NewDiscardLogger()).ServeHTTP(response, request)
			if response.Code != tt.wantStatus || response.Header().Get("Location") != tt.wantPath {
				t.Fatalf("status/location = %d/%q, want %d/%q", response.Code, response.Header().Get("Location"), tt.wantStatus, tt.wantPath)
			}
			if gotNoStore := response.Header().Get("Cache-Control") == "no-store"; gotNoStore != tt.wantNoStore {
				t.Errorf("no-store = %t, want %t", gotNoStore, tt.wantNoStore)
			}
			if tt.name == "success" {
				if !strings.Contains(response.Header().Get("Set-Cookie"), "Max-Age=600") {
					t.Error("transaction cookie was not set with the contract lifetime")
				}
				if tt.service.startInput.PreviousReference != "old" || len(tt.service.startInput.Origins) != 1 {
					t.Errorf("start input = %+v", tt.service.startInput)
				}
			}
		})
	}
}

func TestOAuthCallbackHTTPContract(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		query      string
		service    *routerOAuthService
		wantStatus int
		wantPath   string
	}{
		{
			name:  "success",
			query: "?state=state&code=code",
			service: &routerOAuthService{
				callbackOut: usecaseauth.CallbackOutput{
					SessionReference: "session",
					ReturnPath:       "/admin",
				},
			},
			wantStatus: http.StatusSeeOther,
			wantPath:   "/admin",
		},
		{
			name:  "service failure",
			query: "?state=state&code=code",
			service: &routerOAuthService{
				callbackErr: errors.New("rejected"),
			},
			wantStatus: http.StatusSeeOther,
			wantPath:   failureRedirect,
		},
		{
			name:  "empty session",
			query: "?state=state&code=code",
			service: &routerOAuthService{
				callbackOut: usecaseauth.CallbackOutput{
					ReturnPath: "/admin",
				},
			},
			wantStatus: http.StatusSeeOther,
			wantPath:   failureRedirect,
		},
		{
			name:  "invalid return path",
			query: "?state=state&code=code",
			service: &routerOAuthService{
				callbackOut: usecaseauth.CallbackOutput{
					SessionReference: "session",
					ReturnPath:       "/other",
				},
			},
			wantStatus: http.StatusSeeOther,
			wantPath:   failureRedirect,
		},
		{
			name:  "provider error",
			query: "?state=state&error=access_denied",
			service: &routerOAuthService{
				callbackErr: errors.New("cancelled"),
			},
			wantStatus: http.StatusSeeOther,
			wantPath:   failureRedirect,
		},
		{
			name:       "duplicate state",
			query:      "?state=a&state=b&code=code",
			service:    &routerOAuthService{},
			wantStatus: http.StatusSeeOther,
			wantPath:   failureRedirect,
		},
		{
			name:       "duplicate code",
			query:      "?state=state&code=a&code=b",
			service:    &routerOAuthService{},
			wantStatus: http.StatusSeeOther,
			wantPath:   failureRedirect,
		},
		{
			name:       "duplicate provider error",
			query:      "?state=state&error=a&error=b",
			service:    &routerOAuthService{},
			wantStatus: http.StatusSeeOther,
			wantPath:   failureRedirect,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/auth/google/callback"+tt.query, nil)
			request.AddCookie(&http.Cookie{
				Name:     oauthTransactionCookie,
				Value:    "tx",
				Secure:   true,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			})
			response := httptest.NewRecorder()
			NewRouter(tt.service, noOpAdminSessionService{}, observability.NewDiscardLogger()).ServeHTTP(response, request)
			if response.Code != tt.wantStatus || response.Header().Get("Location") != tt.wantPath {
				t.Fatalf("status/location = %d/%q, want %d/%q", response.Code, response.Header().Get("Location"), tt.wantStatus, tt.wantPath)
			}
			if response.Header().Get("Cache-Control") != "no-store" {
				t.Error("missing no-store")
			}
			if !strings.Contains(response.Header().Get("Set-Cookie"), oauthTransactionCookie) {
				t.Error("transaction cookie was not deleted")
			}
			if tt.name == "success" && !strings.Contains(strings.Join(response.Header().Values("Set-Cookie"), "\n"), "Max-Age=28800") {
				t.Error("session cookie was not set with the contract lifetime")
			}
			if tt.name == "success" && (tt.service.callbackInput.TransactionReference != "tx" || tt.service.callbackInput.State != "state") {
				t.Errorf("callback input = %+v", tt.service.callbackInput)
			}
		})
	}
}

func TestRouterHelpers(t *testing.T) {
	t.Parallel()
	panicRouter := gin.New()
	panicRouter.Use(safeRecovery(observability.NewDiscardLogger()))
	panicRouter.GET("/panic", func(*gin.Context) { panic("secret") })
	panicResponse := httptest.NewRecorder()
	panicRouter.ServeHTTP(panicResponse, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/panic", nil))
	if panicResponse.Code != http.StatusInternalServerError {
		t.Fatalf("panic status = %d", panicResponse.Code)
	}
	if NewRouter(noOpOAuthService{}, noOpAdminSessionService{}, observability.NewDiscardLogger()) == nil {
		t.Fatal("router is nil")
	}
	formRequest := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/", strings.NewReader("return_path=%2Fadmin"))
	formRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if !isFormPost(formRequest) {
		t.Fatal("valid form was rejected")
	}
	if isFormPost(httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/", nil)) {
		t.Fatal("missing content type was accepted")
	}
	if got, ok := parseReturnPath(httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/", nil)); !ok || got != nil {
		t.Fatalf("missing return path = %#v, %t", got, ok)
	}
	for _, testCase := range []struct {
		name string
		body string
	}{
		{
			name: "valid return path",
			body: "return_path=%2Fadmin",
		},
		{
			name: "unexpected parameter",
			body: "unexpected=x",
		},
		{
			name: "duplicate return path",
			body: "return_path=%2Fadmin&return_path=%2Fadmin",
		},
		{
			name: "empty return path",
			body: "return_path=",
		},
		{
			name: "malformed form encoding",
			body: "return_path=%zz",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/", strings.NewReader(testCase.body))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			_, _ = parseReturnPath(request)
		})
	}
	for _, test := range []struct {
		values url.Values
		key    string
		want   bool
	}{
		{
			values: url.Values{},
			key:    "missing",
			want:   true,
		},
		{
			values: url.Values{
				"value": {"one"},
			},
			key:  "value",
			want: true,
		},
		{
			values: url.Values{
				"value": {"one", "two"},
			},
			key:  "value",
			want: false,
		},
	} {
		t.Run(test.key, func(t *testing.T) {
			_, ok := singleQueryValue(test.values, test.key)
			if ok != test.want {
				t.Errorf("singleQueryValue(%v) = %t, want %t", test.values, ok, test.want)
			}
		})
	}
}
