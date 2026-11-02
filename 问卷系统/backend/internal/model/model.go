// Package model 定义数据库表模型与前端读写视图结构。
// 设计规范：一个业务模块一张表，不使用 JSON / TEXT 存储结构化数据。
package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// Date 以 YYYY-MM-DD 承载 DATE 列。
// DSN 开启 parseTime 后，驱动会把 DATE 扫成 time.Time 再转成 RFC3339 字符串，
// 因此这里自定义类型，保证 Go 侧与 API 层始终是 YYYY-MM-DD。
type Date string

// Scan 实现 sql.Scanner。
func (d *Date) Scan(v any) error {
	switch x := v.(type) {
	case nil:
		*d = ""
	case time.Time:
		*d = Date(x.Format("2006-01-02"))
	case []byte:
		s := string(x)
		if len(s) > 10 {
			s = s[:10]
		}
		*d = Date(s)
	case string:
		s := x
		if len(s) > 10 {
			s = s[:10]
		}
		*d = Date(s)
	default:
		return fmt.Errorf("不支持的 DATE 值: %T", v)
	}
	return nil
}

// Value 实现 driver.Valuer。
func (d Date) Value() (driver.Value, error) {
	if d == "" {
		return nil, nil
	}
	return string(d), nil
}

// Clock 以 HH:MM 承载 TIME 列（同上，避免扫成 15:04:05 以外的形式）。
type Clock string

// Scan 实现 sql.Scanner。
func (c *Clock) Scan(v any) error {
	switch x := v.(type) {
	case nil:
		*c = ""
	case time.Time:
		*c = Clock(x.Format("15:04"))
	case []byte:
		s := string(x)
		if len(s) > 5 {
			s = s[:5]
		}
		*c = Clock(s)
	case string:
		s := x
		if len(s) > 5 {
			s = s[:5]
		}
		*c = Clock(s)
	default:
		return fmt.Errorf("不支持的 TIME 值: %T", v)
	}
	return nil
}

// Value 实现 driver.Valuer。
func (c Clock) Value() (driver.Value, error) {
	if c == "" {
		return nil, nil
	}
	return string(c), nil
}

// 设备类型的存储值：叉车 3/5/7 吨、吊车、救援车。
const (
	EquipFork3  = "FORK3"
	EquipFork5  = "FORK5"
	EquipFork7  = "FORK7"
	EquipCrane  = "CRANE"
	EquipRescue = "RESCUE"
)

// 班次的存储值：白天 / 晚上 / 凌晨。
const (
	ShiftDay     = "DAY"
	ShiftEvening = "EVENING"
	ShiftDawn    = "DAWN"
)

// EquipDef 设备定义：数据库存储值 ↔ 前端使用的键。
type EquipDef struct {
	Code string
	Key  string
}

// ShiftDef 班次定义：数据库存储值 ↔ 前端使用的键。
type ShiftDef struct {
	Code string
	Key  string
}

// EquipDefs 设备清单（顺序与结算单打印顺序一致）。
var EquipDefs = []EquipDef{
	{EquipFork3, "fork3"},
	{EquipFork5, "fork5"},
	{EquipFork7, "fork7"},
	{EquipCrane, "crane"},
	{EquipRescue, "rescue"},
}

// ShiftDefs 班次清单。
var ShiftDefs = []ShiftDef{
	{ShiftDay, "day"},
	{ShiftEvening, "eve"},
	{ShiftDawn, "dawn"},
}

// ---------------- 主数据表 ----------------

// User 管理员账号。
type User struct {
	ID        uint      `gorm:"primaryKey;comment:主键 ID" json:"id"`
	Username  string    `gorm:"size:64;uniqueIndex;comment:登录名" json:"username"`
	Password  string    `gorm:"comment:登录密码（bcrypt 哈希）" json:"-"`
	CreatedAt time.Time `gorm:"comment:创建时间" json:"createdAt"`
}

