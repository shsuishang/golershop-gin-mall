// Package gfpreboot 完成：1) 注册 GoFrame MySQL/Redis 驱动；2) 将 g.Cfg() 绑定到含 database/redis 的 config.yaml。
// 由 internal/service 全局 blank import 一次即可。
//
// 注意：gcfg.NewAdapterFile 的首参是「默认配置文件名」而非目录；目录须通过 SetPath/AddPath 设置（曾误传 configs 目录会导致 g.Redis 读不到配置）。
package gfpreboot

import (
	"os"
	"path/filepath"
	"strings"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2" // 与 golershop main 一致，否则 g.DB 报未注册 mysql 驱动
	_ "github.com/gogf/gf/contrib/nosql/redis/v2"   // 与 golershop main 一致，否则 g.Redis 无法建连
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"
)

const configFile = "config.yaml"

// resolveConfigDir 返回含 config.yaml 的配置目录（一般为模块根下 configs）。
// 顺序：GOLERSHOP_GFCFG → 自 Getwd 向上找 configs/config.yaml → 自可执行文件目录向上找。
func resolveConfigDir() string {
	if v := strings.TrimSpace(os.Getenv("GOLERSHOP_GFCFG")); v != "" {
		v = filepath.Clean(v)
		if st, err := os.Stat(v); err == nil && !st.IsDir() {
			return filepath.Dir(v)
		}
		return v
	}
	if wd, err := os.Getwd(); err == nil {
		if d := findConfigsDirUpward(wd); d != "" {
			return d
		}
	}
	if exe, err := os.Executable(); err == nil {
		exe = filepath.Clean(exe)
		if d := findConfigsDirUpward(filepath.Dir(exe)); d != "" {
			return d
		}
	}
	return ""
}

func findConfigsDirUpward(start string) string {
	start = filepath.Clean(start)
	for i := 0; i < 20; i++ {
		parent := filepath.Dir(start)
		if start == parent {
			break
		}
		cfgPath := filepath.Join(start, "configs", configFile)
		if st, err := os.Stat(cfgPath); err == nil && !st.IsDir() {
			return filepath.Join(start, "configs")
		}
		start = parent
	}
	return ""
}

func init() {
	dir := resolveConfigDir()
	adapter, err := gcfg.NewAdapterFile()
	if err != nil {
		panic("golershop-gin gfpreboot: NewAdapterFile: " + err.Error())
	}
	if dir != "" {
		if err := adapter.SetPath(dir); err != nil {
			panic("golershop-gin gfpreboot: SetPath(" + dir + "): " + err.Error() +
				"（可设置环境变量 GOLERSHOP_GFCFG 为 configs 目录或 config.yaml 绝对路径）")
		}
	}
	adapter.SetFileName(configFile)
	g.Cfg().SetAdapter(adapter)
}
