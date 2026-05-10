// Gin 适配层：与 internal/logic/shopsuite/middleware/middleware.go 逐段对齐，
// 仅替换 ghttp.Request / r.Response / r.Middleware.Next 为 gin.Context 等价写法；导出名与 service.Middleware() 一致。
// AfterOutput 对应 cmd 中 BizCtx 的 HookAfterOutput（Gin 用 AfterOutputCtx）。
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/mallsuite/gocore/core/ml"
	"golershop.cn/internal/consts"
	"golershop.cn/internal/dao"
	"golershop.cn/internal/model"
	"golershop.cn/internal/model/do"
	"golershop.cn/internal/service"
	"golershop.cn/utility"
)

// 管理端接口权限别名：对齐 Java @PreAuthorize 的权限点与路由不一致场景。
var managePermissionAliases = map[string][]string{
	"/manage/pay/consumeDeposit/offline": {"/manage/trade/orderBase/finance"},
	"/manage/o2o/chainItem/editState":    {"/manage/o2o/chainItem/edit"},
}

// 管理端查询接口放行：仅跳过 menu_permission 精确匹配，登录与角色校验仍生效。
var managePermissionBypassPaths = map[string]struct{}{
	"/manage/shop/storeExpressLogistics/list": {},
	"/manage/shop/storeShippingAddress/list":  {},
	"/manage/shop/storeTransportType/list":    {},
	"/manage/account/userLevel/list":          {},
}

// AfterOutput 对应 cmd HookAfterOutput 的 BizCtx 缓存清理（最外层 Use 最先注册）。
func AfterOutput() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		service.BizCtx().AfterOutputCtx(c.Request.Context())
	}
}

// I18N 与 logic middleware.I18N 一致。
func I18N() gin.HandlerFunc {
	return func(c *gin.Context) {
		//r.SetCtx(gi18n.WithLanguage(r.Context(), r.GetString("source_lang", "zh-CN")))
		c.Next()
	}
}

// Ctx 与 logic middleware.Ctx 一致（Gin 无 ghttp.Session，Session 为零值；业务上下文用 WithBizContext 注入）。
func Ctx() gin.HandlerFunc {
	return func(c *gin.Context) {
		customCtx := &model.Context{}
		ctx := service.BizCtx().WithBizContext(c.Request.Context(), customCtx, c.Request)
		c.Request = c.Request.WithContext(ctx)

		if c.Request.URL.Path == "/front/sys/config/info" {
			c.Next()
			return
		}

		// 须用 g.Cfg()：gfpreboot 仅挂载 g.Cfg()，gcfg.Instance() 为另一套默认适配器，读不到 configs 会导致永不解析 Authorization。
		tokenHeader, _ := g.Cfg().Get(c.Request.Context(), "jwt.tokenHeader")
		token := c.GetHeader(tokenHeader.String())

		tokenPrefix, _ := g.Cfg().Get(c.Request.Context(), "jwt.tokenPrefix")
		token = gstr.Replace(token, tokenPrefix.String(), "")
		token = gstr.Trim(token)

		if token == "" {
			token = permKeyString(c)
		}

		if token != "" {
			claim, err := utility.ParseToken(c.Request.Context(), token)
			if err != nil {
				fmt.Println("解析token出现错误：", err)
			} else if time.Now().Unix() > claim.ExpiresAt {
				fmt.Println("时间超时")
			} else {
				clientId := consts.MOBILE_CLIENT_ID
				if strings.HasPrefix(c.Request.RequestURI, "/manage") {
					clientId = consts.ADMIN_CLIENT_ID
				} else {
					if strings.HasPrefix(c.Request.RequestURI, "/front/account/user/info") {
						clientId = consts.ADMIN_CLIENT_ID
					}
				}

				service.BizCtx().SetUser(c.Request.Context(), &model.ContextUser{
					UserId:       claim.UserId,
					UserAccount:  claim.UserAccount,
					UserNickname: claim.UserNickname,
					UserSalt:     claim.UserSalt,
					ClientId:     clientId,
				})
			}
		}

		c.Next()
	}
}

