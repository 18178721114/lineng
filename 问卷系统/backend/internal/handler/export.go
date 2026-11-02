package handler

import (
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"

	"lineng/internal/model"
	"lineng/pkg/resp"
)

// PriceConf 结算单价配置。
type PriceConf struct {
	Labor, Meal, Water, Fork3, Fork5, Fork7, Crane, Rescue float64
	TaxRate                                                float64
}

func (h *Handler) loadPrices(c *gin.Context, projectID uint) PriceConf {
	p := PriceConf{Labor: 380, Meal: 20, Water: 5, Fork3: 1200, Fork5: 1600, Fork7: 2000, Crane: 1200, Rescue: 800, TaxRate: 6}
	var pp model.ProjectPrices
	if err := h.db.Where("project_id = ?", projectID).First(&pp).Error; err == nil {
		get := func(v, d float64) float64 {
			if v > 0 {
				return v
			}
			return d
		}
		p.Labor, p.Meal, p.Water = get(pp.LaborPrice, p.Labor), get(pp.MealPrice, p.Meal), get(pp.WaterPrice, p.Water)
		p.Fork3, p.Fork5, p.Fork7 = get(pp.Fork3Price, p.Fork3), get(pp.Fork5Price, p.Fork5), get(pp.Fork7Price, p.Fork7)
		p.Crane, p.Rescue, p.TaxRate = get(pp.CranePrice, p.Crane), get(pp.RescuePrice, p.Rescue), get(pp.TaxRate, p.TaxRate)
	}
	override := func(key string, dst *float64) {
		if v := c.Query(key); v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil && f >= 0 {
				*dst = f
			}
		}
	}
	override("laborPrice", &p.Labor)
	override("mealPrice", &p.Meal)
	override("waterPrice", &p.Water)
	override("fork3Price", &p.Fork3)
	override("fork5Price", &p.Fork5)
	override("fork7Price", &p.Fork7)
	override("cranePrice", &p.Crane)
	override("rescuePrice", &p.Rescue)
	override("taxRate", &p.TaxRate)
	return p
}

// ---------------- 结算单 ----------------

