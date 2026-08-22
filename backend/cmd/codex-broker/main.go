package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/yukihito-jokyu/topic2html/backend/repository/codex"
)

const (
	brokerEndpointEnv = "TOPIC2HTML_CODEX_EXECUTION_BROKER_ENDPOINT"
	appServerEnv      = "TOPIC2HTML_CODEX_APP_SERVER_EXECUTABLE"
	workdirEnv        = "TOPIC2HTML_CODEX_APP_SERVER_WORKDIR"
)

var (
	brokerLookup       = os.LookupEnv
	brokerExit         = os.Exit
	brokerRun          = run
	cleanupGrace       = 5 * time.Second
	newCommand         = exec.Command
	removePath         = os.Remove
	chmodPath          = os.Chmod
	shutdownContext    = signal.NotifyContext
	killProcessGroup   = syscall.Kill
	cleanupTimeout     = time.After
	serverBuildVersion = "dev"
	socketOwnerID      = func(info os.FileInfo) uint32 { return info.Sys().(*syscall.Stat_t).Uid }
	brokerUserID       = os.Geteuid
)

type config struct{ endpoint, executable, workdir string }

func loadConfig(lookup func(string) (string, bool)) (config, error) {
	if lookup == nil {
		return config{}, errors.New("lookup is required")
	}
	endpoint, ok := lookup(brokerEndpointEnv)
	if !ok {
		return config{}, errors.New("endpoint is required")
	}
	u, err := url.Parse(endpoint)
	if err != nil || u.Scheme != "unix" || u.Host != "" || u.RawQuery != "" || u.Fragment != "" || !filepath.IsAbs(u.Path) {
		return config{}, errors.New("endpoint must be an absolute unix socket")
	}
	executable, ok := lookup(appServerEnv)
	if !ok || !filepath.IsAbs(executable) {
		return config{}, errors.New("executable must be absolute")
	}
	info, err := os.Stat(executable)
	if err != nil || info.IsDir() || info.Mode()&0111 == 0 {
		return config{}, errors.New("executable is not runnable")
	}
	workdir, ok := lookup(workdirEnv)
	if !ok || !filepath.IsAbs(workdir) {
		return config{}, errors.New("workdir must be absolute")
	}
	entries, err := os.ReadDir(workdir)
	if err != nil || len(entries) != 0 {
		return config{}, errors.New("workdir must be an empty directory")
	}
	workdirInfo, _ := os.Stat(workdir)
	if workdirInfo.Mode().Perm()&0077 != 0 {
		return config{}, errors.New("workdir permissions must be broker-private")
	}

	return config{endpoint: u.Path, executable: executable, workdir: workdir}, nil
}

