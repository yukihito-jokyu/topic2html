package apperr

import (
	"errors"
	"testing"
)

func TestCodeOf(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		want    Code
		wantMsg string
	}{
		{
			name:    "application error",
			err:     New(CodeRejected),
			want:    CodeRejected,
			wantMsg: CodeRejected.String(),
		},
		{
			name: "unclassified error",
			err:  errors.New("database password: secret"),
			want: CodeUnavailable,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CodeOf(tt.err); got != tt.want {
				t.Fatalf("CodeOf() = %q, want %q", got, tt.want)
			}
			if tt.wantMsg != "" && tt.err.Error() != tt.wantMsg {
				t.Errorf("Error() = %q, want %q", tt.err, tt.wantMsg)
			}
		})
	}
	for _, code := range []Code{
		CodeInvalidConfiguration,
		CodeInvalidRequest,
		CodeRejected,
		CodeUnavailable,
		CodeInternal,
		Code(99),
	} {
		if code.String() == "" {
			t.Fatal("empty code")
		}
	}
}