func permKeyString(c *gin.Context) string {
	if v := c.Query("perm_key"); v != "" {
		return strings.TrimSpace(v)
	}
	_ = c.Request.ParseForm()
	return strings.TrimSpace(c.Request.PostFormValue("perm_key"))
}

// Auth 与 logic middleware.Auth 一致。
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if service.Login().IsSignedIn(c.Request.Context()) {
			c.Next()
		} else {
			c.AbortWithStatus(http.StatusForbidden)
		}
	}
}

const corsDefaultAllowHeaders = "Origin,Content-Type,Accept,User-Agent,Cookie,Authorization,X-Auth-Token,X-Requested-With"

const corsSupportedMethods = "GET,PUT,POST,DELETE,PATCH,HEAD,CONNECT,OPTIONS,TRACE"

func mergeCORSAllowHeaders(reqHeaders string) string {
	allowHeaders := corsDefaultAllowHeaders
	if reqHeaders == "" {
		return allowHeaders
	}
	allowSet := make(map[string]struct{})
	for _, h := range gstr.SplitAndTrim(corsDefaultAllowHeaders, ",") {
		if h != "" {
			allowSet[h] = struct{}{}
		}
	}
	merged := corsDefaultAllowHeaders
	for _, h := range gstr.SplitAndTrim(reqHeaders, ",") {
		if h == "" {
			continue
		}
		if _, ok := allowSet[h]; !ok {
			merged += "," + h
			allowSet[h] = struct{}{}
		}
	}
	return merged
}

// CORS 与 logic middleware.CORS 一致（ghttp CORSDefault 语义，OPTIONS 中止后续处理）。
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		allowHeaders := mergeCORSAllowHeaders(c.GetHeader("Access-Control-Request-Headers"))
		allowOrigin := "*"
		if o := c.GetHeader("Origin"); o != "" {
			allowOrigin = o
		} else if ref := c.Request.Referer(); ref != "" {
			if u, err := url.Parse(ref); err == nil {
				allowOrigin = u.Scheme + "://" + u.Host
			} else {
				allowOrigin = ref
			}
		}
		c.Header("Access-Control-Allow-Origin", allowOrigin)
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", allowHeaders)
		c.Header("Access-Control-Max-Age", "3628800")
		c.Header("Access-Control-Allow-Methods", corsSupportedMethods)

		// 无匹配路由时 Gin 会先设 status=404 再跑 NoRoute 链；此处若仍判断 Status()==0 会跳过写状态，预检一直 404。
		if strings.EqualFold(c.Request.Method, http.MethodOptions) {
			c.Status(http.StatusOK)
			c.Abort()
			return
		}
		c.Next()
	}
}

// NeverDoneCtx 与 logic middleware.NeverDoneCtx 一致。
func NeverDoneCtx() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request = c.Request.WithContext(gctx.NeverDone(c.Request.Context()))
		c.Next()
	}
}

func writeJSONExit(c *gin.Context, payload model.DefaultHandlerResponse) {
	c.AbortWithStatusJSON(http.StatusOK, payload)
}