// ExportSettlement 导出结算单 Excel。
func (h *Handler) ExportSettlement(c *gin.Context) {
	projectID := parseUint(c.Query("projectId"))
	start, end := c.Query("start"), c.Query("end")
	if projectID == 0 || start == "" || end == "" {
		resp.BadRequest(c, "projectId、start、end 必填")
		return
	}
	var proj model.Project
	if err := h.db.First(&proj, projectID).Error; err != nil {
		resp.Error(c, http.StatusNotFound, "项目不存在")
		return
	}
	prices := h.loadPrices(c, projectID)

	dates := dateRange(start, end)

	// ② 用工人数
	type shiftRow struct {
		ReportDate   string
		DayCount     int
		EveningCount int
		DawnCount    int
	}
	var shiftRows []shiftRow
	h.db.Table("report_shifts s").
		Select("DATE_FORMAT(r.report_date, '%Y-%m-%d') AS report_date, s.day_count, s.evening_count, s.dawn_count").
		Joins("JOIN daily_reports r ON r.id = s.report_id").
		Where("r.project_id = ? AND r.report_date BETWEEN ? AND ?", projectID, start, end).
		Scan(&shiftRows)
	byShift := make(map[string]shiftRow, len(shiftRows))
	for _, s := range shiftRows {
		byShift[s.ReportDate] = s
	}

	// ③ 餐费（含水）
	type mealRow struct {
		ReportDate   string
		BreakfastQty int
		LunchQty     int
		DinnerQty    int
		SupperQty    int
		WaterQty     int
	}
	var mealRows []mealRow
	h.db.Table("report_meals m").
		Select("DATE_FORMAT(r.report_date, '%Y-%m-%d') AS report_date, m.breakfast_qty, m.lunch_qty, m.dinner_qty, m.supper_qty, m.water_qty").
		Joins("JOIN daily_reports r ON r.id = m.report_id").
		Where("r.project_id = ? AND r.report_date BETWEEN ? AND ?", projectID, start, end).
		Scan(&mealRows)
	byMeal := make(map[string]mealRow, len(mealRows))
	for _, m := range mealRows {
		byMeal[m.ReportDate] = m
	}

	// ④ 设备台班
	type equipRow struct {
		ReportDate string
		EquipType  string
		ShiftType  string
		Qty        int
	}
	var equipRows []equipRow
	h.db.Table("report_equipment e").
		Select("DATE_FORMAT(r.report_date, '%Y-%m-%d') AS report_date, e.equip_type, e.shift_type, e.qty").
		Joins("JOIN daily_reports r ON r.id = e.report_id").
		Where("r.project_id = ? AND r.report_date BETWEEN ? AND ?", projectID, start, end).
		Scan(&equipRows)
	byEquip := map[string]map[string]map[string]int{} // 日期 → 设备 → 班次 → 台班
	for _, e := range equipRows {
		if byEquip[e.ReportDate] == nil {
			byEquip[e.ReportDate] = map[string]map[string]int{}
		}
		if byEquip[e.ReportDate][e.EquipType] == nil {
			byEquip[e.ReportDate][e.EquipType] = map[string]int{}
		}
		byEquip[e.ReportDate][e.EquipType][e.ShiftType] = e.Qty
	}

	f := excelize.NewFile()
	sheet := "Sheet1"
	f.SetColWidth(sheet, "A", "A", 22)
	f.SetColWidth(sheet, "B", "G", 12)

	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 14},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	headStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	amountStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Color: "C00000"}})

	set := func(row int, col int, v any) string {
		cell, _ := excelize.CoordinatesToCellName(col, row)
		f.SetCellValue(sheet, cell, v)
		return cell
	}

	r := 1
	set(r, 1, "北京力能文化艺术服务有限公司")
	f.MergeCell(sheet, fmt.Sprintf("A%d", r), fmt.Sprintf("F%d", r))
	f.SetCellStyle(sheet, "A1", "A1", titleStyle)
	r++
	year := start[:4]
	set(r, 1, year+"年"+proj.Unit+proj.ActivityName+"用工记录")
	f.MergeCell(sheet, fmt.Sprintf("A%d", r), fmt.Sprintf("F%d", r))
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", r), fmt.Sprintf("A%d", r), headStyle)
	r += 2

	// —— 人工 ——
	set(r, 1, "班次日期")
	set(r, 2, "白天")
	set(r, 3, "晚上")
	set(r, 4, "凌晨")
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", r), fmt.Sprintf("D%d", r), headStyle)
	r++
	sumDay, sumEve, sumDawn := 0.0, 0.0, 0.0
	for _, d := range dates {
		s, ok := byShift[d]
		if !ok {
			continue
		}
		day, eve, dawn := float64(s.DayCount), float64(s.EveningCount), float64(s.DawnCount)
		if day == 0 && eve == 0 && dawn == 0 {
			continue
		}
		set(r, 1, md(d))
		set(r, 2, day)
		set(r, 3, eve)
		set(r, 4, dawn)
		sumDay, sumEve, sumDawn = sumDay+day, sumEve+eve, sumDawn+dawn
		r++
	}
	set(r, 1, "合计")
	set(r, 2, sumDay)
	set(r, 3, sumEve)
	set(r, 4, sumDawn)
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", r), fmt.Sprintf("D%d", r), headStyle)
	r++
	totalHeads := sumDay + sumEve + sumDawn
	laborAmount := round2(totalHeads * prices.Labor)
	set(r, 1, "人工合计金额")
	set(r, 2, laborAmount)
	set(r, 3, laborAmount)
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", r), fmt.Sprintf("C%d", r), amountStyle)
	r += 2

	// —— 餐费（含水） ——
	set(r, 1, "餐费日期")
	for i, hh := range []string{"早点", "中餐", "晚餐", "夜宵", "水"} {
		set(r, 2+i, hh)
	}
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", r), fmt.Sprintf("F%d", r), headStyle)
	r++
	mealSum, waterSum := 0.0, 0.0
	for _, d := range dates {
		m, ok := byMeal[d]
		if !ok {
			continue
		}
		b, l, dn, sp, w := float64(m.BreakfastQty), float64(m.LunchQty), float64(m.DinnerQty), float64(m.SupperQty), float64(m.WaterQty)
		if b+l+dn+sp+w == 0 {
			continue
		}
		set(r, 1, md(d))
		set(r, 2, b)
		set(r, 3, l)
		set(r, 4, dn)
		set(r, 5, sp)
		set(r, 6, w)
		mealSum += b + l + dn + sp
		waterSum += w
		r++
	}
	set(r, 1, "水合计")
	set(r, 2, waterSum)
	set(r, 3, "元/份")
	set(r, 4, prices.Water)
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", r), fmt.Sprintf("D%d", r), headStyle)
	r++
	set(r, 1, "合计")
	set(r, 2, mealSum)
	set(r, 3, "元/份")
	set(r, 4, prices.Meal)
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", r), fmt.Sprintf("D%d", r), headStyle)
	r++
	mealAmount := round2(mealSum*prices.Meal + waterSum*prices.Water)
	set(r, 1, "餐费合计金额")
	set(r, 2, mealAmount)
	set(r, 3, mealAmount)
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", r), fmt.Sprintf("C%d", r), amountStyle)
	r += 2

	// —— 设备台班：叉车3/5/7吨、吊车、救援车 ——
	type equipDef struct {
		code, title string
		price       float64
		unitPrice   bool // 叉车打印算式行
	}
	equips := []equipDef{
		{model.EquipFork3, "叉车3吨", prices.Fork3, true},
		{model.EquipFork5, "叉车5吨", prices.Fork5, true},
		{model.EquipFork7, "叉车7吨", prices.Fork7, true},
		{model.EquipCrane, "吊车", prices.Crane, false},
		{model.EquipRescue, "救援车", prices.Rescue, false},
	}
	equipAmounts := map[string]float64{}
	for _, eq := range equips {
		// 先收集该设备有台班的日期，避免两轮遍历；台班数不分班次，取三班之和
		type equipLine struct {
			date string
			qty  float64
		}
		var lines []equipLine
		totalUnits := 0.0
		for _, d := range dates {
			shifts := byEquip[d][eq.code]
			qty := float64(shifts[model.ShiftDay] + shifts[model.ShiftEvening] + shifts[model.ShiftDawn])
			if qty == 0 {
				continue
			}
			lines = append(lines, equipLine{d, qty})
			totalUnits += qty
		}
		if len(lines) == 0 {
			continue
		}
		set(r, 1, eq.title+"台班")
		set(r, 2, "台班数")
		f.SetCellStyle(sheet, fmt.Sprintf("A%d", r), fmt.Sprintf("B%d", r), headStyle)
		r++
		for _, ln := range lines {
			set(r, 1, md(ln.date))
			set(r, 2, ln.qty)
			r++
		}
		amount := round2(totalUnits * eq.price)
		equipAmounts[eq.code] = amount
		if eq.unitPrice {
			set(r, 1, fmt.Sprintf("%s：%.0f台*%.0f=%.0f元", eq.title, totalUnits, eq.price, amount))
		} else {
			set(r, 1, eq.title+"合计金额")
			set(r, 2, amount)
		}
		f.SetCellStyle(sheet, fmt.Sprintf("A%d", r), fmt.Sprintf("B%d", r), headStyle)
		r++
	}
	r++

	// —— 汇总 ——
	grand := laborAmount + mealAmount
	for _, v := range equipAmounts {
		grand += v
	}
	grand = round2(grand)
	tax := round2(grand * prices.TaxRate / 100)
	final := round2(grand + tax)
	set(r, 1, "人工、餐费、设备合计金额")
	set(r, 2, grand)
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", r), fmt.Sprintf("B%d", r), headStyle)
	r++
	set(r, 1, fmt.Sprintf("税费%.0f%%", prices.TaxRate))
	set(r, 2, tax)
	r++
	set(r, 1, "劳务总合计金额")
	set(r, 2, final)
	set(r, 3, final)
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", r), fmt.Sprintf("C%d", r), amountStyle)

	buf, err := f.WriteToBuffer()
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, "生成 Excel 失败")
		return
	}
	name := fmt.Sprintf("结算单_%s_%s至%s.xlsx", proj.Unit, start, end)
	download(c, buf.Bytes(), name)
}

