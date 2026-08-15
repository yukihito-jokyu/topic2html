package auth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/yukihito-jokyu/topic2html/backend/apperr"
	"github.com/yukihito-jokyu/topic2html/backend/domain/auth"
)

// ServiceはOAuth開始とcallbackのapplication実装です。
type Service struct {
	store         ProtectedRecordStore
	provider      OAuthProvider
	security      Security
	clock         Clock
	logger        EventLogger
	trustedOrigin string
	allowedEmail  string
}

// DependenciesはOAuth usecaseが必要とする外部境界です。
type Dependencies struct {
	Store    ProtectedRecordStore
	Provider OAuthProvider
	Security Security
	Clock    Clock
	Logger   EventLogger
}

// NewServiceはOAuth認証usecaseを作成します。
func NewService(dependencies Dependencies, trustedOrigin, allowedEmail string) (*Service, error) {
	if dependencies.Store == nil || dependencies.Provider == nil || dependencies.Security == nil || dependencies.Clock == nil || dependencies.Logger == nil || trustedOrigin == "" || allowedEmail == "" {
		return nil, apperr.New(apperr.CodeInvalidConfiguration)
	}

	return &Service{
		store:         dependencies.Store,
		provider:      dependencies.Provider,
		security:      dependencies.Security,
		clock:         dependencies.Clock,
		logger:        dependencies.Logger,
		trustedOrigin: trustedOrigin,
		allowedEmail:  allowedEmail,
	}, nil
}

// StartはOAuth transactionを作成し、Google認可URLを返します。
func (s *Service) Start(ctx context.Context, input StartInput) (StartOutput, error) {
	if !s.validStartInput(input) {
		return StartOutput{}, s.startError(ctx, apperr.CodeInvalidRequest)
	}
	values, err := s.newOAuthTransactionValues()
	if err != nil {
		return StartOutput{}, s.startError(ctx, apperr.CodeUnavailable)
	}
	ciphertext, err := s.security.Seal([]byte(values.verifier))
	if err != nil {
		return StartOutput{}, s.startError(ctx, apperr.CodeUnavailable)
	}
	challengeHash := sha256.Sum256([]byte(values.verifier))
	authorizationURL, err := s.provider.AuthorizationURL(ctx, AuthorizationRequest{
		State:         values.state,
		Nonce:         values.nonce,
		CodeChallenge: base64.RawURLEncoding.EncodeToString(challengeHash[:]),
	})
	if err != nil || authorizationURL == "" {
		return StartOutput{}, s.startError(ctx, apperr.CodeOf(err))
	}
	now := s.clock.Now()
	record := auth.OAuthTransaction{
		ID:                     values.id,
		ReferenceHash:          s.security.Hash([]byte(values.reference)),
		StateHash:              s.security.Hash([]byte(values.state)),
		NonceHash:              s.security.Hash([]byte(values.nonce)),
		PKCEVerifierCiphertext: ciphertext,
		ReturnPath:             startReturnPath(input),
		CreatedAt:              now,
		ExpiresAt:              now.Add(auth.OAuthTransactionLifetime),
	}
	var previousHash auth.Hash
	if input.PreviousReference != "" {
		previousHash = s.security.Hash([]byte(input.PreviousReference))
	}
	if err := s.store.ReplaceOAuthTransaction(ctx, previousHash, record); err != nil {
		return StartOutput{}, s.startError(ctx, apperr.CodeUnavailable)
	}
	s.logger.Info(ctx, "oauth.transaction.created")

	return StartOutput{
		TransactionReference: values.reference,
		AuthorizationURL:     authorizationURL,
	}, nil
}

// Callbackはtransactionを一回消費してOIDC検証後にsessionを保存します。
func (s *Service) Callback(ctx context.Context, input CallbackInput) (CallbackOutput, error) {
	if !validCallbackInput(input) {
		return CallbackOutput{}, s.callbackError(ctx, apperr.CodeInvalidRequest)
	}
	referenceHash := s.security.Hash([]byte(input.TransactionReference))
	stateHash := s.security.Hash([]byte(input.State))
	record, found, err := s.store.ConsumeOAuthTransaction(ctx, referenceHash, stateHash, s.clock.Now())
	if err != nil {
		return CallbackOutput{}, s.callbackError(ctx, apperr.CodeUnavailable)
	}
	if !found {
		return CallbackOutput{}, s.callbackError(ctx, apperr.CodeRejected)
	}
	if input.ProviderError != "" {
		return CallbackOutput{}, s.callbackError(ctx, apperr.CodeRejected)
	}
	if err := record.Validate(); err != nil {
		return CallbackOutput{}, s.callbackError(ctx, apperr.CodeRejected)
	}
	verifier, err := s.security.Open(record.PKCEVerifierCiphertext)
	if err != nil || len(verifier) == 0 {
		return CallbackOutput{}, s.callbackError(ctx, apperr.CodeUnavailable)
	}
	identity, err := s.provider.ExchangeAndVerify(ctx, input.Code, string(verifier), record.NonceHash)
	if err != nil || !identity.EmailVerified || identity.Email != s.allowedEmail {
		return CallbackOutput{}, s.callbackError(ctx, apperr.CodeOf(err))
	}
	sessionReference, err := s.createAdminSession(ctx, identity.Email)
	if err != nil {
		return CallbackOutput{}, s.callbackError(ctx, apperr.CodeUnavailable)
	}
	s.logger.Info(ctx, "oauth.session.created")

	return CallbackOutput{
		SessionReference: sessionReference,
		ReturnPath:       record.ReturnPath,
	}, nil
}

