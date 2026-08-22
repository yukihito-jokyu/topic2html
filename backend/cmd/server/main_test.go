//nolint:noctx // Unix listenerの境界を検証する。
package main

import (
	"bytes"
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yukihito-jokyu/topic2html/backend/repository/security"
	usecaseauth "github.com/yukihito-jokyu/topic2html/backend/usecase/auth"
)

var chmodBrokerSocket = os.Chmod

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
			dependencies := productionDependencies()
			dependencies.verifyBroker = func(context.Context, string) error { return nil }
			err := runWithDependencies(tt.lookup, func(server *http.Server) error {
				started = true
				response := httptest.NewRecorder()
				server.Handler.ServeHTTP(response, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/health", nil))
				if response.Code != http.StatusNoContent {
					t.Errorf("health status = %d, want %d", response.Code, http.StatusNoContent)
				}

				return tt.serveResult
			}, dependencies)
			if (err != nil) != tt.wantError {
				t.Fatalf("run() error = %v, wantError %t", err, tt.wantError)
			}
			if started != tt.wantStarted {
				t.Errorf("server started = %t, want %t", started, tt.wantStarted)
			}
		})
	}
}

func TestLoadDotEnv(t *testing.T) {
	missing := errors.New("missing environment file")
	loadFile := func(results map[string]error) func(string) error {
		return func(filename string) error { return results[filename] }
	}
	tests := []struct {
		name    string
		results map[string]error
		wantErr bool
	}{
		{
			name: "root environment file",
			results: map[string]error{
				"../.env": nil,
			},
		},
		{
			name: "working directory environment file",
			results: map[string]error{
				"../.env": os.ErrNotExist,
				".env":    nil,
			},
		},
		{
			name: "no environment file",
			results: map[string]error{
				"../.env": os.ErrNotExist,
				".env":    os.ErrNotExist,
			},
		},
		{
			name: "unreadable environment file",
			results: map[string]error{
				"../.env": missing,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := loadDotEnv(loadFile(tt.results))
			if (err != nil) != tt.wantErr {
				t.Fatalf("loadDotEnv() error = %v, wantErr %t", err, tt.wantErr)
			}
		})
	}
}

func TestStartWithEnvironmentLoad(t *testing.T) {
	originalLoadEnvironment := loadEnvironment
	originalRunServer := runServer
	originalExitProcess := exitProcess
	originalLogWriter := serverLogWriter
	t.Cleanup(func() {
		loadEnvironment = originalLoadEnvironment
		runServer = originalRunServer
		exitProcess = originalExitProcess
		serverLogWriter = originalLogWriter
	})

	tests := []struct {
		name               string
		loadEnvironmentErr error
		runErr             error
		wantExit           bool
	}{
		{
			name:               "environment file error",
			loadEnvironmentErr: errors.New("environment file error"),
			wantExit:           true,
		},
		{
			name:     "server error",
			runErr:   errors.New("server error"),
			wantExit: true,
		},
		{
			name: "server stops cleanly",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var log bytes.Buffer
			exitCode := -1
			loadEnvironment = func() error { return tt.loadEnvironmentErr }
			runServer = func(LookupEnv, func(*http.Server) error) error { return tt.runErr }
			exitProcess = func(code int) { exitCode = code }
			serverLogWriter = &log

			start()

			if tt.wantExit && exitCode != 1 {
				t.Errorf("exit code = %d, want 1", exitCode)
			}
			if !tt.wantExit && exitCode != -1 {
				t.Errorf("exit code = %d, want no exit", exitCode)
			}
		})
	}
}

func TestRunDependencyFailures(t *testing.T) {
	for _, testCase := range []struct {
		name      string
		configure func(*dependencies, *bool)
		serve     error
		wantError bool
		wantClose bool
	}{
		{
			name: "broker is unavailable",
			configure: func(dependencies *dependencies, _ *bool) {
				dependencies.verifyBroker = func(context.Context, string) error { return errors.New("broker unavailable") }
			},
			wantError: true,
		},
		{
			name: "pool initialization fails",
			configure: func(dependencies *dependencies, _ *bool) {
				dependencies.newPool = func(context.Context, string) (*pgxpool.Pool, error) { return nil, errors.New("pool unavailable") }
			},
			wantError: true,
		},
		{
			name: "protection initialization fails",
			configure: func(dependencies *dependencies, _ *bool) {
				dependencies.newPool = func(context.Context, string) (*pgxpool.Pool, error) { return nil, nil }
				dependencies.closePool = func(*pgxpool.Pool) {}
				dependencies.newProtection = func(string) (*security.Service, error) { return nil, errors.New("protection unavailable") }
			},
			wantError: true,
		},
		{
			name: "service initialization fails",
			configure: func(dependencies *dependencies, _ *bool) {
				dependencies.newPool = func(context.Context, string) (*pgxpool.Pool, error) { return nil, nil }
				dependencies.closePool = func(*pgxpool.Pool) {}
				dependencies.newService = func(usecaseauth.Dependencies, string, string) (*usecaseauth.Service, error) {
					return nil, errors.New("service unavailable")
				}
			},
			wantError: true,
		},
		{
			name: "server closes cleanly",
			configure: func(dependencies *dependencies, closed *bool) {
				dependencies.newPool = func(context.Context, string) (*pgxpool.Pool, error) { return nil, nil }
				dependencies.closePool = func(*pgxpool.Pool) { *closed = true }
			},
			serve:     http.ErrServerClosed,
			wantClose: true,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			dependencies := productionDependencies()
			dependencies.verifyBroker = func(context.Context, string) error { return nil }
			closed := false
			testCase.configure(&dependencies, &closed)
			err := runWithDependencies(lookup(validEnvironment()), func(*http.Server) error { return testCase.serve }, dependencies)
			if (err != nil) != testCase.wantError || closed != testCase.wantClose {
				t.Fatalf("error=%v closed=%t", err, closed)
			}
		})
	}
}

