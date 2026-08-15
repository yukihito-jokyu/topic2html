package ginadapter

import "github.com/gin-gonic/gin"

// NewRouterはHTTPルーターを作成します。
func NewRouter() *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.GET("/health", func(c *gin.Context) {
		c.Status(204)
	})

	return router
}
