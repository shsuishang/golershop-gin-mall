package consts

const (
	UserSessionKey = "UserSessionKey"
)

const (
	USER_CREDIT_STATUS_PENDING  uint = 1 // 未开通
	USER_CREDIT_STATUS_OPEN     uint = 2 // 已开通（通过）
	USER_CREDIT_STATUS_FROZEN   uint = 3 // 冻结
	USER_CREDIT_STATUS_CLOSED   uint = 4 // 关闭
	USER_CREDIT_STATUS_REJECTED uint = 5 // 驳回
)
