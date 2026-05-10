package consts

// 分销模式
const (
	DISTRIBUTION_MODE_ALL     = 1 // 全民分销
	DISTRIBUTION_MODE_SPECIFY = 2 // 指定分销
)

// 分销绑定对象
const (
	DISTRIBUTION_BIND_ALL = 1 // 所有用户可绑
	DISTRIBUTION_BIND_NEW = 2 // 仅新用户
)

// 分销绑定模式
const (
	DISTRIBUTION_BIND_MODE_FOREVER    = 1 // 永久
	DISTRIBUTION_BIND_MODE_PERIOD     = 2 // 有效期（到期清理）
	DISTRIBUTION_BIND_MODE_TEMPORARY  = 3 // 临时（可换绑）
)
