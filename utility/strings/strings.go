package strings

import (
	"context"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"strings"
)

func IsEmail(ctx context.Context, email string) (bool, error) {
	// 验证email格式
	if err := g.Validator().Rules("email").Data(email).Run(ctx); err != nil {
		return false, gerror.New("Email不准确！")
	}

	return true, nil
}

func AddSlashes(str string) string {
	var builder strings.Builder
	for _, char := range str {
		switch char {
		case '\'', '"', '\\':
			builder.WriteRune('\\')
			builder.WriteRune(char)
		default:
			builder.WriteRune(char)
		}
	}

	return builder.String()
}

func SubString[T ~string](s T, start, length int) T {
	// 转换为普通字符串处理
	str := string(s)

	// 转换为rune处理多字节字符
	runes := []rune(str)

	// 处理边界
	if start < 0 {
		start = 0
	}
	if start >= len(runes) {
		return T("")
	}

	end := start + length
	if end > len(runes) {
		end = len(runes)
	}

	return T(string(runes[start:end]))
}
