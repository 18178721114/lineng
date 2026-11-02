// Command importworkers 将《项目报销记录(1)_员工表.xlsx》的人员数据导入 lineng.workers 表。
//
// 用法（在 backend 目录下执行）：
//
//	go run ./cmd/importworkers                       # 默认读取 ../项目报销记录(1)_员工表.xlsx
//	go run ./cmd/importworkers /path/to/file.xlsx    # 指定文件
//
// 映射关系：
//
//	Excel 列         -> workers 字段
//	状态(在职/离职)  -> active（离职=0，其余=1）
//	姓名            -> name（"同上" 沿用上一行姓名；空则跳过）
//	公司            -> company（"---" 视为空）
//	身份证          -> id_card（空则空串）
//	父记录          -> 忽略（源数据全为空）
//	pinyin_full/initial -> 由姓名实时计算（与线上 toPinyin 一致）
//
// 采用 INSERT ... ON DUPLICATE KEY UPDATE（按 name 唯一键），重复执行可安全更新。
package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"unicode"

	"github.com/mozillazg/go-pinyin"
	"github.com/xuri/excelize/v2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"

	"lineng/configs"
	"lineng/internal/model"
)

// toPinyin 与 internal/handler/helper.go 中的实现保持一致：中文转全拼+首字母。
func toPinyin(name string) (full, initial string) {
	a := pinyin.NewArgs()
	a.Style = pinyin.Normal
	for _, r := range name {
		s := string(r)
		py := pinyin.Pinyin(s, a)
		if len(py) > 0 && len(py[0]) > 0 && py[0][0] != "" && py[0][0] != s {
			full += py[0][0]
			initial += py[0][0][:1]
		} else if unicode.IsLetter(r) || unicode.IsDigit(r) {
			l := unicode.ToLower(r)
			full += string(l)
			initial += string(l)
		}
	}
	return full, initial
}

func cell(row []string, i int) string {
	if i >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[i])
}

func clean(v string) string {
	v = strings.TrimSpace(v)
	if v == "---" || v == "同上" {
		return ""
	}
	return v
}

func main() {
	cfg := configs.Load()

	db, err := gorm.Open(mysql.Open(cfg.DSN(true)), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	if err := db.AutoMigrate(&model.Worker{}); err != nil {
		log.Fatalf("自动建表/迁移失败: %v", err)
	}

	path := "../项目报销记录(1)_员工表.xlsx"
	if len(os.Args) > 1 {
		path = os.Args[1]
	}
	f, err := excelize.OpenFile(path)
	if err != nil {
		log.Fatalf("打开 Excel 失败(%s): %v", path, err)
	}
	defer f.Close()

	sheet := f.GetSheetName(0)
	rows, err := f.GetRows(sheet)
	if err != nil {
		log.Fatalf("读取行失败: %v", err)
	}

	var prev string
	inserted, updated, skipped := 0, 0, 0

	for i, row := range rows {
		if i == 0 { // 表头
			continue
		}
		status := cell(row, 0)
		name := cell(row, 1)
		company := cell(row, 2)
		idCard := cell(row, 3)
		// 父记录(row[4]) 全为空，忽略

		if name == "同上" {
			name = prev
		}
		name = strings.TrimSpace(name)
		if name == "" {
			skipped++
			continue
		}
		prev = name

		active := true
		if status == "离职" {
			active = false
		}
		company = clean(company)
		idCard = clean(idCard)
		full, ini := toPinyin(name)

		w := model.Worker{
			Name:          name,
			Phone:         "",
			JobType:       "",
			Note:          "",
			Company:       company,
			IDCard:        idCard,
			PinyinFull:    full,
			PinyinInitial: ini,
			Active:        active,
		}

		res := db.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "name"}},
			DoUpdates: clause.Assignments(map[string]any{
				"company":         w.Company,
				"id_card":         w.IDCard,
				"active":          w.Active,
				"pinyin_full":     w.PinyinFull,
				"pinyin_initial":  w.PinyinInitial,
			}),
		}).Create(&w)

		if res.Error != nil {
			log.Printf("行 %d [%s] 写入失败: %v", i+1, name, res.Error)
			skipped++
			continue
		}
		if res.RowsAffected == 1 {
			inserted++
		} else {
			updated++
		}
	}

	fmt.Printf("导入完成：新增 %d，更新 %d，跳过 %d（文件: %s）\n", inserted, updated, skipped, path)
}
