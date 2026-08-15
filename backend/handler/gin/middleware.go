package ginadapter

import "github.com/gin-gonic/gin"

import "github.com/yukihito-jokyu/topic2html/backend/observability"
import "github.com/yukihito-jokyu/topic2html/backend/apperr"

func noStore(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	c.Next()
}

func requestLog(logger observability.EventLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		logger.RequestCompleted(c.Request.Context(), c.Request.Method, c.FullPath(), c.Writer.Status())
	}
}

func safeRecovery(logger observability.EventLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recover() != nil {
				logger.Error(c.Request.Context(), "http.request.panic", apperr.New(apperr.CodeInternal))
				c.AbortWithStatus(500)
			}
		}()
		c.Next()
	}
}