func TestVerifyBrokerEndpoint(t *testing.T) {
	originalOwnerID, originalUserID := socketOwnerID, serverUserID
	originalDialBroker := dialBroker
	t.Cleanup(func() { socketOwnerID, serverUserID, dialBroker = originalOwnerID, originalUserID, originalDialBroker })
	serverUserID = func() int { return 1000 }
	ownerID := uint32(1001)
	socketOwnerID = func(os.FileInfo) uint32 { return ownerID }

	tests := []struct {
		name      string
		setup     func(*testing.T) string
		ownerID   uint32
		dialError error
		wantError bool
	}{
		{
			name: "allowed",
			setup: func(t *testing.T) string {
				path := brokerSocketPath(t)
				listener, err := net.Listen("unix", path)
				if err != nil {
					t.Fatal(err)
				}
				if err := chmodBrokerSocket(path, 0660); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = listener.Close() })
				go func() {
					connection, acceptErr := listener.Accept()
					if acceptErr == nil {
						_ = connection.Close()
					}
				}()

				return "unix://" + path
			},
		},
		{
			name: "missing",
			setup: func(t *testing.T) string {
				return "unix:///missing/broker.sock"
			},
			wantError: true,
		},
		{
			name: "malformed-endpoint",
			setup: func(t *testing.T) string {
				return "tcp://127.0.0.1:1"
			},
			wantError: true,
		},
		{
			name: "dial-failure",
			setup: func(t *testing.T) string {
				path := brokerSocketPath(t)
				listener, err := net.Listen("unix", path)
				if err != nil {
					t.Fatal(err)
				}
				if err := chmodBrokerSocket(path, 0660); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = listener.Close() })

				return "unix://" + path
			},
			dialError: errors.New("dial failed"),
			wantError: true,
		},
		{
			name: "same-owner",
			setup: func(t *testing.T) string {
				path := brokerSocketPath(t)
				listener, err := net.Listen("unix", path)
				if err != nil {
					t.Fatal(err)
				}
				if err := chmodBrokerSocket(path, 0660); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = listener.Close() })

				return "unix://" + path
			},
			ownerID:   1000,
			wantError: true,
		},
		{
			name: "unsafe-mode",
			setup: func(t *testing.T) string {
				path := brokerSocketPath(t)
				listener, err := net.Listen("unix", path)
				if err != nil {
					t.Fatal(err)
				}
				if err := chmodBrokerSocket(path, 0666); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = listener.Close() })

				return "unix://" + path
			},
			wantError: true,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			dialBroker = originalDialBroker
			if testCase.dialError != nil {
				dialBroker = func(context.Context, string, string) (net.Conn, error) { return nil, testCase.dialError }
			}
			ownerID = testCase.ownerID
			if ownerID == 0 {
				ownerID = 1001
			}
			err := verifyBrokerEndpoint(context.Background(), testCase.setup(t))
			if (err != nil) != testCase.wantError {
				t.Fatalf("verifyBrokerEndpoint() error = %v, wantError %t", err, testCase.wantError)
			}
		})
	}
}

func brokerSocketPath(t *testing.T) string {
	t.Helper()
	directory, err := os.MkdirTemp("", "t2h-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(directory) })

	return filepath.Join(directory, "b.sock")
}

func TestSocketOwnerID(t *testing.T) {
	info, err := os.Stat(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if got := socketOwnerID(info); int(got) != os.Geteuid() {
		t.Fatalf("socketOwnerID() = %d, want %d", got, os.Geteuid())
	}
}

func TestStart(t *testing.T) {
	originalLookup := lookupEnvironment
	originalListen := listenAndServe
	originalLogWriter := serverLogWriter
	originalExit := exitProcess
	originalRunServer := runServer
	t.Cleanup(func() {
		lookupEnvironment = originalLookup
		listenAndServe = originalListen
		serverLogWriter = originalLogWriter
		exitProcess = originalExit
		runServer = originalRunServer
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
			if tt.wantLog {
				runServer = originalRunServer
			} else {
				runServer = func(LookupEnv, func(*http.Server) error) error { return nil }
			}
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