// CheckLogin 与 logic middleware.CheckLogin 一致。
func CheckLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		u, _ := url.Parse(c.Request.RequestURI)

		urlItem := []string{}
		si, _ := g.Cfg().Get(c.Request.Context(), "secure.ignore")
		for _, urlStr := range si.Interfaces() {
			urlItem = append(urlItem, urlStr.(string))
		}

		if !gstr.InArray(urlItem, u.Path) {
			if service.Login().IsSignedIn(c.Request.Context()) {
				if strings.HasPrefix(u.Path, "/manage") {
					ctxUser := service.BizCtx().GetUser(c.Request.Context())
					allowed := ctxUser != nil && (ctxUser.RoleId == consts.ROLE_SELLER ||
						ctxUser.RoleId == consts.ROLE_CHAIN ||
						ctxUser.RoleId == consts.ROLE_SITE ||
						ctxUser.RoleId == consts.ROLE_ADMIN)
					if !allowed {
						writeJSONExit(c, model.DefaultHandlerResponse{
							Code:    http.StatusForbidden,
							Message: "无权限访问管理端接口",
							Data:    nil,
							Status:  250,
						})
						return
					}

					if _, ok := managePermissionBypassPaths[u.Path]; ok {
						c.Next()
						return
					}

					candidatePaths := []string{u.Path}
					if aliases, ok := managePermissionAliases[u.Path]; ok && len(aliases) > 0 {
						candidatePaths = append(candidatePaths, aliases...)
					}

					hasPermission := false
					for _, menu := range ctxUser.Authorities {
						if menu == nil || menu.MenuPermission == "" {
							continue
						}
						permissions := strings.Split(menu.MenuPermission, ",")
						for _, permission := range permissions {
							permission = strings.TrimSpace(permission)
							if permission == "" {
								continue
							}
							for _, candidate := range candidatePaths {
								if permission == candidate {
									hasPermission = true
									break
								}
								if strings.HasSuffix(permission, "/*") {
									prefix := strings.TrimSuffix(permission, "/*")
									if strings.HasPrefix(candidate, prefix+"/") || candidate == prefix {
										hasPermission = true
										break
									}
								}
							}
							if hasPermission {
								break
							}
						}
						if hasPermission {
							break
						}
					}
					if !hasPermission {
						if ctxUser.RoleId == consts.ROLE_SELLER ||
							ctxUser.RoleId == consts.ROLE_CHAIN ||
							ctxUser.RoleId == consts.ROLE_SITE ||
							ctxUser.RoleId == consts.ROLE_ADMIN {
							configured, e := managePermissionConfigured(c.Request.Context(), u.Path)
							if e == nil && !configured {
								c.Next()
								return
							}
						}

						writeJSONExit(c, model.DefaultHandlerResponse{
							Code:    http.StatusForbidden,
							Message: "无权限访问该接口",
							Data:    nil,
							Status:  250,
						})
						return
					}
				}
				c.Next()
			} else {
				writeJSONExit(c, model.DefaultHandlerResponse{
					Code:    http.StatusUnauthorized,
					Message: "需要登录",
					Data:    nil,
					Status:  250,
				})
			}
		} else {
			c.Next()
		}
	}
}

func managePermissionConfigured(ctx context.Context, path string) (bool, error) {
	query := &do.MenuBaseListInput{}
	query.WhereExt = []*ml.WhereExt{{
		Column: dao.MenuBase.Columns().MenuPermission,
		Val:    "%" + path + "%",
		Symbol: ml.LIKE,
	}}
	rows, err := dao.MenuBase.Find(ctx, query)
	if err != nil {
		return false, err
	}
	if g.IsEmpty(rows) {
		return false, nil
	}
	for _, row := range rows {
		if row == nil || row.MenuPermission == "" {
			continue
		}
		permissions := strings.Split(row.MenuPermission, ",")
		for _, permission := range permissions {
			permission = strings.TrimSpace(permission)
			if permission == "" {
				continue
			}
			if permission == path {
				return true, nil
			}
			if strings.HasSuffix(permission, "/*") {
				prefix := strings.TrimSuffix(permission, "/*")
				if strings.HasPrefix(path, prefix+"/") || path == prefix {
					return true, nil
				}
			}
		}
	}
	return false, nil
}

// MiddlewareHandlerResponse / MiddlewareErrorHandler：Gin 由 binder 写 JSON，cmd 中未挂到 Gin 路由；此处不导出工厂，避免误用。
// 若需对齐 ghttp 全局链，可参考 logic 同名方法及 errorToString 实现后再注册。