// Project 对应纸质《出勤表》表头字段。
type Project struct {
	ID           uint   `gorm:"primaryKey;comment:主键 ID" json:"id"`
	No           string `gorm:"size:32;uniqueIndex;comment:单据编号（LN-YYYYMMDD-XXX）" json:"no"`
	Unit         string `gorm:"size:128;comment:用工单位" json:"unit"`
	ActivityName string `gorm:"size:128;comment:活动名称" json:"activityName"`
	Location     string `gorm:"size:128;comment:用工地点" json:"location"`
	PartyAOwner  string `gorm:"size:64;comment:甲方负责人" json:"partyAOwner"`
	PartyAPhone  string `gorm:"size:32;comment:甲方联系电话" json:"partyAPhone"`
	StartDate    Date   `gorm:"type:date;comment:用工开始日期" json:"startDate"`
	EndDate      Date   `gorm:"type:date;comment:用工结束日期" json:"endDate"`
	// Prices 不落库：真实单价存于 project_prices 表，读取时序列化填充，
	// 仅用于兼容前端导出页读取项目默认单价（JSON 字符串）。
	Prices    string    `gorm:"-" json:"prices"`
	CreatedAt time.Time `gorm:"comment:创建时间" json:"createdAt"`
	UpdatedAt time.Time `gorm:"comment:更新时间" json:"updatedAt"`
}

// ProjectPrices 项目默认单价，与 projects 一对一。
type ProjectPrices struct {
	ProjectID   uint      `gorm:"primaryKey;constraint:OnDelete:CASCADE;comment:项目 ID（主键，与 projects.id 对应）" json:"projectId"`
	LaborPrice  float64   `gorm:"type:decimal(10,2);not null;default:0;comment:人工单价，单位 元/人" json:"laborPrice"`
	MealPrice   float64   `gorm:"type:decimal(10,2);not null;default:0;comment:餐费单价，单位 元/份" json:"mealPrice"`
	WaterPrice  float64   `gorm:"type:decimal(10,2);not null;default:0;comment:水单价，单位 元/份" json:"waterPrice"`
	Fork3Price  float64   `gorm:"type:decimal(10,2);not null;default:0;comment:叉车 3 吨单价，单位 元/台班" json:"fork3Price"`
	Fork5Price  float64   `gorm:"type:decimal(10,2);not null;default:0;comment:叉车 5 吨单价，单位 元/台班" json:"fork5Price"`
	Fork7Price  float64   `gorm:"type:decimal(10,2);not null;default:0;comment:叉车 7 吨单价，单位 元/台班" json:"fork7Price"`
	CranePrice  float64   `gorm:"type:decimal(10,2);not null;default:0;comment:吊车单价，单位 元/台班" json:"cranePrice"`
	RescuePrice float64   `gorm:"type:decimal(10,2);not null;default:0;comment:救援车单价，单位 元/台班" json:"rescuePrice"`
	TaxRate     float64   `gorm:"type:decimal(5,2);not null;default:6;comment:税率百分比，6 表示 6%" json:"taxRate"`
	UpdatedAt   time.Time `gorm:"comment:更新时间" json:"updatedAt"`
}

// DefaultProjectPrices 返回一套系统默认单价。
func DefaultProjectPrices(projectID uint) ProjectPrices {
	return ProjectPrices{
		ProjectID:   projectID,
		LaborPrice:  380,
		MealPrice:   20,
		WaterPrice:  5,
		Fork3Price:  1200,
		Fork5Price:  1600,
		Fork7Price:  2000,
		CranePrice:  1200,
		RescuePrice: 800,
		TaxRate:     6,
	}
}

// JSON 序列化为前端使用的单价对象（键名与前端一致）。
func (p ProjectPrices) JSON() string {
	m := map[string]float64{
		"labor":   p.LaborPrice,
		"meal":    p.MealPrice,
		"water":   p.WaterPrice,
		"fork3":   p.Fork3Price,
		"fork5":   p.Fork5Price,
		"fork7":   p.Fork7Price,
		"crane":   p.CranePrice,
		"rescue":  p.RescuePrice,
		"taxRate": p.TaxRate,
	}
	b, _ := json.Marshal(m)
	return string(b)
}

