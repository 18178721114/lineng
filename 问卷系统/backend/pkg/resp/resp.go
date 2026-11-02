// Package resp 提供统一的 HTTP 响应辅助函数。
package resp

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// BadRequest 返回 400 业务错误。
func BadRequest(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, gin.H{"error": msg})
}

// Error 返回指定状态码的业务错误。
func Error(c *gin.Context, code int, msg string) {
	c.JSON(code, gin.H{"error": msg})
}
