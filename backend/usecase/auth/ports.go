package auth

import (
	"context"

	"github.com/yukihito-jokyu/topic2html/backend/domain/auth"
)

// IdentityVerifierはID検証のportです。
type IdentityVerifier interface {
	Verify(context.Context, string) (auth.VerifiedIdentity, error)
}

// AuthorizationRequestはGoogle認可要求の公開可能な値です。
type AuthorizationRequest struct {
	State         string
	Nonce         string
	CodeChallenge string
}

// OAuthProviderはGoogle固有の通信とOIDC検証を隠すportです。
type OAuthProvider interface {
	AuthorizationURL(context.Context, AuthorizationRequest) (string, error)
	ExchangeAndVerify(context.Context, string, string, auth.Hash) (auth.VerifiedIdentity, error)
}

// Securityは乱数生成と保護値操作のportです。
type Security interface {
	RandomBytes(int) ([]byte, error)
	Hash([]byte) auth.Hash
	Seal([]byte) (auth.Ciphertext, error)
	Open(auth.Ciphertext) ([]byte, error)
}

// Clockは時刻取得のportです。
type Clock interface {
	Now() auth.Time
}

// EventLoggerは秘密値を出力しないアプリケーションイベント記録のportです。
type EventLogger interface {
	Info(context.Context, string)
	Error(context.Context, string, error)
}

// StartInputはOAuth開始に必要なHTTP入力です。
type StartInput struct {
	Origins           []string
	ReturnPaths       []string
	PreviousReference string
}

// StartOutputは認可開始後にhandlerがcookieとredirectへ変換する結果です。
type StartOutput struct {
	TransactionReference string
	AuthorizationURL     string
}

// CallbackInputはGoogle callbackに必要なHTTP入力です。
type CallbackInput struct {
	TransactionReference string
	Code                 string
	State                string
	ProviderError        string
}

// CallbackOutputは認証成功後にhandlerがcookieとredirectへ変換する結果です。
type CallbackOutput struct {
	SessionReference string
	ReturnPath       string
}

// OAuthServiceは認証操作のapplication portです。
type OAuthService interface {
	Start(context.Context, StartInput) (StartOutput, error)
	Callback(context.Context, CallbackInput) (CallbackOutput, error)
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
