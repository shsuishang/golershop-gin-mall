package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"golershop.cn/internal/service"
)

// BizLog 对齐 cmd.go 中 HookAfterOutput：/manage/* → OperateLog，/front/* → OperateAccess（在 BizCtx.AfterOutput 之前执行，与 ghttp 注册顺序一致）。
func BizLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/manage/") && c.Request.Method == "POST" {
			service.LogAction().OperateLogWithHTTP(c.Request.Context(), c.Request)
			return
		}
		if strings.HasPrefix(path, "/front/") {
			service.AccessHistory().OperateAccessWithHTTP(c.Request.Context(), c.Request)
		}
	}
}
