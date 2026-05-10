/**
 * Jwt工具类
 * @author Xinze
 * @since 2021/1/12
 * @File : jwt
 */
package utility

import (
	"fmt"
	"math"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gtime"
)

// 获取客户端IP
func GetClientIp(r *ghttp.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.GetClientIp()
	}
	return ip
}

// GetClientIpHTTP 与 GetClientIp 语义对齐，供 Gin 等 net/http 场景（来源：Java 迁移侧访问日志与 OperateLog）。
func GetClientIpHTTP(r *http.Request) string {
	if r == nil {
		return ""
	}
	ip := strings.TrimSpace(r.Header.Get("X-Forwarded-For"))
	if ip != "" {
		if i := strings.IndexByte(ip, ','); i > 0 {
			ip = strings.TrimSpace(ip[:i])
		}
		return ip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// RequestParamMap 合并 URL query 与 PostForm，近似 ghttp.Request.GetMap()，供 Gin Hook 与 OperateLog/OperateAccess 对齐。
func RequestParamMap(r *http.Request) g.Map {
	out := g.Map{}
	if r == nil {
		return out
	}
	for k, vs := range r.URL.Query() {
		if len(vs) == 1 {
			out[k] = vs[0]
		} else {
			out[k] = vs
		}
	}
	if r.Form != nil {
		for k, vs := range r.Form {
			if len(vs) == 1 {
				out[k] = vs[0]
			} else {
				out[k] = vs
			}
		}
	}
	return out
}

func InStringArray(value string, array []string) bool {
	for _, v := range array {
		if v == value {
			return true
		}
	}
	return false
}

// 数组反转
func Reverse(arr *[]string) {
	length := len(*arr)
	var temp string
	for i := 0; i < length/2; i++ {
		temp = (*arr)[i]
		(*arr)[i] = (*arr)[length-1-i]
		(*arr)[length-1-i] = temp
	}
}

// GetPlantformUid 获取平台用户Id
func GetPlantformUid(serviceUserId string, userId string) int {
	puid := fmt.Sprintf("%s-%s", serviceUserId, userId)
	return BkdrHash(puid)
}

// BkdrHash 计算BKDR哈希值
func BkdrHash(str string) int {
	seed := 131
	var hash int

	for _, char := range str {
		hash = seed*hash + int(char)
	}

	return hash & 0x7FFFFFFF
}

// 生成4位随机数字字符串
func GenerateRandomNumbers(n int) string {
	rand.Seed(time.Now().UnixNano())
	min := int64(1 * int(math.Pow10(n-1))) // 1000 for 4 digits
	max := int64(1*int(math.Pow10(n)) - 1) // 9999 for 4 digits
	return strconv.FormatInt(rand.Int63n(max-min+1)+min, 10)
}

func IsEmpty(value interface{}, traceSource ...bool) bool {
	return g.IsEmpty(value, traceSource...)
}

func IsNotEmpty(value interface{}, traceSource ...bool) bool {
	return !g.IsEmpty(value, traceSource...)
}

// 判断是否为数据库中的零值
func IsZeroTime(t *gtime.Time) bool {
	if t == nil {
		return true
	}

	// 检查是否是Go的零值时间
	if t.IsZero() {
		return true
	}

	// 检查格式化后是否是 0000-00-00 或 0001-01-01
	formatted := t.Format("Y-m-d")
	return formatted == "0000-00-00" || formatted == "0001-01-01"
}
