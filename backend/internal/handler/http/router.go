package http

import "github.com/gin-gonic/gin"

// NewRouter creates the Gin HTTP boundary.
func NewRouter(readiness ReadinessChecker) *gin.Engine {
	router := gin.New()
	router.GET("/readyz", func(context *gin.Context) {
		result := readiness.Check(context.Request.Context())
		context.JSON(statusCode(result.Ready), gin.H{"ready": result.Ready})
	})

	return router
}
