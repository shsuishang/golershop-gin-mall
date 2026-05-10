package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golershop.cn/internal/model"
)

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, model.DefaultHandlerResponse{
		Status:  200,
		Code:    0,
		Message: "操作成功",
		Data:    data,
	})
}

func Fail(c *gin.Context, httpStatus int, code int, msg string) {
	if httpStatus == 0 {
		httpStatus = http.StatusOK
	}
	c.JSON(httpStatus, model.DefaultHandlerResponse{
		Status:  250,
		Code:    code,
		Message: msg,
		Data:    nil,
	})
}
