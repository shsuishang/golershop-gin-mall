package binder

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/gogf/gf/v2/util/gconv"
	"golershop.cn/internal/pkg/gerror"
	"golershop.cn/internal/pkg/httpx"
	"golershop.cn/internal/pkg/response"
)

// NewAppError 兼容旧调用，内部转发到 gerror.NewCode。
func NewAppError(code int, message string) error {
	return gerror.NewCode(code, message)
}

// GoFrame API 结构体常见标签分工（与本包职责）：
//   - json / p / form：HTTP 入参键名，由 mergeFormValues + JSON Scan 写入 req；
//   - d：未出现对应键时的默认值，由 applyGfStructTagDefaults 补齐（ghttp 语义）；
//   - field / type（LIKE、GE、LE、IN、FIND_IN_SET_STR 等）：列表 Where 生成用，由控制器里 ml.ConvertReqToInputWhere
//     等逻辑读取，不表示 query 键名，本 binder 不处理；
//   - v / dc 等：校验与文档，由 binding.Validator 或 swagger 侧处理。
//
// mergeFormValues 将 url.Values 按多组 struct tag 写入 req（与 ghttp 一致：api 中 p: 表示 query/form 键名）。
func mergeFormValues(req any, src map[string][]string, tags ...string) error {
	for _, tag := range tags {
		if err := binding.MapFormWithTag(req, src, tag); err != nil {
			return err
		}
	}
	return nil
}

// collectParamKeys 汇总本次请求中「显式出现过的」参数名（与 GoFrame ghttp 中是否应用 d 默认一致：键未出现才填默认）。
func collectParamKeys(r *http.Request) map[string]struct{} {
	out := make(map[string]struct{})
	for k := range r.URL.Query() {
		out[k] = struct{}{}
	}
	for k := range r.PostForm {
		out[k] = struct{}{}
	}
	return out
}

// applyGfStructTagDefaults 对 api 定义中带 `d:"..."` 的字段：若对应 json / p 键未出现在请求参数中，则按默认值写入。
// 对齐 GoFrame 请求解析在未传 query/form/json 键时对 `d` 标签的处理；Gin 的 MapForm 不会读 `d`。
func applyGfStructTagDefaults(req any, seenKeys map[string]struct{}) error {
	rv := reflect.ValueOf(req)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return nil
	}
	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return nil
	}
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		sf := rt.Field(i)
		if !sf.IsExported() {
			continue
		}
		if sf.Anonymous {
			continue
		}
		jsonTag := sf.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}
		jsonKey := strings.Split(strings.TrimSpace(jsonTag), ",")[0]
		if jsonKey == "" || jsonKey == "-" {
			continue
		}
		pKey := strings.TrimSpace(sf.Tag.Get("p"))
		inRequest := false
		if _, ok := seenKeys[jsonKey]; ok {
			inRequest = true
		}
		if !inRequest && pKey != "" {
			if _, ok := seenKeys[pKey]; ok {
				inRequest = true
			}
		}
		if inRequest {
			continue
		}
		defTag, hasDef := sf.Tag.Lookup("d")
		if !hasDef {
			continue
		}
		fv := rv.Field(i)
		if !fv.CanSet() {
			continue
		}
		if err := gconv.Scan(defTag, fv.Addr().Interface()); err != nil {
			return err
		}
	}
	return nil
}

// bindRequest 将请求参数绑定到 req。
// - GET/HEAD：query 上同时按 json、p 标签映射（GoFrame api 常用 p:）。
// - POST/PUT/PATCH/DELETE：先 ParseForm；query 与 PostForm 按 json、p、form 合并；若为 application/json 且 body 非空，再按 JSON 对象键做增量合并（避免 body 为 {} 时冲掉 query 已绑定的字段）。
func bindRequest(ctx *gin.Context, req any) error {
	if err := ctx.Request.ParseMultipartForm(32 << 20); err != nil && !errors.Is(err, http.ErrNotMultipart) {
		return err
	}
	if err := ctx.Request.ParseForm(); err != nil {
		return err
	}

	seenKeys := collectParamKeys(ctx.Request)

	switch ctx.Request.Method {
	case http.MethodGet, http.MethodHead:
		if err := mergeFormValues(req, ctx.Request.URL.Query(), "json", "p"); err != nil {
			return err
		}
	default:
		if err := mergeFormValues(req, ctx.Request.URL.Query(), "json", "p"); err != nil {
			return err
		}
		if err := mergeFormValues(req, ctx.Request.PostForm, "json", "p", "form"); err != nil {
			return err
		}
		ct := strings.ToLower(strings.TrimSpace(ctx.ContentType()))
		if strings.HasPrefix(ct, "application/json") {
			body, err := io.ReadAll(ctx.Request.Body)
			if err != nil {
				return err
			}
			ctx.Request.Body = io.NopCloser(bytes.NewReader(body))
			body = bytes.TrimSpace(body)
			if len(body) > 0 && !bytes.Equal(body, []byte("null")) {
				var m map[string]interface{}
				if err := json.Unmarshal(body, &m); err != nil {
					return err
				}
				for k := range m {
					seenKeys[k] = struct{}{}
				}
				if len(m) > 0 {
					if err := gconv.Scan(m, req); err != nil {
						return err
					}
				}
			}
		}
	}

	if err := applyGfStructTagDefaults(req, seenKeys); err != nil {
		return err
	}

	if binding.Validator == nil {
		return nil
	}
	return binding.Validator.ValidateStruct(req)
}

func BindResult[T any, R any](fn func(ctx context.Context, req *T) (R, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req T
		if err := bindRequest(ctx, &req); err != nil {
			log.Printf("[binder] bind failed method=%s path=%s err=%v", ctx.Request.Method, ctx.Request.URL.Path, err)
			response.Fail(ctx, http.StatusOK, 400, "参数错误")
			return
		}

		// 传入 Request.Context()：与中间件 SetUser/WithBizContext 使用的链一致，ctx.Value(consts.ContextKey) 可读 BizCtx；
		// 同时注入 ResponseWriter，供验证码等需写响应体、但依赖 g.RequestFromCtx 在 Gin 下为 nil 的逻辑回退使用。
		reqCtx := httpx.WithResponseWriter(ctx.Request.Context(), ctx.Writer)
		data, err := fn(reqCtx, &req)
		if err != nil {
			var e *gerror.Error
			if errors.As(err, &e) {
				httpStatus := e.HTTPStatus
				if httpStatus == 0 {
					httpStatus = http.StatusOK
				}
				code := e.Code
				if code == 0 {
					code = http.StatusInternalServerError
				}
				msg := e.Message
				if msg == "" {
					msg = "系统异常"
				}
				response.Fail(ctx, httpStatus, code, msg)
				return
			}

			response.Fail(ctx, http.StatusOK, http.StatusInternalServerError, err.Error())
			return
		}

		response.Success(ctx, data)
	}
}
