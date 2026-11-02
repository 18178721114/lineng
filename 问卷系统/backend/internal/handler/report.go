package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"lineng/internal/model"
	"lineng/pkg/resp"
)

// reportIDs 提取日报主键，用于子表批量查询。
func reportIDs(rs []model.DailyReport) []uint {
	ids := make([]uint, 0, len(rs))
	for _, r := range rs {
		ids = append(ids, r.ID)
	}
	return ids
}

// loadDetails 把日报主表记录与五个模块子表组装成前端视图；子表各查一次，避免逐单 N+1。
func (h *Handler) loadDetails(rs []model.DailyReport) []model.ReportDetail {
	ids := reportIDs(rs)
	workersOf := map[uint][]map[string]any{}
	shiftsOf := map[uint]model.ReportShift{}
	mealsOf := map[uint]model.ReportMeal{}
	equipOf := map[uint]map[string]map[string]int{}
	outsOf := map[uint][]map[string]any{}

	if len(ids) > 0 {
		// ① 用工名单
		var ws []model.ReportWorker
		h.db.Where("report_id IN ?", ids).Order("report_id, sort_order, id").Find(&ws)
		for _, w := range ws {
			workersOf[w.ReportID] = append(workersOf[w.ReportID], map[string]any{
				"workerId": w.WorkerID,
				"name":     w.WorkerName,
				"start":    hhmm(string(w.StartTime)),
				"end":      hhmm(string(w.EndTime)),
			})
		}
		// ② 用工人数
		var ss []model.ReportShift
		h.db.Where("report_id IN ?", ids).Find(&ss)
		for _, s := range ss {
			shiftsOf[s.ReportID] = s
		}
		// ③ 餐费
		var ms []model.ReportMeal
		h.db.Where("report_id IN ?", ids).Find(&ms)
		for _, m := range ms {
			mealsOf[m.ReportID] = m
		}
		// ④ 设备台班
		var es []model.ReportEquipment
		h.db.Where("report_id IN ?", ids).Find(&es)
		for _, e := range es {
			if equipOf[e.ReportID] == nil {
				equipOf[e.ReportID] = map[string]map[string]int{}
			}
			if equipOf[e.ReportID][e.EquipType] == nil {
				equipOf[e.ReportID][e.EquipType] = map[string]int{}
			}
			equipOf[e.ReportID][e.EquipType][e.ShiftType] = e.Qty
		}
		// ⑤ 外包
		var os []model.ReportOutsourcing
		h.db.Where("report_id IN ?", ids).Order("report_id, sort_order, id").Find(&os)
		for _, o := range os {
			// 旧数据只有一条通用加班时段，回填到白班，避免历史记录丢失展示。
			dayOtStart, dayOtEnd := o.DayOtStart, o.DayOtEnd
			if dayOtStart == nil && dayOtEnd == nil && (o.OtStart != nil || o.OtEnd != nil) {
				dayOtStart, dayOtEnd = o.OtStart, o.OtEnd
			}
			outsOf[o.ReportID] = append(outsOf[o.ReportID], map[string]any{
				"name": o.UnitName,
				"day":  o.DayCount,
				"eve":  o.EveningCount,
				"dawn": o.DawnCount,
				// 各班次加班人数与加班时段
				"dayOt":       o.DayOtCount,
				"dayOtStart":  derefTime(dayOtStart),
				"dayOtEnd":    derefTime(dayOtEnd),
				"eveOt":       o.EveningOtCount,
				"eveOtStart":  derefTime(o.EveningOtStart),
				"eveOtEnd":    derefTime(o.EveningOtEnd),
				"dawnOt":      o.DawnOtCount,
				"dawnOtStart": derefTime(o.DawnOtStart),
				"dawnOtEnd":   derefTime(o.DawnOtEnd),
			})
		}
	}

	out := make([]model.ReportDetail, 0, len(rs))
	for _, r := range rs {
		d := model.EmptyReportDetail(r.ProjectID, string(r.ReportDate))
		d.ID = r.ID
		d.Notes = r.Notes
		d.Outsourced = r.IsOutsourced
		d.UpdatedAt = r.UpdatedAt.Format("2006-01-02 15:04")
		if v, ok := workersOf[r.ID]; ok {
			d.Workers = v
		}
		if s, ok := shiftsOf[r.ID]; ok {
			d.Shifts = map[string]any{"day": s.DayCount, "eve": s.EveningCount, "dawn": s.DawnCount}
		}
		if m, ok := mealsOf[r.ID]; ok {
			d.Meals = map[string]any{
				"breakfast": m.BreakfastQty, "lunch": m.LunchQty, "dinner": m.DinnerQty,
				"supper": m.SupperQty, "water": m.WaterQty,
			}
		}
		for _, ed := range model.EquipDefs {
			shifts, _ := d.Equipment[ed.Key].(map[string]any)
			for _, sd := range model.ShiftDefs {
				if q, ok := equipOf[r.ID][ed.Code][sd.Code]; ok {
					shifts[sd.Key] = q
				}
			}
		}
		if v, ok := outsOf[r.ID]; ok {
			d.Outsourcing = v
		}
		out = append(out, d)
	}
	return out
}

