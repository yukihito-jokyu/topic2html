package auth

// VerifiedIdentityは検証済みのID情報です。
type VerifiedIdentity struct {
	Email         string
	EmailVerified bool
}