// Worker 人员档案。
type Worker struct {
	ID            uint      `gorm:"primaryKey;comment:主键 ID" json:"id"`
	Name          string    `gorm:"size:64;uniqueIndex;comment:姓名（唯一）" json:"name"`
	Phone         string    `gorm:"size:32;comment:联系电话" json:"phone"`
	JobType       string    `gorm:"size:64;comment:工种" json:"jobType"`
	Note          string    `gorm:"size:255;comment:备注" json:"note"`
	Company       string    `gorm:"size:64;comment:所属公司（员工表导入）" json:"company"`
	IDCard        string    `gorm:"size:32;comment:身份证号（员工表导入）" json:"idCard"`
	PinyinFull    string    `gorm:"size:255;index;comment:姓名全拼，小写无空格，用于搜索" json:"pinyinFull"`
	PinyinInitial string    `gorm:"size:64;index;comment:姓名拼音首字母，用于搜索" json:"pinyinInitial"`
	Active        bool      `gorm:"default:true;comment:是否在职：1 在职，0 停用" json:"active"`
	CreatedAt     time.Time `gorm:"comment:创建时间" json:"createdAt"`
}

// DailyReport 日报主表：一项目一天一单（保存即 upsert）。
// 五个业务模块的数据分别存放于下列子表，主表只保留标量与标记。
type DailyReport struct {
	ID           uint      `gorm:"primaryKey;comment:主键 ID" json:"id"`
	ProjectID    uint      `gorm:"uniqueIndex:uk_proj_date;not null;constraint:OnDelete:CASCADE;comment:所属项目 ID" json:"projectId"`
	ReportDate   Date      `gorm:"type:date;uniqueIndex:uk_proj_date;not null;comment:出勤日期" json:"reportDate"`
	IsOutsourced bool      `gorm:"not null;default:false;comment:是否外包：1 是，0 否" json:"isOutsourced"`
	Notes        string    `gorm:"size:500;comment:备注" json:"notes"`
	CreatedAt    time.Time `gorm:"comment:创建时间" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"comment:更新时间" json:"updatedAt"`
}

// ---------------- 业务模块表（一模块一表） ----------------

// ReportWorker ① 用工名单：先填工作时间，再选人名（可多选）。
// 一次勾选多人 = 多行共享同一 start_time/end_time。
type ReportWorker struct {
	ID         uint   `gorm:"primaryKey;comment:主键 ID" json:"id"`
	ReportID   uint   `gorm:"not null;index;constraint:OnDelete:CASCADE;comment:所属日报 ID" json:"reportId"`
	WorkerID   uint   `gorm:"not null;index;comment:人员 ID（关联 workers.id）" json:"workerId"`
	WorkerName string `gorm:"size:64;not null;comment:下单时的姓名快照，人员改名不影响历史单据" json:"workerName"`
	// 说明：GORM 的 type:time 会被解析为时间类型枚举并建出 datetime(3)，
	// 这里用 time(0) 强制落地为 MySQL 的 TIME 列（秒精度为 0）。
	StartTime Clock `gorm:"type:time(0);not null;comment:上班时间 HH:MM" json:"startTime"`
	EndTime   Clock `gorm:"type:time(0);not null;comment:下班时间 HH:MM" json:"endTime"`
	Overnight bool  `gorm:"not null;default:false;comment:是否跨夜：1 表示下班时间在次日" json:"overnight"`
	SortOrder uint  `gorm:"not null;default:0;comment:名单内排序序号（决定考勤表行序）" json:"sortOrder"`
}

// ReportShift ② 用工人数：白天 / 晚上 / 凌晨（按名单自动带出，可改）。
type ReportShift struct {
	ReportID     uint `gorm:"primaryKey;constraint:OnDelete:CASCADE;comment:所属日报 ID（主键）" json:"reportId"`
	DayCount     int  `gorm:"not null;default:0;comment:白天人数" json:"day"`
	EveningCount int  `gorm:"not null;default:0;comment:晚上人数" json:"eve"`
	DawnCount    int  `gorm:"not null;default:0;comment:凌晨人数" json:"dawn"`
}

