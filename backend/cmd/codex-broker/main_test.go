//nolint:gosec,noctx // 実行可能fixtureとローカルソケットを使う。
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/yukihito-jokyu/topic2html/backend/repository/codex"
)

type testCloser struct{ closed bool }

func (c *testCloser) Write(value []byte) (int, error) { return len(value), nil }
func (*testCloser) Read([]byte) (int, error)          { return 0, io.EOF }
func (c *testCloser) Close() error {
	c.closed = true

	return nil
}

func brokerSocketPath(t *testing.T, name string) string {
	t.Helper()
	directory, err := os.MkdirTemp("", "t2h-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(directory) })

	return filepath.Join(directory, name)
}

func TestLoadConfig(t *testing.T) {
	if _, err := loadConfig(nil); err == nil {
		t.Fatal("loadConfig(nil) succeeded")
	}
	workdir := t.TempDir()
	if err := os.Chmod(workdir, 0700); err != nil {
		t.Fatal(err)
	}
	values := map[string]string{
		brokerEndpointEnv: "unix://" + filepath.Join(t.TempDir(), "broker.sock"),
		appServerEnv:      "/bin/sh",
		workdirEnv:        workdir,
	}
	if _, err := loadConfig(environment(values)); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{brokerEndpointEnv, appServerEnv, workdirEnv} {
		t.Run("requires "+key, func(t *testing.T) {
			copy := map[string]string{}
			for name, value := range values {
				copy[name] = value
			}
			delete(copy, key)
			if _, err := loadConfig(environment(copy)); err == nil {
				t.Fatal("loadConfig() succeeded")
			}
		})
	}
	t.Run("rejects unsafe values", func(t *testing.T) {
		for _, mutate := range []func(map[string]string){
			func(v map[string]string) { v[brokerEndpointEnv] = "tcp://127.0.0.1:1" },
			func(v map[string]string) { v[brokerEndpointEnv] += "?unsafe=true" },
			func(v map[string]string) { v[appServerEnv] = "sh" },
			func(v map[string]string) { v[workdirEnv] = "/missing" },
			func(v map[string]string) { v[appServerEnv] = t.TempDir() },
		} {
			copy := map[string]string{}
			for name, value := range values {
				copy[name] = value
			}
			mutate(copy)
			if _, err := loadConfig(environment(copy)); err == nil {
				t.Fatal("loadConfig() succeeded")
			}
		}
		if err := os.WriteFile(filepath.Join(workdir, "not-empty"), []byte("x"), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := loadConfig(environment(values)); err == nil {
			t.Fatal("loadConfig() succeeded")
		}
		sharedWorkdir := t.TempDir()
		if err := os.Chmod(sharedWorkdir, 0755); err != nil {
			t.Fatal(err)
		}
		if _, err := loadConfig(environment(map[string]string{
			brokerEndpointEnv: values[brokerEndpointEnv],
			appServerEnv:      values[appServerEnv],
			workdirEnv:        sharedWorkdir,
		})); err == nil {
			t.Fatal("loadConfig() accepted a shared workdir")
		}
	})
}

func TestListen(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{
			name: "ソケットを0660で作成する",
			run: func(t *testing.T) {
				path := brokerSocketPath(t, "broker.sock")
				listener, err := listen(path)
				if err != nil {
					t.Fatal(err)
				}
				info, err := os.Stat(path)
				if err != nil || info.Mode().Perm() != 0660 {
					t.Fatalf("mode=%v err=%v", info.Mode(), err)
				}
				if err := listener.Close(); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "通常ファイルを置換しない",
			run: func(t *testing.T) {
				path := filepath.Join(t.TempDir(), "broker.sock")
				if err := os.WriteFile(path, []byte("regular"), 0600); err != nil {
					t.Fatal(err)
				}
				if _, err := listen(path); err == nil {
					t.Fatal("listen() replaced a regular file")
				}
			},
		},
		{
			name: "存在しない親ディレクトリを拒否する",
			run: func(t *testing.T) {
				if _, err := listen(filepath.Join(t.TempDir(), "missing", "broker.sock")); err == nil {
					t.Fatal("listen() accepted a missing parent")
				}
			},
		},
		{
			name: "既存ソケットを置換する",
			run: func(t *testing.T) {
				socket := brokerSocketPath(t, "existing.sock")
				existing, err := net.Listen("unix", socket)
				if err != nil {
					t.Fatal(err)
				}
				replaced, err := listen(socket)
				if err != nil {
					t.Fatal(err)
				}
				_ = existing.Close()
				_ = replaced.Close()
			},
		},
		{
			name: "シンボリックリンクを拒否する",
			run: func(t *testing.T) {
				blocked := filepath.Join(t.TempDir(), "blocked.sock")
				if err := os.Symlink(filepath.Join(t.TempDir(), "target"), blocked); err != nil {
					t.Fatal(err)
				}
				if _, err := listen(blocked); err == nil {
					t.Fatal("listen() accepted an unsafe endpoint")
				}
			},
		},
		{
			name: "既存ソケットの削除失敗を返す",
			run: func(t *testing.T) {
				originalRemove := removePath
				t.Cleanup(func() { removePath = originalRemove })
				removePath = func(string) error { return errors.New("remove") }
				removeTarget := brokerSocketPath(t, "remove.sock")
				removeListener, err := net.Listen("unix", removeTarget)
				if err != nil {
					t.Fatal(err)
				}
				defer removeListener.Close()
				if _, err := listen(removeTarget); err == nil {
					t.Fatal("listen() accepted removal failure")
				}
			},
		},
		{
			name: "chmod失敗を返す",
			run: func(t *testing.T) {
				originalChmod := chmodPath
				t.Cleanup(func() { chmodPath = originalChmod })
				chmodPath = func(string, os.FileMode) error { return errors.New("chmod") }
				if _, err := listen(brokerSocketPath(t, "chmod.sock")); err == nil {
					t.Fatal("listen() accepted chmod failure")
				}
			},
		},
		{
			name: "利用できない親を拒否する",
			run: func(t *testing.T) {
				if _, err := listen("/dev/null/broker.sock"); err == nil {
					t.Fatal("listen() accepted an invalid parent")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestServe(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{
			name: "nilを拒否する",
			run: func(t *testing.T) {
				if err := serve(nil, nil); err == nil {
					t.Fatal("serve() accepted nil values")
				}
			},
		},
		{
			name: "TCP listenerを拒否する",
			run: func(t *testing.T) {
				tcp, err := net.Listen("tcp", "127.0.0.1:0")
				if err != nil {
					t.Fatal(err)
				}
				defer tcp.Close()
				broker, err := codex.NewBroker(func() (codex.Runner, error) {
					return &testRunner{}, nil
				})
				if err != nil {
					t.Fatal(err)
				}
				if err := serve(tcp, broker); err == nil {
					t.Fatal("serve() accepted TCP")
				}
			},
		},
		{
			name: "安全でないmodeを拒否する",
			run: func(t *testing.T) {
				path := brokerSocketPath(t, "broker.sock")
				listener, err := listen(path)
				if err != nil {
					t.Fatal(err)
				}
				defer listener.Close()
				if err := os.Chmod(path, 0600); err != nil {
					t.Fatal(err)
				}
				if err := secureSocket(listener); err == nil {
					t.Fatal("secureSocket() accepted unsafe mode")
				}
			},
		},
		{
			name: "別所有者を拒否する",
			run: func(t *testing.T) {
				originalBrokerUserID := brokerUserID
				t.Cleanup(func() { brokerUserID = originalBrokerUserID })
				path := brokerSocketPath(t, "broker.sock")
				listener, err := listen(path)
				if err != nil {
					t.Fatal(err)
				}
				defer listener.Close()
				if err := os.Chmod(path, 0660); err != nil {
					t.Fatal(err)
				}
				brokerUserID = func() int { return os.Geteuid() + 1 }
				if err := secureSocket(listener); err == nil {
					t.Fatal("secureSocket() accepted a different owner")
				}
			},
		},
		{
			name: "消えたsocketを拒否する",
			run: func(t *testing.T) {
				path := brokerSocketPath(t, "broker.sock")
				listener, err := listen(path)
				if err != nil {
					t.Fatal(err)
				}
				defer listener.Close()
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
				if err := secureSocket(listener); err == nil {
					t.Fatal("secureSocket() accepted a missing socket")
				}
			},
		},
		{
			name: "有効な要求を返す",
			run: func(t *testing.T) {
				path := brokerSocketPath(t, "broker.sock")
				listener, err := listen(path)
				if err != nil {
					t.Fatal(err)
				}
				broker, err := codex.NewBroker(func() (codex.Runner, error) {
					return &testRunner{}, nil
				})
				if err != nil {
					t.Fatal(err)
				}
				result := make(chan error, 1)
				go func() { result <- serve(listener, broker) }()
				connection, err := net.Dial("unix", path)
				if err != nil {
					t.Fatal(err)
				}
				if err := json.NewEncoder(connection).Encode(ipcRequest{Prompt: "prompt"}); err != nil {
					t.Fatal(err)
				}
				var response codex.Result
				if err := json.NewDecoder(connection).Decode(&response); err != nil || response.HTML != "<html></html>" {
					t.Fatalf("response=%+v err=%v", response, err)
				}
				_ = connection.Close()
				if err := listener.Close(); err != nil {
					t.Fatal(err)
				}
				if err := <-result; err == nil {
					t.Fatal("serve() did not return listener close error")
				}
			},
		},
		{
			name: "空のpromptを安全に拒否する",
			run: func(t *testing.T) {
				broker, err := codex.NewBroker(func() (codex.Runner, error) {
					return &testRunner{}, nil
				})
				if err != nil {
					t.Fatal(err)
				}
				left, right := net.Pipe()
				go serveConnection(left, broker)
				if _, err := right.Write([]byte(`{"prompt":""}`)); err != nil {
					t.Fatal(err)
				}
				var response codex.Result
				if err := json.NewDecoder(right).Decode(&response); err != nil || !response.Unavailable {
					t.Fatalf("response=%+v err=%v", response, err)
				}
				_ = right.Close()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

type testRunner struct{}

func (*testRunner) Run(context.Context, string) codex.Result {
	return codex.Result{HTML: "<html></html>"}
}

func (*testRunner) Close() codex.Result { return codex.Result{Unavailable: true} }

func runnerFixture(t *testing.T) config {
	t.Helper()
	workdir := t.TempDir()
	script := filepath.Join(t.TempDir(), "app-server")
	contents := "#!/bin/sh\nwhile IFS= read -r line; do\ncase \"$line\" in\n*'\"id\":1'*) echo '{\"id\":1,\"result\":{}}' ;;\n*'\"id\":2'*) echo '{\"id\":2,\"result\":{\"thread\":{\"id\":\"thread\"}}}' ;;\n*'\"id\":3'*) echo '{\"id\":3,\"result\":{\"turn\":{\"id\":\"turn\"}}}'; echo '{\"method\":\"item/started\",\"params\":{\"threadId\":\"thread\",\"turnId\":\"turn\",\"item\":{\"id\":\"item\",\"type\":\"agentMessage\"}}}'; echo '{\"method\":\"item/completed\",\"params\":{\"threadId\":\"thread\",\"turnId\":\"turn\",\"item\":{\"id\":\"item\",\"type\":\"agentMessage\",\"text\":\"<html></html>\"}}}'; echo '{\"method\":\"turn/completed\",\"params\":{\"threadId\":\"thread\",\"turnId\":\"turn\",\"status\":\"completed\"}}'; exit 0 ;;\nesac\ndone\n"
	contents = strings.ReplaceAll(contents, "'{\"", "'{\"jsonrpc\":\"2.0\",\"")
	if err := os.WriteFile(script, []byte(contents), 0700); err != nil {
		t.Fatal(err)
	}

	return config{executable: script, workdir: workdir}
}

func TestRunner(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{
			name: "wire結果を返す",
			run: func(t *testing.T) {
				runner := newRunner(runnerFixture(t))
				if got := runner.Run(context.Background(), "prompt"); got.HTML != "<html></html>" {
					t.Fatalf("result=%+v", got)
				}
			},
		},
		{
			name: "close済みrunnerを実行しない",
			run: func(t *testing.T) {
				runner := newRunner(runnerFixture(t))
				if got := runner.Close(); !got.Unavailable {
					t.Fatalf("close=%+v", got)
				}
				if got := runner.Run(context.Background(), "prompt"); !got.Unavailable {
					t.Fatalf("closed runner result=%+v", got)
				}
			},
		},
		{
			name: "起動失敗を安全に返す",
			run: func(t *testing.T) {
				configuration := runnerFixture(t)
				runner := newRunner(configuration)
				runner.configuration.executable = filepath.Join(configuration.workdir, "missing")
				if got := runner.Run(context.Background(), "prompt"); !got.Unavailable {
					t.Fatalf("result=%+v", got)
				}
			},
		},
		{
			name: "pipe作成失敗を安全に返す",
			run: func(t *testing.T) {
				originalCommand := newCommand
				t.Cleanup(func() { newCommand = originalCommand })
				cases := []func() *exec.Cmd{
					func() *exec.Cmd {
						command := exec.Command("/bin/sh")
						command.Stdin = bytes.NewBuffer(nil)

						return command
					},
					func() *exec.Cmd {
						command := exec.Command("/bin/sh")
						command.Stdout = bytes.NewBuffer(nil)

						return command
					},
				}
				for _, command := range cases {
					newCommand = func(string, ...string) *exec.Cmd { return command() }
					if got := newRunner(config{}).Run(context.Background(), "prompt"); !got.Unavailable {
						t.Fatalf("result=%+v", got)
					}
				}
			},
		},
		{
			name: "停止時に稼働processを終了する",
			run: func(t *testing.T) {
				configuration := runnerFixture(t)
				blocking := filepath.Join(t.TempDir(), "blocking-app-server")
				if err := os.WriteFile(blocking, []byte("#!/bin/sh\ntrap '' TERM\nwhile :; do :; done\n"), 0700); err != nil {
					t.Fatal(err)
				}
				originalGrace := cleanupGrace
				cleanupGrace = 20 * time.Millisecond
				t.Cleanup(func() { cleanupGrace = originalGrace })
				active := newRunner(config{executable: blocking, workdir: configuration.workdir})
				done := make(chan codex.Result, 1)
				go func() { done <- active.Run(context.Background(), "prompt") }()
				for deadline := time.Now().Add(time.Second); ; {
					active.mu.Lock()
					started := active.command != nil
					active.mu.Unlock()
					if started {
						break
					}
					if time.Now().After(deadline) {
						t.Fatal("runner did not start")
					}
					time.Sleep(time.Millisecond)
				}
				active.Close()
				if got := <-done; !got.Unavailable {
					t.Fatalf("result=%+v", got)
				}
			},
		},
		{
			name: "closeは入出力を閉じる",
			run: func(t *testing.T) {
				runner := newRunner(runnerFixture(t))
				stdin := &testCloser{}
				stdout := &testCloser{}
				runner.stdin = stdin
				runner.stdout = stdout
				if got := runner.Close(); !got.Unavailable || !stdin.closed || !stdout.closed {
					t.Fatalf("close=%+v stdin=%t stdout=%t", got, stdin.closed, stdout.closed)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestRunnerStartConfiguration(t *testing.T) {
	originalCommand := newCommand
	t.Cleanup(func() { newCommand = originalCommand })
	var executable string
	var arguments []string
	var command *exec.Cmd
	newCommand = func(name string, args ...string) *exec.Cmd {
		executable = name
		arguments = append([]string(nil), args...)
		command = exec.Command("/bin/sh", "-c", "exit 0")

		return command
	}
	runner := newRunner(config{
		executable: "/private/app-server",
		workdir:    "/private/workdir",
	})
	if got := runner.Run(context.Background(), "prompt"); !got.Unavailable {
		t.Fatalf("result=%+v", got)
	}
	if executable != "/private/app-server" || len(arguments) != 2 || arguments[0] != "app-server" || arguments[1] != "--stdio" || command == nil || command.Dir != "/private/workdir" || len(command.Env) != 0 || command.SysProcAttr == nil || !command.SysProcAttr.Setpgid {
		t.Fatalf("executable=%q arguments=%q command=%+v", executable, arguments, command)
	}
}

func TestRun(t *testing.T) {
	workdir := t.TempDir()
	if err := os.Chmod(workdir, 0700); err != nil {
		t.Fatal(err)
	}
	values := map[string]string{
		brokerEndpointEnv: "unix:///unused/broker.sock",
		appServerEnv:      "/bin/sh",
		workdirEnv:        workdir,
	}
	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{
			name: "設定不正を返す",
			run: func(t *testing.T) {
				if err := run(func(string) (string, bool) { return "", false }, func(string) (net.Listener, error) { return nil, nil }, nil); err == nil {
					t.Fatal("run() accepted config failure")
				}
			},
		},
		{
			name: "listener失敗時にserveしない",
			run: func(t *testing.T) {
				called := false
				if err := run(environment(values), func(string) (net.Listener, error) {
					return nil, errors.New("listen")
				}, func(net.Listener, *codex.Broker) error {
					called = true

					return nil
				}); err == nil || called {
					t.Fatal("run() accepted a listener failure")
				}
			},
		},
		{
			name: "serveを実行する",
			run: func(t *testing.T) {
				listener, err := net.Listen("unix", brokerSocketPath(t, "broker.sock"))
				if err != nil {
					t.Fatal(err)
				}
				called := false
				if err := run(environment(values), func(string) (net.Listener, error) {
					return listener, nil
				}, func(net.Listener, *codex.Broker) error {
					called = true

					return nil
				}); err != nil || !called {
					t.Fatalf("run() error=%v called=%t", err, called)
				}
			},
		},
		{
			name: "serveの失敗を返す",
			run: func(t *testing.T) {
				listener, err := net.Listen("unix", brokerSocketPath(t, "broker.sock"))
				if err != nil {
					t.Fatal(err)
				}
				if err := run(environment(values), func(string) (net.Listener, error) {
					return listener, nil
				}, func(net.Listener, *codex.Broker) error {
					return errors.New("serve")
				}); err == nil {
					t.Fatal("run() accepted serve failure")
				}
			},
		},
		{
			name: "brokerのfactory失敗を安全に返す",
			run: func(t *testing.T) {
				listener, err := net.Listen("unix", brokerSocketPath(t, "broker.sock"))
				if err != nil {
					t.Fatal(err)
				}
				if err := run(environment(values), func(string) (net.Listener, error) { return listener, nil }, func(_ net.Listener, broker *codex.Broker) error {
					if got := broker.Execute(context.Background(), "prompt"); !got.Unavailable {
						t.Fatalf("result=%+v", got)
					}

					return nil
				}); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "shutdown後はserveの失敗を返さない",
			run: func(t *testing.T) {
				originalShutdownContext := shutdownContext
				t.Cleanup(func() { shutdownContext = originalShutdownContext })
				listener, err := net.Listen("unix", brokerSocketPath(t, "broker.sock"))
				if err != nil {
					t.Fatal(err)
				}
				shutdown, cancel := context.WithCancel(context.Background())
				cancel()
				shutdownContext = func(context.Context, ...os.Signal) (context.Context, context.CancelFunc) {
					return shutdown, func() {}
				}
				if err := run(environment(values), func(string) (net.Listener, error) { return listener, nil }, func(net.Listener, *codex.Broker) error { return errors.New("closed") }); err != nil {
					t.Fatalf("shutdown error=%v", err)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestMain(t *testing.T) {
	originalLookup, originalExit, originalRun := brokerLookup, brokerExit, brokerRun
	t.Cleanup(func() { brokerLookup, brokerExit, brokerRun = originalLookup, originalExit, originalRun })
	brokerLookup = func(string) (string, bool) { return "", false }
	code := -1
	brokerExit = func(value int) { code = value }
	brokerRun = func(func(string) (string, bool), func(string) (net.Listener, error), func(net.Listener, *codex.Broker) error) error {
		return errors.New("failed")
	}
	main()
	if code != 1 {
		t.Fatalf("exit=%d", code)
	}
	brokerRun = func(func(string) (string, bool), func(string) (net.Listener, error), func(net.Listener, *codex.Broker) error) error {
		return nil
	}
	code = -1
	main()
	if code != -1 {
		t.Fatalf("exit=%d", code)
	}
}

func TestWaitCleanup(t *testing.T) {
	tests := []struct {
		name            string
		command         string
		grace           time.Duration
		expectForced    bool
		expectedSignals []syscall.Signal
	}{
		{
			name:         "通常終了を強制終了と扱わない",
			command:      "exit 0",
			grace:        time.Second,
			expectForced: false,
		},
		{
			name:         "SIGTERMで終了するprocess groupを停止する",
			command:      "sleep 1",
			grace:        time.Millisecond,
			expectForced: true,
		},
		{
			name:         "SIGKILLで終了するprocess groupを停止する",
			command:      "trap '' TERM; while :; do :; done",
			grace:        time.Millisecond,
			expectForced: true,
		},
		{
			name:            "TERMとKILLを順に送る",
			command:         "while :; do :; done",
			grace:           time.Millisecond,
			expectForced:    true,
			expectedSignals: []syscall.Signal{syscall.SIGTERM, syscall.SIGKILL},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			originalGrace := cleanupGrace
			originalKill, originalTimeout := killProcessGroup, cleanupTimeout
			cleanupGrace = testCase.grace
			t.Cleanup(func() {
				cleanupGrace, killProcessGroup, cleanupTimeout = originalGrace, originalKill, originalTimeout
			})
			command := exec.Command("/bin/sh", "-c", testCase.command)
			command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
			if err := command.Start(); err != nil {
				t.Fatal(err)
			}
			kills := []syscall.Signal{}
			if len(testCase.expectedSignals) != 0 {
				cleanupTimeout = func(time.Duration) <-chan time.Time {
					value := make(chan time.Time, 1)
					value <- time.Now()

					return value
				}
				killProcessGroup = func(pid int, signal syscall.Signal) error {
					kills = append(kills, signal)
					if signal == syscall.SIGKILL {
						return originalKill(pid, signal)
					}

					return nil
				}
			}
			if got := newRunner(config{}).waitCleanup(command); got != testCase.expectForced {
				t.Fatalf("forced=%t", got)
			}
			if len(kills) != len(testCase.expectedSignals) {
				t.Fatalf("signals=%v", kills)
			}
			for index, signal := range testCase.expectedSignals {
				if kills[index] != signal {
					t.Fatalf("signals=%v", kills)
				}
			}
		})
	}
}
