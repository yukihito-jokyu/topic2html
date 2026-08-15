package ginadapter

import (
	"github.com/gin-gonic/gin"
	"github.com/yukihito-jokyu/topic2html/backend/observability"
	usecaseauth "github.com/yukihito-jokyu/topic2html/backend/usecase/auth"
)

// NewRouterはHTTPルーターを作成します。
func NewRouter(service usecaseauth.OAuthService, logger observability.EventLogger) *gin.Engine {
	router := gin.New()
	oauthHandler := NewOAuthHandler(service, logger)
	router.HandleMethodNotAllowed = true
	router.Use(safeRecovery(logger), requestLog(logger))
	router.GET("/health", healthHandler)
	router.POST("/admin/auth/google/start", noStore, oauthHandler.Start)
	router.GET("/auth/google/callback", noStore, oauthHandler.Callback)

	return router
}