// ReportMeal ③ 餐费：早点 / 中餐 / 晚餐 / 夜宵 + 水。
type ReportMeal struct {
	ReportID     uint `gorm:"primaryKey;constraint:OnDelete:CASCADE;comment:所属日报 ID（主键）" json:"reportId"`
	BreakfastQty int  `gorm:"not null;default:0;comment:早点份数" json:"breakfast"`
	LunchQty     int  `gorm:"not null;default:0;comment:中餐份数" json:"lunch"`
	DinnerQty    int  `gorm:"not null;default:0;comment:晚餐份数" json:"dinner"`
	SupperQty    int  `gorm:"not null;default:0;comment:夜宵份数" json:"supper"`
	WaterQty     int  `gorm:"not null;default:0;comment:水份数" json:"water"`
}

// TableName 固定表名，避免 GORM 复数化后与 schema.sql 不一致。
func (ReportOutsourcing) TableName() string { return "report_outsourcing" }

// ReportEquipment ④ 设备台班：5 种设备 × 3 个班次，每单固定 15 行。
type ReportEquipment struct {
	ID        uint   `gorm:"primaryKey;comment:主键 ID" json:"id"`
	ReportID  uint   `gorm:"uniqueIndex:uk_res;not null;constraint:OnDelete:CASCADE;comment:所属日报 ID" json:"reportId"`
	EquipType string `gorm:"type:enum('FORK3','FORK5','FORK7','CRANE','RESCUE');uniqueIndex:uk_res;not null;comment:设备类型：FORK3 叉车3吨 / FORK5 叉车5吨 / FORK7 叉车7吨 / CRANE 吊车 / RESCUE 救援车" json:"equipType"`
	ShiftType string `gorm:"type:enum('DAY','EVENING','DAWN');uniqueIndex:uk_res;not null;comment:班次：DAY 白天 / EVENING 晚上 / DAWN 凌晨" json:"shiftType"`
	Qty       int    `gorm:"not null;default:0;comment:台班数" json:"qty"`
}

// TableName 固定表名，避免 GORM 复数化后与 schema.sql 不一致。
func (ReportEquipment) TableName() string { return "report_equipment" }

// ReportOutsourcing ⑤ 外包：一个外包单位一行，可加多套。
// 白班 / 晚班 / 凌晨 三个班次各自记录用工人数、加班人数与加班起止时间。
type ReportOutsourcing struct {
	ID           uint   `gorm:"primaryKey;comment:主键 ID" json:"id"`
	ReportID     uint   `gorm:"not null;index;constraint:OnDelete:CASCADE;comment:所属日报 ID" json:"reportId"`
	UnitName     string `gorm:"size:128;not null;comment:外包单位名称" json:"unitName"`
	DayCount     int    `gorm:"not null;default:0;comment:白天人数" json:"day"`
	EveningCount int    `gorm:"not null;default:0;comment:晚上人数" json:"eve"`
	DawnCount    int    `gorm:"not null;default:0;comment:凌晨人数" json:"dawn"`
	// 白班加班
	DayOtCount int    `gorm:"not null;default:0;comment:白班加班人数" json:"dayOt"`
	DayOtStart *Clock `gorm:"type:time(0);comment:白班加班开始时间，可空" json:"dayOtStart"`
	DayOtEnd   *Clock `gorm:"type:time(0);comment:白班加班结束时间，可空" json:"dayOtEnd"`
	// 晚班加班
	EveningOtCount int    `gorm:"not null;default:0;comment:晚班加班人数" json:"eveOt"`
	EveningOtStart *Clock `gorm:"type:time(0);comment:晚班加班开始时间，可空" json:"eveOtStart"`
	EveningOtEnd   *Clock `gorm:"type:time(0);comment:晚班加班结束时间，可空" json:"eveOtEnd"`
	// 凌晨加班
	DawnOtCount int    `gorm:"not null;default:0;comment:凌晨加班人数" json:"dawnOt"`
	DawnOtStart *Clock `gorm:"type:time(0);comment:凌晨加班开始时间，可空" json:"dawnOtStart"`
	DawnOtEnd   *Clock `gorm:"type:time(0);comment:凌晨加班结束时间，可空" json:"dawnOtEnd"`
	// 旧版通用加班时段，仅做历史数据兼容，新数据不再写入。
	OtStart   *Clock `gorm:"type:time(0);comment:旧版通用加班开始时间（历史兼容，不再写入）" json:"otStart"`
	OtEnd     *Clock `gorm:"type:time(0);comment:旧版通用加班结束时间（历史兼容，不再写入）" json:"otEnd"`
	Overnight bool   `gorm:"not null;default:false;comment:是否跨夜：1 表示加班结束时间在次日" json:"overnight"`
	SortOrder uint   `gorm:"not null;default:0;comment:排列顺序" json:"sortOrder"`
}

