package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golershop.cn/internal/pkg/response"
)

func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		response.Fail(c, http.StatusOK, http.StatusInternalServerError, "服务器异常")
		c.Abort()
	})
}
