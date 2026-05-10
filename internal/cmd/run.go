package cmd

import (
	"context"
	"path/filepath"

	"golershop.cn/internal/config"
	_ "golershop.cn/internal/logic"
	"golershop.cn/internal/router"
	"golershop.cn/internal/service"
	"golershop.cn/internal/system"
)

func Run() error {
	cfgPath := filepath.Join("configs", "config.yaml")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return err
	}
	config.SetCurrent(cfg)

	// 与 golershop internal/cmd/cmd.go 一致：启动 HTTP 前执行 ConfigBase.Init（其中写入 global.Debug/Cache/Namespace/Separator/BaseUrl/UrlH5/UrlPc 等；global.Lk 在 SetCurrent 中由 yaml licence 写入）
	if _, err := service.ConfigBase().Init(context.Background()); err != nil {
		return err
	}

	bootstrap := system.NewBootstrap()
	bootstrap.Init()
	defer bootstrap.Stop()

	r := router.New(cfg)
	return r.Run(cfg.Server.Address)
}
