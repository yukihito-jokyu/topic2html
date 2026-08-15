package ginadapter

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yukihito-jokyu/topic2html/backend/apperr"
	"github.com/yukihito-jokyu/topic2html/backend/observability"
	usecaseauth "github.com/yukihito-jokyu/topic2html/backend/usecase/auth"
)

// SessionHandlerは管理session HTTP契約をusecaseへ変換します。
type SessionHandler struct {
	service usecaseauth.AdminSessionService
	logger  observability.EventLogger
}

// NewSessionHandlerは管理session handlerを作成します。
func NewSessionHandler(service usecaseauth.AdminSessionService, logger observability.EventLogger) *SessionHandler {
	return &SessionHandler{
		service: service,
		logger:  logger,
	}
}

// Bootstrapは管理画面へCSRF tokenを初期化応答で返します。
func (h *SessionHandler) Bootstrap(c *gin.Context) {
	reference := sessionCookie(c.Request)
	output, err := h.service.Bootstrap(c.Request.Context(), reference)
	if err != nil {
		h.authenticationUnavailable(c)

		return
	}
	if !output.Authenticated {
		if reference != "" {
			deleteCookie(c, adminSessionCookie)
		}
		c.JSON(http.StatusOK, gin.H{"authenticated": false})

		return
	}
	c.JSON(http.StatusOK, gin.H{
		"authenticated": true,
		"csrf_token":    output.CSRFToken,
	})
}

// Logoutは有効sessionをCSRF保護して失効します。
func (h *SessionHandler) Logout(c *gin.Context) {
	decision, err := h.service.Logout(c.Request.Context(), usecaseauth.SessionInput{
		SessionReference: sessionCookie(c.Request),
		Origins:          c.Request.Header.Values("Origin"),
		CSRFToken:        c.Request.Header.Get("X-CSRF-Token"),
	})
	if err != nil {
		h.authenticationUnavailable(c)

		return
	}
	if decision == usecaseauth.GuardForbidden {
		c.JSON(http.StatusForbidden, errorResponse("forbidden"))

		return
	}
	deleteCookie(c, adminSessionCookie)
	c.JSON(http.StatusOK, gin.H{"authenticated": false})
}

func (h *SessionHandler) authenticationUnavailable(c *gin.Context) {
	h.logger.Error(c.Request.Context(), "admin.session.unavailable", apperr.New(apperr.CodeUnavailable))
	c.JSON(http.StatusServiceUnavailable, errorResponse("authentication_unavailable"))
}

func sessionCookie(request *http.Request) string {
	cookie, err := request.Cookie(adminSessionCookie)
	if err != nil {
		return ""
	}

	return cookie.Value
}

func errorResponse(code string) gin.H {
	return gin.H{"error": gin.H{"code": code}}
}