// ListReports 某项目的日报列表，可按日期区间过滤。
func (h *Handler) ListReports(c *gin.Context) {
	pid := c.Query("projectId")
	if pid == "" {
		resp.BadRequest(c, "缺少 projectId")
		return
	}
	var rs []model.DailyReport
	q := h.db.Where("project_id = ?", pid)
	if st, en := c.Query("start"), c.Query("end"); st != "" && en != "" {
		q = q.Where("report_date BETWEEN ? AND ?", st, en)
	}
	q.Order("report_date DESC").Find(&rs)
	c.JSON(http.StatusOK, h.loadDetails(rs))
}

// GetReport 取某天日报；无则返回空壳（供表单初始化）。
func (h *Handler) GetReport(c *gin.Context) {
	pid, date := c.Param("projectId"), c.Param("date")
	var r model.DailyReport
	err := h.db.Where("project_id = ? AND report_date = ?", pid, date).First(&r).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusOK, model.EmptyReportDetail(parseUint(pid), date))
		return
	}
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	c.JSON(http.StatusOK, h.loadDetails([]model.DailyReport{r})[0])
}

// saveReportDetail 同一事务内 upsert 主表 + 五个模块子表，返回保存后的视图。
// PC 端（UpsertReport）与 H5 分享链接（H5SubmitReport）共用，避免逻辑分叉。
func (h *Handler) saveReportDetail(in model.ReportDetail) (model.ReportDetail, error) {
	tx := h.db.Begin()
	if tx.Error != nil {
		return model.ReportDetail{}, tx.Error
	}
	committed := false
	defer func() {
		if !committed {
			tx.Rollback()
		}
	}()

	var r model.DailyReport
	err := tx.Where("project_id = ? AND report_date = ?", in.ProjectID, in.ReportDate).First(&r).Error
	switch {
	case err == nil:
	case errors.Is(err, gorm.ErrRecordNotFound):
		r = model.DailyReport{ProjectID: in.ProjectID, ReportDate: model.Date(in.ReportDate)}
	default:
		return model.ReportDetail{}, err
	}
	r.IsOutsourced = in.Outsourced
	r.Notes = in.Notes
	if err := tx.Save(&r).Error; err != nil {
		return model.ReportDetail{}, err
	}

	steps := []struct {
		name string
		run  func(*gorm.DB, uint) error
	}{
		{"用工名单", func(t *gorm.DB, id uint) error { return saveReportWorkers(t, id, in.Workers) }},
		{"用工人数", func(t *gorm.DB, id uint) error { return saveReportShift(t, id, in.Shifts) }},
		{"餐费", func(t *gorm.DB, id uint) error { return saveReportMeal(t, id, in.Meals) }},
		{"设备台班", func(t *gorm.DB, id uint) error { return saveReportEquipment(t, id, in.Equipment) }},
		{"外包", func(t *gorm.DB, id uint) error { return saveReportOutsourcing(t, id, in.Outsourcing) }},
	}
	for _, s := range steps {
		if err := s.run(tx, r.ID); err != nil {
			return model.ReportDetail{}, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return model.ReportDetail{}, err
	}
	committed = true

	return h.loadDetails([]model.DailyReport{r})[0], nil
}

// UpsertReport 同项目同日期只保留一单：主表 + 五个模块子表在同一事务内写入。
func (h *Handler) UpsertReport(c *gin.Context) {
	var in model.ReportDetail
	if err := c.ShouldBindJSON(&in); err != nil {
		resp.BadRequest(c, "参数错误")
		return
	}
	if in.ProjectID == 0 || in.ReportDate == "" {
		resp.BadRequest(c, "projectId 与 reportDate 必填")
		return
	}
	if len(in.ReportDate) != 10 || strings.Count(in.ReportDate, "-") != 2 {
		resp.BadRequest(c, "reportDate 格式应为 YYYY-MM-DD")
		return
	}
	d, err := h.saveReportDetail(in)
	if err != nil {
		if errors.Is(err, errOutsourceNameEmpty) {
			resp.BadRequest(c, err.Error())
			return
		}
		resp.Error(c, http.StatusInternalServerError, "保存失败: "+err.Error())
		return
	}
	c.JSON(http.StatusOK, d)
}

// DeleteReport 删除日报，并显式清理五个模块子表（避免留下孤儿行）。
func (h *Handler) DeleteReport(c *gin.Context) {
	id := c.Param("id")
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("report_id = ?", id).Delete(&model.ReportWorker{}).Error; err != nil {
			return err
		}
		if err := tx.Where("report_id = ?", id).Delete(&model.ReportShift{}).Error; err != nil {
			return err
		}
		if err := tx.Where("report_id = ?", id).Delete(&model.ReportMeal{}).Error; err != nil {
			return err
		}
		if err := tx.Where("report_id = ?", id).Delete(&model.ReportEquipment{}).Error; err != nil {
			return err
		}
		if err := tx.Where("report_id = ?", id).Delete(&model.ReportOutsourcing{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.DailyReport{}, id).Error
	})
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, "删除失败: "+err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ---------------- ① 用工名单 ----------------

// saveReportWorkers 全量覆盖名单：先删后插，一行一工人。
func saveReportWorkers(tx *gorm.DB, reportID uint, workers []map[string]any) error {
	if err := tx.Where("report_id = ?", reportID).Delete(&model.ReportWorker{}).Error; err != nil {
		return err
	}
	rows := make([]model.ReportWorker, 0, len(workers))
	order := 0
	for _, w := range workers {
		name := trimStr(anyToStr(w["name"]))
		if name == "" {
			continue
		}
		workerID, err := ensureWorker(tx, name)
		if err != nil {
			return err
		}
		start, end := workTime(w["start"]), workTime(w["end"])
		rows = append(rows, model.ReportWorker{
			ReportID:   reportID,
			WorkerID:   workerID,
			WorkerName: name,
			StartTime:  start,
			EndTime:    end,
			Overnight:  end < start, // 下班早于上班视为跨次日
			SortOrder:  uint(order),
		})
		order++
	}
	if len(rows) == 0 {
		return nil
	}
	return tx.Create(&rows).Error
}

// ensureWorker 按姓名取人员，不存在则自动建档（历史名单中的人员也必须可追溯）。
func ensureWorker(tx *gorm.DB, name string) (uint, error) {
	var w model.Worker
	err := tx.Where("name = ?", name).First(&w).Error
	if err == nil {
		return w.ID, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}
	full, ini := toPinyin(name)
	w = model.Worker{Name: name, PinyinFull: full, PinyinInitial: ini, Active: true}
	if err := tx.Create(&w).Error; err != nil {
		return 0, err
	}
	return w.ID, nil
}

// ---------------- ② 用工人数 ----------------

func saveReportShift(tx *gorm.DB, reportID uint, shifts map[string]any) error {
	s := model.ReportShift{
		ReportID:     reportID,
		DayCount:     anyToInt(shifts["day"]),
		EveningCount: anyToInt(shifts["eve"]),
		DawnCount:    anyToInt(shifts["dawn"]),
	}
	return tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "report_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"day_count", "evening_count", "dawn_count"}),
	}).Create(&s).Error
}

