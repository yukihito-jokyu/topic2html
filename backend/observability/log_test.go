package observability

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/yukihito-jokyu/topic2html/backend/apperr"
)

func TestLoggerWritesSafeStructuredFields(t *testing.T) {
	var output bytes.Buffer
	logger := NewLogger(&output)
	logger.Info(context.Background(), "oauth.started")
	logger.Error(context.Background(), "oauth.failed", apperr.New(apperr.CodeRejected))
	logger.Error(context.Background(), "store.failed", errors.New("password=secret"))
	logger.RequestCompleted(context.Background(), "POST", "/admin/auth/google/start", 303)
	got := output.String()
	for _, wanted := range []string{
		"oauth.started",
		"oauth.failed",
		"rejected",
		"unavailable",
		"http.request.completed",
		"/admin/auth/google/start",
	} {
		if !bytes.Contains([]byte(got), []byte(wanted)) {
			t.Errorf("log missing %q: %s", wanted, got)
		}
	}
	if bytes.Contains([]byte(got), []byte("password=secret")) {
		t.Fatalf("unsafe error detail was logged: %s", got)
	}
	NewDiscardLogger().Info(context.Background(), "discarded")
}