// ReportModuleDraft H5 模块暂存草稿：同项目同日期同模块只保留一条，
// 保存即 upsert（有则更新、无则新增），提交日报后整组删除。
// 各模块字段不同，内容以 JSON 对象整体存储（字段名 → 值）。
type ReportModuleDraft struct {
	ID         uint      `gorm:"primaryKey;comment:主键 ID" json:"id"`
	ProjectID  uint      `gorm:"uniqueIndex:uk_draft;not null;comment:所属项目 ID" json:"projectId"`
	ReportDate Date      `gorm:"type:date;uniqueIndex:uk_draft;not null;comment:出勤日期" json:"reportDate"`
	ModuleKey  string    `gorm:"size:32;uniqueIndex:uk_draft;not null;comment:模块键（names/labor/meal/fork/out/note）" json:"moduleKey"`
	Content    string    `gorm:"type:json;not null;comment:模块暂存内容（JSON 对象，字段名→值）" json:"content"`
	CreatedAt  time.Time `gorm:"comment:创建时间" json:"createdAt"`
	UpdatedAt  time.Time `gorm:"comment:更新时间" json:"updatedAt"`
}

// ---------------- 视图结构（API 契约，与前端保持一致） ----------------

// ReportDetail 前端读写的日报结构：五个模块子表组装后的视图。
type ReportDetail struct {
	ID          uint             `json:"id"`
	ProjectID   uint             `json:"projectId"`
	ReportDate  string           `json:"reportDate"`
	Workers     []map[string]any `json:"workers"`     // [{workerId,name,start,end}]
	Shifts      map[string]any   `json:"shifts"`      // {day,eve,dawn}
	Meals       map[string]any   `json:"meals"`       // {breakfast,lunch,dinner,supper,water}
	Equipment   map[string]any   `json:"equipment"`   // {fork3:{day,eve,dawn},...}
	Outsourced  bool             `json:"outsourced"`  // 是否外包
	Outsourcing []map[string]any `json:"outsourcing"` // [{name,day,eve,dawn,otStart,otEnd}]
	Notes       string           `json:"notes"`
	UpdatedAt   string           `json:"updatedAt"`
}

// EmptyReportDetail 空壳日报，供前端表单初始化。
func EmptyReportDetail(projectID uint, reportDate string) ReportDetail {
	equipment := map[string]any{}
	for _, e := range EquipDefs {
		shifts := map[string]any{}
		for _, s := range ShiftDefs {
			shifts[s.Key] = 0
		}
		equipment[e.Key] = shifts
	}
	return ReportDetail{
		ProjectID:   projectID,
		ReportDate:  reportDate,
		Workers:     []map[string]any{},
		Shifts:      map[string]any{"day": 0, "eve": 0, "dawn": 0},
		Meals:       map[string]any{"breakfast": 0, "lunch": 0, "dinner": 0, "supper": 0, "water": 0},
		Equipment:   equipment,
		Outsourcing: []map[string]any{},
	}
}

// EmptyEquipmentShifts 返回某设备三个班次均为 0 的结构。
func EmptyEquipmentShifts() map[string]any {
	m := map[string]any{}
	for _, s := range ShiftDefs {
		m[s.Key] = 0
	}
	return m
}