// ---------------- ③ 餐费 ----------------

func saveReportMeal(tx *gorm.DB, reportID uint, meals map[string]any) error {
	m := model.ReportMeal{
		ReportID:     reportID,
		BreakfastQty: anyToInt(meals["breakfast"]),
		LunchQty:     anyToInt(meals["lunch"]),
		DinnerQty:    anyToInt(meals["dinner"]),
		SupperQty:    anyToInt(meals["supper"]),
		WaterQty:     anyToInt(meals["water"]),
	}
	return tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "report_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"breakfast_qty", "lunch_qty", "dinner_qty", "supper_qty", "water_qty",
		}),
	}).Create(&m).Error
}

// ---------------- ④ 设备台班 ----------------

// saveReportEquipment 固定写入 5 设备 × 3 班次 = 15 行，未填的补 0。
func saveReportEquipment(tx *gorm.DB, reportID uint, equipment map[string]any) error {
	rows := make([]model.ReportEquipment, 0, len(model.EquipDefs)*len(model.ShiftDefs))
	for _, ed := range model.EquipDefs {
		shifts, _ := equipment[ed.Key].(map[string]any)
		for _, sd := range model.ShiftDefs {
			rows = append(rows, model.ReportEquipment{
				ReportID:  reportID,
				EquipType: ed.Code,
				ShiftType: sd.Code,
				Qty:       anyToInt(shifts[sd.Key]),
			})
		}
	}
	return tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "report_id"}, {Name: "equip_type"}, {Name: "shift_type"}},
		DoUpdates: clause.AssignmentColumns([]string{"qty"}),
	}).Create(&rows).Error
}

