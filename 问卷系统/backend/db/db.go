// Package db 负责数据库连接与初始化。
package db

import (
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"lineng/configs"
	"lineng/internal/model"
)

// Init 连接 MySQL，确保数据库存在并自动迁移表结构。
func Init(cfg *configs.Config) *gorm.DB {
	silent := &gorm.Config{Logger: logger.Default.LogMode(logger.Warn)}

	// 先无库连接，创建数据库
	tmp, err := gorm.Open(mysql.Open(cfg.DSN(false)), silent)
	if err != nil {
		log.Fatalf("连接 MySQL 失败(请确认 %s:%s %s 已启动): %v", cfg.DBHost, cfg.DBPort, cfg.DBUser, err)
	}
	if err := tmp.Exec("CREATE DATABASE IF NOT EXISTS " + cfg.DBName + " DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci").Error; err != nil {
		log.Fatalf("创建数据库失败: %v", err)
	}
	if sqlDB, err := tmp.DB(); err == nil {
		_ = sqlDB.Close()
	}

	database, err := gorm.Open(mysql.Open(cfg.DSN(true)), silent)
	if err != nil {
		log.Fatalf("连接库 %s 失败: %v", cfg.DBName, err)
	}
	// 按依赖顺序迁移：主数据表 → 日报主表 → 各业务模块子表
	if err := database.AutoMigrate(
		&model.User{},
		&model.Project{},
		&model.ProjectPrices{},
		&model.Worker{},
		&model.DailyReport{},
		&model.ReportWorker{},
		&model.ReportShift{},
		&model.ReportMeal{},
		&model.ReportEquipment{},
		&model.ReportOutsourcing{},
		&model.ReportModuleDraft{},
	); err != nil {
		log.Fatalf("自动建表失败: %v", err)
	}
	return database
}
