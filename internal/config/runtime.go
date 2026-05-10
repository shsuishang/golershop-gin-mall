package config

var current *AppConfig

// SetCurrent 保存应用配置，并写入 Init 链路依赖的 global（对齐 goframe：业务 global 在 ConfigBase.Init 内由 utility+g.Cfg 填充；授权 key 仅 manifest/yaml 提供，在此写入 global.Lk）。
func SetCurrent(cfg *AppConfig) {
	current = cfg
}

func Current() *AppConfig {
	return current
}
