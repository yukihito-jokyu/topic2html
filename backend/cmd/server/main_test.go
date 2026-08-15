package main

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yukihito-jokyu/topic2html/backend/repository/security"
	usecaseauth "github.com/yukihito-jokyu/topic2html/backend/usecase/auth"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name        string
		lookup      LookupEnv
		serveResult error
		wantError   bool
		wantStarted bool
	}{
		{
			name: "invalid configuration",
			lookup: func(string) (string, bool) {
				return "", false
			},
			wantError: true,
		},
		{
			name:        "successful serve",
			lookup:      lookup(validEnvironment()),
			wantStarted: true,
		},
		{
			name:        "closed server",
			lookup:      lookup(validEnvironment()),
			serveResult: http.ErrServerClosed,
			wantStarted: true,
		},
		{
			name:        "serve failure",
			lookup:      lookup(validEnvironment()),
			serveResult: errors.New("serve failed"),
			wantError:   true,
			wantStarted: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			started := false
			err := run(tt.lookup, func(server *http.Server) error {
				started = true
				response := httptest.NewRecorder()
				server.Handler.ServeHTTP(response, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/health", nil))
				if response.Code != http.StatusNoContent {
					t.Errorf("health status = %d, want %d", response.Code, http.StatusNoContent)
				}

				return tt.serveResult
			})
			if (err != nil) != tt.wantError {
				t.Fatalf("run() error = %v, wantError %t", err, tt.wantError)
			}
			if started != tt.wantStarted {
				t.Errorf("server started = %t, want %t", started, tt.wantStarted)
			}
		})
	}
}

func TestRunDependencyFailures(t *testing.T) {
	poolFailure := productionDependencies()
	poolFailure.newPool = func(context.Context, string) (*pgxpool.Pool, error) {
		return nil, errors.New("pool unavailable")
	}
	if err := runWithDependencies(lookup(validEnvironment()), func(*http.Server) error {
		t.Fatal("server started")

		return nil
	}, poolFailure); err == nil {
		t.Fatal("pool failure succeeded")
	}
	protectionFailure := productionDependencies()
	protectionFailure.newPool = func(context.Context, string) (*pgxpool.Pool, error) {
		return nil, nil
	}
	protectionFailure.closePool = func(*pgxpool.Pool) {}
	protectionFailure.newProtection = func(string) (*security.Service, error) {
		return nil, errors.New("protection unavailable")
	}
	if err := runWithDependencies(lookup(validEnvironment()), func(*http.Server) error {
		t.Fatal("server started")

		return nil
	}, protectionFailure); err == nil {
		t.Fatal("protection failure succeeded")
	}
	serviceFailure := productionDependencies()
	serviceFailure.newPool = func(context.Context, string) (*pgxpool.Pool, error) {
		return nil, nil
	}
	serviceFailure.closePool = func(*pgxpool.Pool) {}
	serviceFailure.newService = func(usecaseauth.Dependencies, string, string) (*usecaseauth.Service, error) {
		return nil, errors.New("service unavailable")
	}
	if err := runWithDependencies(lookup(validEnvironment()), func(*http.Server) error {
		t.Fatal("server started")

		return nil
	}, serviceFailure); err == nil {
		t.Fatal("service failure succeeded")
	}
	successfulDependencies := productionDependencies()
	successfulDependencies.newPool = func(context.Context, string) (*pgxpool.Pool, error) {
		return nil, nil
	}
	closed := false
	successfulDependencies.closePool = func(*pgxpool.Pool) { closed = true }
	if err := runWithDependencies(lookup(validEnvironment()), func(*http.Server) error {
		return http.ErrServerClosed
	}, successfulDependencies); err != nil {
		t.Fatalf("nil pool run error = %v", err)
	}
	if !closed {
		t.Fatal("pool was not closed")
	}
}

func TestStart(t *testing.T) {
	originalLookup := lookupEnvironment
	originalListen := listenAndServe
	originalLogWriter := serverLogWriter
	originalExit := exitProcess
	t.Cleanup(func() {
		lookupEnvironment = originalLookup
		listenAndServe = originalListen
		serverLogWriter = originalLogWriter
		exitProcess = originalExit
	})
	tests := []struct {
		name     string
		lookup   LookupEnv
		invoke   func()
		wantLog  bool
		wantExit bool
	}{
		{
			name:   "start with valid configuration",
			lookup: lookup(validEnvironment()),
			invoke: start,
		},
		{
			name: "main with invalid configuration",
			lookup: func(string) (string, bool) {
				return "", false
			},
			invoke:   main,
			wantLog:  true,
			wantExit: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			exited := false
			lookupEnvironment = tt.lookup
			listenAndServe = func(*http.Server) error { return http.ErrServerClosed }
			serverLogWriter = &output
			exitProcess = func(int) { exited = true }
			tt.invoke()
			if logged := output.Len() > 0; logged != tt.wantLog {
				t.Errorf("logged = %t, want %t", logged, tt.wantLog)
			}
			if exited != tt.wantExit {
				t.Errorf("exited = %t, want %t", exited, tt.wantExit)
			}
		})
	}
}