func (s *Service) startError(ctx context.Context, code apperr.Code) error {
	err := apperr.New(code)
	s.logger.Error(ctx, "oauth.start.failed", err)

	return err
}

func (s *Service) callbackError(ctx context.Context, code apperr.Code) error {
	err := apperr.New(code)
	s.logger.Error(ctx, "oauth.callback.failed", err)

	return err
}

type oauthTransactionValues struct {
	reference string
	state     string
	nonce     string
	verifier  string
	id        string
}

func (s *Service) validStartInput(input StartInput) bool {
	return len(input.Origins) == 1 &&
		input.Origins[0] == s.trustedOrigin &&
		(len(input.ReturnPaths) == 0 || (len(input.ReturnPaths) == 1 && input.ReturnPaths[0] == "/admin"))
}

func startReturnPath(input StartInput) string {
	if len(input.ReturnPaths) == 1 {
		return input.ReturnPaths[0]
	}

	return "/admin"
}

func validCallbackInput(input CallbackInput) bool {
	return input.TransactionReference != "" &&
		input.State != "" &&
		((input.Code == "") != (input.ProviderError == ""))
}

func (s *Service) newOAuthTransactionValues() (oauthTransactionValues, error) {
	reference, err := s.randomString()
	if err != nil {
		return oauthTransactionValues{}, err
	}
	state, err := s.randomString()
	if err != nil {
		return oauthTransactionValues{}, err
	}
	nonce, err := s.randomString()
	if err != nil {
		return oauthTransactionValues{}, err
	}
	verifier, err := s.randomString()
	if err != nil {
		return oauthTransactionValues{}, err
	}
	id, err := s.randomUUID()
	if err != nil {
		return oauthTransactionValues{}, err
	}

	return oauthTransactionValues{
		reference: reference,
		state:     state,
		nonce:     nonce,
		verifier:  verifier,
		id:        id,
	}, nil
}

func (s *Service) createAdminSession(ctx context.Context, email string) (string, error) {
	reference, err := s.randomString()
	if err != nil {
		return "", err
	}
	csrfToken, err := s.randomString()
	if err != nil {
		return "", err
	}
	csrfTokenCiphertext, err := s.security.Seal([]byte(csrfToken))
	if err != nil {
		return "", err
	}
	sessionID, err := s.randomUUID()
	if err != nil {
		return "", err
	}
	now := s.clock.Now()
	if err := s.store.CreateAdminSession(ctx, auth.AdminSession{
		ID:                  sessionID,
		ReferenceHash:       s.security.Hash([]byte(reference)),
		AuthorizedEmail:     email,
		CSRFTokenHash:       s.security.Hash([]byte(csrfToken)),
		CSRFTokenCiphertext: csrfTokenCiphertext,
		CreatedAt:           now,
		LastMutationAt:      now,
		AbsoluteExpiresAt:   now.Add(auth.SessionAbsoluteLifetime),
		IdleExpiresAt:       now.Add(auth.SessionIdleLifetime),
	}); err != nil {
		return "", err
	}

	return reference, nil
}

func (s *Service) randomString() (string, error) {
	value, err := s.randomBytes(32)
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(value), nil
}

func (s *Service) randomUUID() (string, error) {
	value, err := s.randomBytes(16)
	if err != nil {
		return "", err
	}
	value[6] = value[6]&0x0f | 0x40
	value[8] = value[8]&0x3f | 0x80

	return fmt.Sprintf(
		"%s-%s-%s-%s-%s",
		hex.EncodeToString(value[0:4]),
		hex.EncodeToString(value[4:6]),
		hex.EncodeToString(value[6:8]),
		hex.EncodeToString(value[8:10]),
		hex.EncodeToString(value[10:16]),
	), nil
}

func (s *Service) randomBytes(size int) ([]byte, error) {
	value, err := s.security.RandomBytes(size)
	if err != nil || len(value) != size {
		return nil, errors.New("random value is invalid")
	}

	return value, nil
}

// SystemClockは本番の時刻portです。
type SystemClock struct{}

// Nowは現在時刻を返します。
func (SystemClock) Now() auth.Time { return time.Now().UTC() }
