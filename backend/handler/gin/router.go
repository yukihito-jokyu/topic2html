package ginadapter

import (
	"github.com/gin-gonic/gin"
	"github.com/yukihito-jokyu/topic2html/backend/observability"
	usecaseauth "github.com/yukihito-jokyu/topic2html/backend/usecase/auth"
)

// NewRouterはHTTPルーターを作成します。
func NewRouter(oauthService usecaseauth.OAuthService, sessionService usecaseauth.AdminSessionService, logger observability.EventLogger) *gin.Engine {
	router := gin.New()
	oauthHandler := NewOAuthHandler(oauthService, logger)
	sessionHandler := NewSessionHandler(sessionService, logger)
	router.HandleMethodNotAllowed = true
	router.Use(safeRecovery(logger), requestLog(logger))
	router.GET("/health", healthHandler)
	router.POST("/admin/auth/google/start", noStore, oauthHandler.Start)
	router.GET("/auth/google/callback", noStore, oauthHandler.Callback)
	router.GET("/admin/auth/session", noStore, sessionHandler.Bootstrap)
	router.POST("/admin/auth/logout", noStore, sessionHandler.Logout)

	return router
}
