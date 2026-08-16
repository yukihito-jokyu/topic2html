package auth

import (
	"context"
	"crypto/subtle"

	"github.com/yukihito-jokyu/topic2html/backend/apperr"
	"github.com/yukihito-jokyu/topic2html/backend/domain/auth"
)

// Bootstrapは有効sessionのCSRF tokenを安全に復元します。
func (s *Service) Bootstrap(ctx context.Context, reference string) (SessionBootstrapOutput, error) {
	session, decision, err := s.findAuthorizedSession(ctx, reference)
	if err != nil {
		return SessionBootstrapOutput{}, s.sessionError(ctx, "session.bootstrap.failed")
	}
	if decision != GuardAllowed {
		return SessionBootstrapOutput{
			Authenticated: false,
		}, nil
	}
	plaintext, err := s.security.Open(session.CSRFTokenCiphertext)
	if err != nil || len(plaintext) == 0 || subtle.ConstantTimeCompare(s.security.Hash(plaintext), session.CSRFTokenHash) != 1 {
		return SessionBootstrapOutput{}, s.sessionError(ctx, "session.bootstrap.failed")
	}

	return SessionBootstrapOutput{
		Authenticated: true,
		CSRFToken:     string(plaintext),
	}, nil
}

// AuthorizeReadは管理読取りに有効sessionを要求します。
func (s *Service) AuthorizeRead(ctx context.Context, reference string) (GuardDecision, error) {
	_, decision, err := s.findAuthorizedSession(ctx, reference)
	if err != nil {
		return GuardUnauthenticated, s.sessionError(ctx, "admin.read.failed")
	}

	return decision, nil
}

// RunAuthorizedAdminStateChangeは認可・CSRF成功と業務更新を同一transactionに含めます。
func (s *Service) RunAuthorizedAdminStateChange(ctx context.Context, input SessionInput, operation AdminStateChangeOperation) (GuardDecision, error) {
	session, decision, err := s.authorizeMutation(ctx, input)
	if err != nil || decision != GuardAllowed {
		return decision, err
	}
	if subtle.ConstantTimeCompare(s.security.Hash([]byte(input.CSRFToken)), session.CSRFTokenHash) != 1 {
		return GuardForbidden, nil
	}
	updated, err := s.store.RunAuthorizedAdminStateChange(ctx, s.security.Hash([]byte(input.SessionReference)), s.allowedEmail, s.security.Hash([]byte(input.CSRFToken)), s.clock.Now(), operation)
	if err != nil {
		return GuardUnauthenticated, s.sessionError(ctx, "admin.mutation.failed")
	}
	if !updated {
		return GuardUnauthenticated, nil
	}

	return GuardAllowed, nil
}

// Logoutはtrusted Originのlogoutを冪等に匿名状態へ収束させます。
func (s *Service) Logout(ctx context.Context, input SessionInput) (GuardDecision, error) {
	if !s.trustedOrigins(input.Origins) {
		return GuardForbidden, nil
	}
	session, decision, err := s.findAuthorizedSession(ctx, input.SessionReference)
	if err != nil {
		return GuardUnauthenticated, s.sessionError(ctx, "session.logout.failed")
	}
	if decision != GuardAllowed {
		return GuardAllowed, nil
	}
	if subtle.ConstantTimeCompare(s.security.Hash([]byte(input.CSRFToken)), session.CSRFTokenHash) != 1 {
		return GuardForbidden, nil
	}
	revoked, err := s.store.RevokeAdminSession(ctx, s.security.Hash([]byte(input.SessionReference)), s.clock.Now())
	if err != nil {
		return GuardUnauthenticated, s.sessionError(ctx, "session.logout.failed")
	}
	if !revoked {
		return GuardAllowed, nil
	}

	return GuardAllowed, nil
}

func (s *Service) authorizeMutation(ctx context.Context, input SessionInput) (auth.AdminSession, GuardDecision, error) {
	if !s.trustedOrigins(input.Origins) {
		return auth.AdminSession{}, GuardForbidden, nil
	}

	return s.findAuthorizedSession(ctx, input.SessionReference)
}

func (s *Service) findAuthorizedSession(ctx context.Context, reference string) (auth.AdminSession, GuardDecision, error) {
	if reference == "" {
		return auth.AdminSession{}, GuardUnauthenticated, nil
	}
	session, found, err := s.store.FindAdminSession(ctx, s.security.Hash([]byte(reference)))
	if err != nil {
		return auth.AdminSession{}, GuardUnauthenticated, err
	}
	now := s.clock.Now()
	if !found || session.RevokedAt != nil || !session.AbsoluteExpiresAt.After(now) || !session.IdleExpiresAt.After(now) || session.AuthorizedEmail != s.allowedEmail || len(session.CSRFTokenCiphertext) == 0 {
		return auth.AdminSession{}, GuardUnauthenticated, nil
	}

	return session, GuardAllowed, nil
}

func (s *Service) trustedOrigins(origins []string) bool {
	return len(origins) == 1 && origins[0] == s.trustedOrigin
}

func (s *Service) sessionError(ctx context.Context, event string) error {
	err := apperr.New(apperr.CodeUnavailable)
	s.logger.Error(ctx, event, err)

	return err
}
