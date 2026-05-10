package log

import (
	"context"
	"fmt"
	"runtime"

	"github.com/gogf/gf/v2/frame/g"
	"golershop.cn/internal/consts"
	"golershop.cn/internal/dao/global"
	"golershop.cn/internal/service"
)

// Print 信息
func Print(ctx context.Context, v ...interface{}) {
	g.Log().Print(ctx, v)
}

// Info 信息
func Info(ctx context.Context, v ...interface{}) {
	g.Log().Info(ctx, v)
}

// Debug 信息
func Debug(ctx context.Context, v ...interface{}) {
	g.Log().Debug(ctx, v)
}

// Warning 信息
func Warning(ctx context.Context, v ...interface{}) {
	g.Log().Warning(ctx, v)
}

// Error 错误
func Error(ctx context.Context, v ...interface{}) {
	g.Log().Error(ctx, v)

	_, file, line, _ := runtime.Caller(1) // 1 表示上一层的调用者
	errStr := fmt.Sprintf("%s:%d - %v", file, line, v)

	//errStr := fmt.Sprintf("%v\n%v", v, g.Log().GetStack(1)) // 使用 %v 格式化动词

	service.LogError().Error(ctx, errStr, consts.ERR_NOT_DEFINITION)
}

func Panic(message any) {
	if global.Debug {
		_, file, line, _ := runtime.Caller(1) // 1 表示上一层的调用者
		panic(fmt.Sprintf("%s:%d - %v", file, line, message))
	} else {
		panic(message)
	}
}
