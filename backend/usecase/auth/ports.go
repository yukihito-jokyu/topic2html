package auth

import (
	"context"

	"github.com/yukihito-jokyu/topic2html/backend/domain/auth"
)

// IdentityVerifierはID検証のportです。
type IdentityVerifier interface {
	Verify(context.Context, string) (auth.VerifiedIdentity, error)
}

// ProtectedRecordStoreは保護記録の永続化境界です。
type ProtectedRecordStore interface {
	ReplaceOAuthTransaction(context.Context, auth.Hash, auth.OAuthTransaction) error
	ConsumeOAuthTransaction(context.Context, auth.Hash, auth.Hash, auth.Time) (auth.OAuthTransaction, bool, error)
	CreateAdminSession(context.Context, auth.AdminSession) error
	FindAdminSession(context.Context, auth.Hash) (auth.AdminSession, bool, error)
	RevokeAdminSession(context.Context, auth.Hash, auth.Time) (bool, error)
	RunAdminStateChange(context.Context, auth.Hash, auth.Time, AdminStateChangeOperation) (bool, error)
	DeleteExpiredProtectedRecords(context.Context, auth.Time) error
}

// AdminStateChangeOperationは後続の永続化更新です。
type AdminStateChangeOperation func(context.Context) error
