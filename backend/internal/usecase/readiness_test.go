package usecase

import (
	"context"
	"testing"
)

func TestReadinessServiceReturnsReady(t *testing.T) {
	t.Parallel()

	result := NewReadinessService().Check(context.Background())

	if !result.Ready {
		t.Fatal("ready = false, want true")
	}
}
