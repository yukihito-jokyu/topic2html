// Package codex はapp-server実行Brokerの境界を提供する。
package codex

import (
	"context"
	"errors"
	"sync"
)

const maxIPCMessageBytes = 1 << 20

// ResultはBrokerがServerへ返す結果である。
type Result struct {
	HTML                string `json:"html,omitempty"`
	ShutdownRejected    bool   `json:"shutdown_rejected,omitempty"`
	ShutdownInterrupted bool   `json:"shutdown_interrupted,omitempty"`
	Unavailable         bool   `json:"unavailable,omitempty"`
}

// Runnerは許可済みapp-serverプロセス群を管理する。
type Runner interface {
	Run(context.Context, string) Result
	Close() Result
}

// RunnerFactoryは実行ごとに独立したプロセス群を起動する。
type RunnerFactory func() (Runner, error)

// Brokerは実行受付と停止を直列化する。
type Broker struct {
	mu        sync.Mutex
	drained   *sync.Cond
	closing   bool
	runners   map[Runner]struct{}
	newRunner RunnerFactory
}

// NewBrokerは利用可能なBrokerを生成する。
func NewBroker(factory RunnerFactory) (*Broker, error) {
	if factory == nil {
		return nil, errors.New("runner factory is required")
	}

	broker := &Broker{
		newRunner: factory,
		runners:   make(map[Runner]struct{}),
	}
	broker.drained = sync.NewCond(&broker.mu)

	return broker, nil
}

// Executeは実行を受け付け、停止後は拒否する。
func (b *Broker) Execute(ctx context.Context, prompt string) Result {
	b.mu.Lock()
	if b.closing {
		b.mu.Unlock()

		return Result{ShutdownRejected: true}
	}
	runner, err := b.newRunner()
	if err != nil {
		b.mu.Unlock()

		return Result{Unavailable: true}
	}
	b.runners[runner] = struct{}{}
	b.mu.Unlock()

	result := runner.Run(ctx, prompt)
	b.mu.Lock()
	delete(b.runners, runner)
	b.drained.Broadcast()
	closing := b.closing
	b.mu.Unlock()
	if closing {
		return Result{ShutdownInterrupted: true}
	}

	return normalize(result)
}

// Closeは後続実行を拒否し、実行中プロセス群を停止する。
func (b *Broker) Close() {
	b.mu.Lock()
	if b.closing {
		b.mu.Unlock()

		return
	}
	b.closing = true
	runners := make([]Runner, 0, len(b.runners))
	for runner := range b.runners {
		runners = append(runners, runner)
	}
	b.mu.Unlock()

	for _, runner := range runners {
		runner.Close()
	}

	b.mu.Lock()
	for len(b.runners) != 0 {
		b.drained.Wait()
	}
	b.mu.Unlock()
}

func normalize(result Result) Result {
	if result.HTML != "" && len(result.HTML) <= maxIPCMessageBytes && !result.ShutdownRejected && !result.ShutdownInterrupted && !result.Unavailable {
		return result
	}
	if result.ShutdownRejected {
		return Result{ShutdownRejected: true}
	}
	if result.ShutdownInterrupted {
		return Result{ShutdownInterrupted: true}
	}

	return Result{Unavailable: true}
}