// ---------------- ⑤ 外包 ----------------

// errOutsourceNameEmpty 外包单位填了人数/加班却没填名称：属于参数错误（前端提示补全即可），不是服务端故障。
var errOutsourceNameEmpty = errors.New("外包单位名称未填")

// outsourceRowHasValue 判断一行外包单位是否填了内容（人数或加班时段任一非空/非 0）。
func outsourceRowHasValue(o map[string]any) bool {
	for _, k := range []string{"day", "eve", "dawn", "dayOt", "eveOt", "dawnOt"} {
		if anyToInt(o[k]) != 0 {
			return true
		}
	}
	for _, k := range []string{"dayOtStart", "dayOtEnd", "eveOtStart", "eveOtEnd", "dawnOtStart", "dawnOtEnd", "otStart", "otEnd"} {
		if trimStr(anyToStr(o[k])) != "" {
			return true
		}
	}
	return false
}

// saveReportOutsourcing 全量覆盖外包单位：先删后插，一个单位一行。
func saveReportOutsourcing(tx *gorm.DB, reportID uint, list []map[string]any) error {
	if err := tx.Where("report_id = ?", reportID).Delete(&model.ReportOutsourcing{}).Error; err != nil {
		return err
	}
	rows := make([]model.ReportOutsourcing, 0, len(list))
	order := 0
	for i, o := range list {
		name := trimStr(anyToStr(o["name"]))
		if name == "" {
			// 单位名称是这一行的身份，空名称无法落库：只有完全没填的空行才忽略，
			// 填了人数/加班却漏填名称必须报错，否则前端会「填了却看不到」（历史行为是静默丢弃）
			if !outsourceRowHasValue(o) {
				continue
			}
			return fmt.Errorf("%w：第 %d 个外包单位", errOutsourceNameEmpty, i+1)
		}
		rows = append(rows, model.ReportOutsourcing{
			ReportID:       reportID,
			UnitName:       name,
			DayCount:       anyToInt(o["day"]),
			EveningCount:   anyToInt(o["eve"]),
			DawnCount:      anyToInt(o["dawn"]),
			DayOtCount:     anyToInt(o["dayOt"]),
			DayOtStart:     optTime(o["dayOtStart"]),
			DayOtEnd:       optTime(o["dayOtEnd"]),
			EveningOtCount: anyToInt(o["eveOt"]),
			EveningOtStart: optTime(o["eveOtStart"]),
			EveningOtEnd:   optTime(o["eveOtEnd"]),
			DawnOtCount:    anyToInt(o["dawnOt"]),
			DawnOtStart:    optTime(o["dawnOtStart"]),
			DawnOtEnd:      optTime(o["dawnOtEnd"]),
			// 兼容旧前端仍只传一组 otStart/otEnd 的情况
			OtStart:   optTime(o["otStart"]),
			OtEnd:     optTime(o["otEnd"]),
			SortOrder: uint(order),
		})
		order++
	}
	if len(rows) == 0 {
		return nil
	}
	return tx.Create(&rows).Error
}
