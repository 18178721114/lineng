// Package handler 实现各业务模块的 HTTP 处理器。
package handler

import (
	"gorm.io/gorm"
)

// Handler 持有全局依赖（数据库），所有 HTTP 处理器均作为其方法。
type Handler struct {
	db       *gorm.DB
	h5Secret string // H5 分享链接签名密钥（与前端 share.js / h5/index.html 一致）
}

// New 创建 Handler。
func New(db *gorm.DB, h5Secret string) *Handler {
	return &Handler{db: db, h5Secret: h5Secret}
}