func listen(path string) (net.Listener, error) {
	info, err := os.Lstat(path)
	if err == nil {
		if info.Mode()&os.ModeSocket == 0 {
			return nil, errors.New("refusing to replace a non-socket endpoint")
		}
		if err := removePath(path); err != nil {
			return nil, err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	listener, err := (&net.ListenConfig{}).Listen(context.Background(), "unix", path)
	if err != nil {
		return nil, err
	}
	if err := chmodPath(path, 0660); err != nil {
		_ = listener.Close()

		return nil, err
	}

	return listener, nil
}

func main() {
	if err := brokerRun(brokerLookup, listen, serve); err != nil {
		fmt.Fprintln(os.Stderr, "codex broker startup failed")
		brokerExit(1)
	}
}

func run(lookup func(string) (string, bool), listenSocket func(string) (net.Listener, error), serve func(net.Listener, *codex.Broker) error) error {
	configuration, err := loadConfig(lookup)
	if err != nil {
		return err
	}
	listener, err := listenSocket(configuration.endpoint)
	if err != nil {
		return err
	}
	defer listener.Close()
	broker, _ := codex.NewBroker(func() (codex.Runner, error) { return newRunner(configuration), nil })
	defer broker.Close()
	shutdown, stop := shutdownContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	done := make(chan struct{})
	go func() {
		<-shutdown.Done()
		broker.Close()
		_ = listener.Close()
		close(done)
	}()
	err = serve(listener, broker)
	if shutdown.Err() != nil {
		<-done

		return nil
	}

	return err
}

type ipcRequest struct {
	Prompt string `json:"prompt"`
}

func serve(listener net.Listener, broker *codex.Broker) error {
	if listener == nil || broker == nil {
		return errors.New("listener and broker are required")
	}
	if err := secureSocket(listener); err != nil {
		return err
	}
	for {
		connection, err := listener.Accept()
		if err != nil {
			return err
		}
		go serveConnection(connection, broker)
	}
}

func secureSocket(listener net.Listener) error {
	address, ok := listener.Addr().(*net.UnixAddr)
	if !ok || address.Name == "" {
		return errors.New("broker listener must be a unix socket")
	}
	info, err := os.Stat(address.Name)
	if err != nil {
		return errors.New("broker socket permissions are unsafe")
	}
	ownerID := socketOwnerID(info)
	if info.Mode()&os.ModeSocket == 0 || info.Mode().Perm() != 0660 || int(ownerID) != brokerUserID() {
		return errors.New("broker socket permissions are unsafe")
	}

	return nil
}

func serveConnection(connection net.Conn, broker *codex.Broker) {
	defer connection.Close()
	var request ipcRequest
	if err := json.NewDecoder(io.LimitReader(connection, 1<<20)).Decode(&request); err != nil || strings.TrimSpace(request.Prompt) == "" || len(request.Prompt) > 1<<20 {
		_ = json.NewEncoder(connection).Encode(codex.Result{Unavailable: true})

		return
	}
	_ = json.NewEncoder(connection).Encode(broker.Execute(context.Background(), request.Prompt))
}

type runner struct {
	configuration config
	mu            sync.Mutex
	command       *exec.Cmd
	stdin         io.WriteCloser
	stdout        io.ReadCloser
	cancel        context.CancelFunc
	closed        bool
}

func newRunner(configuration config) *runner { return &runner{configuration: configuration} }

func (r *runner) Run(ctx context.Context, prompt string) codex.Result {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	command := newCommand(r.configuration.executable, "app-server", "--stdio")
	command.Dir = r.configuration.workdir
	command.Env = []string{}
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	stdin, err := command.StdinPipe()
	if err != nil {
		return codex.Result{Unavailable: true}
	}
	stdout, err := command.StdoutPipe()
	if err != nil {
		_ = stdin.Close()

		return codex.Result{Unavailable: true}
	}
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		_ = stdin.Close()
		_ = stdout.Close()

		return codex.Result{Unavailable: true}
	}
	r.command, r.stdin, r.stdout, r.cancel = command, stdin, stdout, cancel
	r.mu.Unlock()
	if err := command.Start(); err != nil {
		r.clear()

		return codex.Result{Unavailable: true}
	}
	result := codex.RunWire(ctx, stdin, io.LimitReader(stdout, 1<<20), serverBuildVersion, r.configuration.workdir, prompt)
	if r.waitCleanup(command) {
		result = codex.Result{Unavailable: true}
	}
	r.clear()

	return result
}

func (r *runner) Close() codex.Result {
	r.mu.Lock()
	r.closed = true
	stdin, stdout, cancel := r.stdin, r.stdout, r.cancel
	r.mu.Unlock()
	if stdin != nil {
		_ = stdin.Close()
	}
	if stdout != nil {
		_ = stdout.Close()
	}
	if cancel != nil {
		cancel()
	}

	return codex.Result{Unavailable: true}
}

func (r *runner) waitCleanup(command *exec.Cmd) bool {
	done := make(chan error, 1)
	go func() { done <- command.Wait() }()
	select {
	case <-done:
		return false
	case <-cleanupTimeout(cleanupGrace):
	}
	_ = killProcessGroup(-command.Process.Pid, syscall.SIGTERM)
	select {
	case <-done:
		return true
	case <-cleanupTimeout(cleanupGrace):
	}
	_ = killProcessGroup(-command.Process.Pid, syscall.SIGKILL)
	<-done

	return true
}

func (r *runner) clear() {
	r.mu.Lock()
	r.command, r.stdin, r.stdout, r.cancel = nil, nil, nil, nil
	r.mu.Unlock()
}

func environment(values map[string]string) func(string) (string, bool) {
	return func(name string) (string, bool) {
		value, ok := values[name]

		return strings.TrimSpace(value), ok
	}
}
