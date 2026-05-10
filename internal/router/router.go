//go:generate go run ../../tools/genroutes
package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golershop.cn/internal/config"
	_ "golershop.cn/internal/logic"
	"golershop.cn/internal/middleware"
	"golershop.cn/internal/pkg/response"
)

func New(cfg *config.AppConfig) *gin.Engine {
	r := gin.New()
	// Gin 1.9+：默认 ContextWithFallback=false 时，*gin.Context 作为 context.Context 调用 Value(key)
	// 不会对任意 string key 回退到 c.Request.Context()，导致 BizCtx.Get(ctx) 读不到 WithBizContext 注入的 consts.ContextKey；
	// SetUser 使用 c.Request.Context() 能写入，与 handler 里 ctx.Value 行为不一致。
	r.ContextWithFallback = true
	if cfg.ShopSuite.AppDebug {
		r.Use(middleware.DevelopNoStore())
	}
	// 与 golershop internal/cmd/cmd.go 一致：最外层 AfterOutput → 日志/Recovery → CORS(NeverDone) → Ctx → CheckLogin
	r.Use(middleware.AfterOutput())
	r.Use(gin.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())
	r.Use(middleware.NeverDoneCtx())
	r.Use(middleware.Ctx())
	r.Use(middleware.CheckLogin())
	r.Use(middleware.BizLog())

	r.GET("/api.json", func(c *gin.Context) {
		response.Success(c, gin.H{"openapi": "todo"})
	})
	r.GET("/swagger/*any", func(c *gin.Context) {
		c.String(http.StatusOK, "swagger placeholder")
	})

	front := r.Group("/front")
	manage := r.Group("/manage")
	// 全量真实路由：由 tools/genroutes 根据 api g.Meta + internal/controller 签名生成（见 api_routes_gen.go）
	RegisterAPIRoutes(front, manage)

	return r
}
