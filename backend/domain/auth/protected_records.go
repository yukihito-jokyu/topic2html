package auth

import (
	"errors"
	"time"
)

// Timeは外部I/Oを持たない時刻の別名です。
type Time = time.Time

// Hashは保護値のSHA-256です。
type Hash []byte

// CiphertextはServer保護鍵で暗号化済みの値です。
type Ciphertext []byte

const (
	OAuthTransactionLifetime = 10 * time.Minute
	SessionAbsoluteLifetime  = 8 * time.Hour
	SessionIdleLifetime      = 30 * time.Minute
	ProtectedRecordRetention = 24 * time.Hour
)

// OAuthTransactionは保護されたOAuth transactionです。
type OAuthTransaction struct {
	ID                     string
	ReferenceHash          Hash
	StateHash              Hash
	NonceHash              Hash
	PKCEVerifierCiphertext Ciphertext
	ReturnPath             string
	CreatedAt              Time
	ExpiresAt              Time
}

// AdminSessionはServer側のopaque session記録です。
type AdminSession struct {
	ID                string
	ReferenceHash     Hash
	AuthorizedEmail   string
	CSRFTokenHash     Hash
	CreatedAt         Time
	LastMutationAt    Time
	AbsoluteExpiresAt Time
	IdleExpiresAt     Time
	RevokedAt         *Time
}

// ValidateOAuthTransactionは保存前の期限・秘密値表現の不変条件を確認します。
func (record OAuthTransaction) Validate() error {
	if record.ID == "" || len(record.ReferenceHash) == 0 || len(record.StateHash) == 0 || len(record.NonceHash) == 0 || len(record.PKCEVerifierCiphertext) == 0 || record.ReturnPath != "/admin" || record.CreatedAt.IsZero() || !record.ExpiresAt.Equal(record.CreatedAt.Add(OAuthTransactionLifetime)) {
		return errors.New("invalid OAuth transaction")
	}

	return nil
}

// ValidateAdminSessionは保存前のsession期限・秘密値表現の不変条件を確認します。
func (session AdminSession) Validate() error {
	if session.ID == "" || len(session.ReferenceHash) == 0 || session.AuthorizedEmail == "" || len(session.CSRFTokenHash) == 0 || session.CreatedAt.IsZero() || !session.AbsoluteExpiresAt.Equal(session.CreatedAt.Add(SessionAbsoluteLifetime)) || !session.LastMutationAt.Equal(session.CreatedAt) || !session.IdleExpiresAt.Equal(session.CreatedAt.Add(SessionIdleLifetime)) {
		return errors.New("invalid admin session")
	}

	return nil
}
