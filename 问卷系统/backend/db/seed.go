package db

import (
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"lineng/internal/model"
)

// SeedAdmin 首次启动时创建默认管理员 admin / admin123。
func SeedAdmin(database *gorm.DB) {
	var n int64
	database.Model(&model.User{}).Count(&n)
	if n == 0 {
		hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		database.Create(&model.User{Username: "admin", Password: string(hash)})
		log.Println("[lineng] 已创建默认管理员 admin / admin123")
	}
}
