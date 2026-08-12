package http

import (
	"context"
	"net/http"

	"github.com/yukihito-jokyu/topic2html/backend/internal/domain"
)

type ReadinessChecker interface {
	Check(context.Context) domain.Readiness
}

func statusCode(ready bool) int {
	if ready {
		return http.StatusOK
	}

	return http.StatusServiceUnavailable
}
