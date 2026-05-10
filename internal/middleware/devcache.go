package middleware

import "github.com/gin-gonic/gin"

// DevelopNoStore 对齐 golershop internal/cmd/cmd.go 中 gmode.IsDevelop 时 HookBeforeServe 设置的 Cache-Control: no-store。
func DevelopNoStore() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		c.Next()
	}
}
