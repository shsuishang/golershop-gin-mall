package httpx

import (
	"context"
	"net/http"
)

// ginResponseWriterCtxKey 用于 context.WithValue，避免与 string 等通用 key 冲突。
type ginResponseWriterCtxKey struct{}

// WithResponseWriter 将 net/http 的 ResponseWriter 写入 context（Gin 下 binder 仅传 Request.Context 时，
// g.RequestFromCtx 无法得到 ghttp.Request，验证码等需直接写响应体时可从此取 Writer）。
func WithResponseWriter(parent context.Context, w http.ResponseWriter) context.Context {
	if parent == nil || w == nil {
		return parent
	}
	return context.WithValue(parent, ginResponseWriterCtxKey{}, w)
}

// ResponseWriterFrom 返回 WithResponseWriter 注入的 Writer；未注入时为 nil。
func ResponseWriterFrom(ctx context.Context) http.ResponseWriter {
	if ctx == nil {
		return nil
	}
	v, _ := ctx.Value(ginResponseWriterCtxKey{}).(http.ResponseWriter)
	return v
}
