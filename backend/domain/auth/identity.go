// Package authは認証ドメインを定義します。
package auth

// VerifiedIdentityは検証済みのID情報です。
type VerifiedIdentity struct {
	Email         string
	EmailVerified bool
}
