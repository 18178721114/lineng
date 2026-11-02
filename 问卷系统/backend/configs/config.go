// Package configs 负责从环境变量加载运行配置（含默认值）。
package configs

import (
	"os"
)

// Config 应用运行配置。
type Config struct {
	DBHost string
	DBPort string
	DBUser string
	DBPass string
	DBName string

	Addr string // HTTP 监听地址，如 :8090

	H5Secret string // H5 分享链接签名密钥
}

// Load 读取环境变量，未设置时回退到本地开发默认值。
func Load() *Config {
	return &Config{
		DBHost:   env("DB_HOST", "127.0.0.1"),
		DBPort:   env("DB_PORT", "3306"),
		DBUser:   env("DB_USER", "lineng"),
		DBPass:   env("DB_PASS", "zjyh7EC6ADnZL2ya"),
		DBName:   env("DB_NAME", "lineng"),
		Addr:     env("SERVER_ADDR", ":8090"),
		H5Secret: env("H5_SHARE_SECRET", "LILIANG_SURVEY_SECRET_2026_CHANGE_ME"),
	}
}

// DSN 生成 MySQL 连接串；withDB 为 false 时不带库名（用于建库）。
func (c *Config) DSN(withDB bool) string {
	s := c.DBUser + ":" + c.DBPass + "@tcp(" + c.DBHost + ":" + c.DBPort + ")/"
	if withDB {
		s += c.DBName
	}
	return s + "?charset=utf8mb4&parseTime=True&loc=Local"
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
