package codex

import (
	"context"
	"errors"
	"sync"
	"testing"
)

type fakeRunner struct {
	result  Result
	started chan struct{}
	finish  chan struct{}
	closed  int
	mu      sync.Mutex
}

type holdingRunner struct {
	started chan struct{}
	finish  chan struct{}
}

func (r *holdingRunner) Run(_ context.Context, _ string) Result {
	close(r.started)
	<-r.finish

	return Result{HTML: "<html>"}
}

func (*holdingRunner) Close() Result { return Result{Unavailable: true} }

func (r *fakeRunner) Run(_ context.Context, _ string) Result {
	if r.started != nil {
		close(r.started)
	}
	if r.finish != nil {
		<-r.finish
	}

	return r.result
}

func (r *fakeRunner) Close() Result {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.closed++
	if r.finish != nil {
		select {
		case <-r.finish:
		default:
			close(r.finish)
		}
	}

	return Result{Unavailable: true}
}

func TestBroker(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{
			name: "nil factory is rejected",
			run: func(t *testing.T) {
				if _, err := NewBroker(nil); err == nil {
					t.Fatal("NewBroker(nil) succeeded")
				}
			},
		},
		{
			name: "returns unavailable when runner creation fails",
			run: func(t *testing.T) {
				broker, err := NewBroker(func() (Runner, error) { return nil, errors.New("unavailable") })
				if err != nil {
					t.Fatal(err)
				}
				if got := broker.Execute(context.Background(), "prompt"); !got.Unavailable {
					t.Fatalf("result=%+v", got)
				}
			},
		},
		{
			name: "rejects after close without starting a runner",
			run:  testBrokerRejectsAfterClose,
		},
		{
			name: "closes an admitted runner once",
			run:  testBrokerClosesAdmittedRunner,
		},
		{
			name: "waits until each admitted runner has left the registry",
			run:  testBrokerWaitsForRunner,
		},
		{
			name: "normalizes unsafe results",
			run: func(t *testing.T) {
				broker, err := NewBroker(func() (Runner, error) {
					return &fakeRunner{
						result: Result{},
					}, nil
				})
				if err != nil {
					t.Fatal(err)
				}
				if got := broker.Execute(context.Background(), "prompt"); !got.Unavailable {
					t.Fatalf("result=%+v", got)
				}
			},
		},
		{
			name: "preserves a shutdown rejection",
			run: func(t *testing.T) {
				if got := normalize(Result{ShutdownRejected: true}); !got.ShutdownRejected {
					t.Fatalf("result=%+v", got)
				}
			},
		},
		{
			name: "preserves a shutdown interruption",
			run: func(t *testing.T) {
				if got := normalize(Result{ShutdownInterrupted: true}); !got.ShutdownInterrupted {
					t.Fatalf("result=%+v", got)
				}
			},
		},
		{
			name: "close is idempotent",
			run: func(t *testing.T) {
				broker, err := NewBroker(func() (Runner, error) { return &fakeRunner{}, nil })
				if err != nil {
					t.Fatal(err)
				}
				broker.Close()
				broker.Close()
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, testCase.run)
	}
}

func testBrokerRejectsAfterClose(t *testing.T) {
	started := false
	broker, err := NewBroker(func() (Runner, error) {
		started = true

		return &fakeRunner{}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	broker.Close()
	if got := broker.Execute(context.Background(), "prompt"); !got.ShutdownRejected || started {
		t.Fatalf("result=%+v started=%t", got, started)
	}
}

func testBrokerClosesAdmittedRunner(t *testing.T) {
	runner := &fakeRunner{
		result: Result{
			HTML: "<html>",
		},
		started: make(chan struct{}),
		finish:  make(chan struct{}),
	}
	broker, err := NewBroker(func() (Runner, error) { return runner, nil })
	if err != nil {
		t.Fatal(err)
	}
	result := make(chan Result, 1)
	go func() { result <- broker.Execute(context.Background(), "prompt") }()
	<-runner.started
	broker.Close()
	got := <-result
	runner.mu.Lock()
	closed := runner.closed
	runner.mu.Unlock()
	if !got.ShutdownInterrupted || closed != 1 {
		t.Fatalf("result=%+v closed=%d", got, closed)
	}
}

func testBrokerWaitsForRunner(t *testing.T) {
	runner := &holdingRunner{
		started: make(chan struct{}),
		finish:  make(chan struct{}),
	}
	broker, err := NewBroker(func() (Runner, error) { return runner, nil })
	if err != nil {
		t.Fatal(err)
	}
	result := make(chan Result, 1)
	go func() { result <- broker.Execute(context.Background(), "prompt") }()
	<-runner.started
	closed := make(chan struct{})
	go func() {
		broker.Close()
		close(closed)
	}()
	select {
	case <-closed:
		t.Fatal("Close returned before the running attempt completed")
	default:
	}
	close(runner.finish)
	<-result
	<-closed
}
