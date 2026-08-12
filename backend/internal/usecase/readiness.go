package usecase

import (
	"context"

	"github.com/yukihito-jokyu/topic2html/backend/internal/domain"
)

type ReadinessService struct{}

// NewReadinessService creates a readiness use case.
func NewReadinessService() *ReadinessService {
	return &ReadinessService{}
}

// Check reports whether the server can accept requests.
func (service *ReadinessService) Check(context.Context) domain.Readiness {
	return domain.Readiness{Ready: true}
}
