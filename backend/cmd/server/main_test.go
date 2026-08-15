package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name        string
		lookup      LookupEnv
		serveResult error
		wantError   bool
		wantStarted bool
	}{
		{name: "invalid configuration", lookup: func(string) (string, bool) { return "", false }, wantError: true},
		{name: "successful serve", lookup: lookup(validEnvironment()), wantStarted: true},
		{name: "closed server", lookup: lookup(validEnvironment()), serveResult: http.ErrServerClosed, wantStarted: true},
		{name: "serve failure", lookup: lookup(validEnvironment()), serveResult: errors.New("serve failed"), wantError: true, wantStarted: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			started := false
			err := run(tt.lookup, func(server *http.Server) error {
				started = true
				response := httptest.NewRecorder()
				server.Handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health", nil))
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

func TestStart(t *testing.T) {
	originalLookup := lookupEnvironment
	originalListen := listenAndServe
	originalPrint := printError
	originalExit := exitProcess
	t.Cleanup(func() {
		lookupEnvironment = originalLookup
		listenAndServe = originalListen
		printError = originalPrint
		exitProcess = originalExit
	})

	tests := []struct {
		name     string
		lookup   LookupEnv
		invoke   func()
		wantLog  bool
		wantExit bool
	}{
		{name: "start with valid configuration", lookup: lookup(validEnvironment()), invoke: start},
		{name: "main with invalid configuration", lookup: func(string) (string, bool) { return "", false }, invoke: main, wantLog: true, wantExit: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logged := false
			exited := false
			lookupEnvironment = tt.lookup
			listenAndServe = func(*http.Server) error { return http.ErrServerClosed }
			printError = func(...any) { logged = true }
			exitProcess = func(int) { exited = true }
			tt.invoke()
			if logged != tt.wantLog {
				t.Errorf("logged = %t, want %t", logged, tt.wantLog)
			}
			if exited != tt.wantExit {
				t.Errorf("exited = %t, want %t", exited, tt.wantExit)
			}
		})
	}
}