func download(c *gin.Context, data []byte, name string) {
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename*=UTF-8''"+url.PathEscape(name))
	c.Data(http.StatusOK, "application/octet-stream", data)
}

// ---------------- 考勤表 ----------------

// ExportAttendance 导出月度考勤表 Excel。
func (h *Handler) ExportAttendance(c *gin.Context) {
	projectID := parseUint(c.Query("projectId"))
	month := c.Query("month") // 2026-10
	if projectID == 0 || len(month) != 7 {
		resp.BadRequest(c, "projectId、month(YYYY-MM) 必填")
		return
	}
	var proj model.Project
	if err := h.db.First(&proj, projectID).Error; err != nil {
		resp.Error(c, http.StatusNotFound, "项目不存在")
		return
	}
	start := month + "-01"
	t, _ := time.Parse("2006-01", month)
	last := time.Date(t.Year(), t.Month()+1, 0, 0, 0, 0, 0, time.Local)
	monthEnd := last.Format("2006-01-02")
	days := last.Day()

	// 名单按「人员 × 日期」聚合，行数即工数
	type attRow struct {
		Name       string
		PinyinFull string
		DayNum     int
		Cnt        int
	}
	var rows []attRow
	h.db.Table("report_workers w").
		Select("wk.name AS name, wk.pinyin_full AS pinyin_full, DAYOFMONTH(r.report_date) AS day_num, COUNT(*) AS cnt").
		Joins("JOIN daily_reports r ON r.id = w.report_id").
		Joins("JOIN workers wk ON wk.id = w.worker_id").
		Where("r.project_id = ? AND r.report_date BETWEEN ? AND ?", projectID, start, monthEnd).
		Group("wk.id, wk.name, wk.pinyin_full, r.report_date").
		Scan(&rows)

	att := map[string]map[int]int{} // name -> day -> 工数
	pyOf := map[string]string{}
	names := map[string]bool{}
	for _, row := range rows {
		names[row.Name] = true
		pyOf[row.Name] = row.PinyinFull
		if att[row.Name] == nil {
			att[row.Name] = map[int]int{}
		}
		att[row.Name][row.DayNum] = row.Cnt
	}
	sorted := make([]string, 0, len(names))
	for n := range names {
		sorted = append(sorted, n)
	}
	sort.Slice(sorted, func(i, j int) bool {
		return (pyOf[sorted[i]] + sorted[i]) < (pyOf[sorted[j]] + sorted[j])
	})

	f := excelize.NewFile()
	sheet := fmt.Sprintf("%d月份", int(t.Month()))
	f.SetSheetName("Sheet1", sheet)
	f.SetColWidth(sheet, "A", "A", 12)
	f.SetColWidth(sheet, "B", fmt.Sprintf("%c", 'B'+days), 5)

	headStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	set := func(row, col int, v any) {
		cell, _ := excelize.CoordinatesToCellName(col, row)
		f.SetCellValue(sheet, cell, v)
	}
	set(1, 1, fmt.Sprintf("%d月份", int(t.Month())))
	for d := 1; d <= days; d++ {
		set(1, 1+d, d)
	}
	set(1, 2+days, "合计")
	f.SetCellStyle(sheet, "A1", fmt.Sprintf("%s1", cellCol(2+days)), headStyle)

	for i, name := range sorted {
		row := 2 + i
		set(row, 1, name)
		total := 0
		for d := 1; d <= days; d++ {
			v := att[name][d]
			total += v
			if v > 0 {
				set(row, 1+d, v)
			}
		}
		set(row, 2+days, total)
	}
	// 底部合计行
	if len(sorted) > 0 {
		row := 2 + len(sorted)
		set(row, 1, "合计")
		dayTotal := 0
		for d := 1; d <= days; d++ {
			s := 0
			for _, name := range sorted {
				s += att[name][d]
			}
			dayTotal += s
			if s > 0 {
				set(row, 1+d, s)
			}
		}
		set(row, 2+days, dayTotal)
		f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("%s%d", cellCol(2+days), row), headStyle)
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, "生成 Excel 失败")
		return
	}
	download(c, buf.Bytes(), fmt.Sprintf("%d月考勤表_%s_%s.xlsx", int(t.Month()), proj.Unit, proj.ActivityName))
}
