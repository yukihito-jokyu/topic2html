// Package authは認証ユースケースのportを定義します。
package auth

import (
	"context"

	"github.com/yukihito-jokyu/topic2html/backend/domain/auth"
)

// IdentityVerifierはID検証のportです。
type IdentityVerifier interface {
	Verify(context.Context, string) (auth.VerifiedIdentity, error)
}
