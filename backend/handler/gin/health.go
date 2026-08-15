package ginadapter

import "github.com/gin-gonic/gin"

func healthHandler(c *gin.Context) {
	c.Status(204)
}
