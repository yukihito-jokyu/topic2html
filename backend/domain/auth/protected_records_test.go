package auth

import (
	"testing"
	"time"
)

func TestProtectedRecordValidation(t *testing.T) {
	now := time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)
	validTransaction := OAuthTransaction{
		ID:                     "id",
		ReferenceHash:          Hash{1},
		StateHash:              Hash{2},
		NonceHash:              Hash{3},
		PKCEVerifierCiphertext: Ciphertext{4},
		ReturnPath:             "/admin",
		CreatedAt:              now,
		ExpiresAt:              now.Add(OAuthTransactionLifetime),
	}
	validSession := AdminSession{
		ID:                  "id",
		ReferenceHash:       Hash{1},
		AuthorizedEmail:     "admin@example.test",
		CSRFTokenHash:       Hash{2},
		CSRFTokenCiphertext: Ciphertext{3},
		CreatedAt:           now,
		LastMutationAt:      now,
		AbsoluteExpiresAt:   now.Add(SessionAbsoluteLifetime),
		IdleExpiresAt:       now.Add(SessionIdleLifetime),
	}
	for _, tt := range []struct {
		name string
		err  error
	}{
		{"valid transaction", validTransaction.Validate()},
		{"invalid transaction", OAuthTransaction{}.Validate()},
		{"valid session", validSession.Validate()},
		{"invalid session", AdminSession{}.Validate()},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if (tt.err != nil) != (tt.name == "invalid transaction" || tt.name == "invalid session") {
				t.Fatalf("error = %v", tt.err)
			}
		})
	}
}
