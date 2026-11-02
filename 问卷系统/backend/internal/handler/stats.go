package handler

import (
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"

	"lineng/internal/model"
)

// GetStats 汇总统计：卡片 + 人员出勤排行 + 项目汇总 + 日报明细。
// 卡片与排行都在 SQL 侧聚合，明细由五个模块子表组装。
func (h *Handler) GetStats(c *gin.Context) {
	q := h.db.Model(&model.DailyReport{})
	if pid := c.Query("projectId"); pid != "" {
		q = q.Where("project_id = ?", pid)
	}
	if st, en := c.Query("start"), c.Query("end"); st != "" && en != "" {
		q = q.Where("report_date BETWEEN ? AND ?", st, en)
	}
	var rs []model.DailyReport
	q.Order("report_date ASC").Find(&rs)
	ids := reportIDs(rs)
	details := h.loadDetails(rs)

	// —— 卡片 ——
	heads, entries, distinct := 0, 0, 0
	if len(ids) > 0 {
		h.db.Raw("SELECT COALESCE(SUM(day_count + evening_count + dawn_count), 0) FROM report_shifts WHERE report_id IN ?", ids).Scan(&heads)
		h.db.Raw("SELECT COUNT(*) FROM report_workers WHERE report_id IN ?", ids).Scan(&entries)
		h.db.Raw("SELECT COUNT(DISTINCT worker_id) FROM report_workers WHERE report_id IN ?", ids).Scan(&distinct)
	}

	// —— 人员出勤排行 ——
	type workerRow struct {
		Name    string `json:"name"`
		Entries int    `json:"entries"`
		Days    int    `json:"days"`
	}
	var wRows []workerRow
	if len(ids) > 0 {
		h.db.Raw(`SELECT w.worker_name AS name, COUNT(*) AS entries, COUNT(DISTINCT w.report_id) AS days
			FROM report_workers w
			WHERE w.report_id IN ?
			GROUP BY w.worker_id, w.worker_name
			ORDER BY entries DESC, name ASC`, ids).Scan(&wRows)
	}
	perWorker := make([]workerRow, 0, len(wRows))
	perWorker = append(perWorker, wRows...)

	// —— 项目汇总 ——
	type pair struct {
		ProjectID uint
		Cnt       int
	}
	type projStat struct {
		ID      uint   `json:"id"`
		Name    string `json:"name"`
		Reports int    `json:"reports"`
		Heads   int    `json:"heads"`
		Entries int    `json:"entries"`
	}
	var ps []model.Project
	h.db.Find(&ps)
	projName := map[uint]string{}
	for _, p := range ps {
		projName[p.ID] = p.Unit + "·" + p.ActivityName
	}
	agg := map[uint]*projStat{}
	var repCnt, headCnt, entryCnt []pair
	if len(ids) > 0 {
		h.db.Raw("SELECT project_id, COUNT(*) AS cnt FROM daily_reports WHERE id IN ? GROUP BY project_id", ids).Scan(&repCnt)
		h.db.Raw(`SELECT r.project_id, COALESCE(SUM(s.day_count + s.evening_count + s.dawn_count), 0) AS cnt
			FROM report_shifts s JOIN daily_reports r ON r.id = s.report_id
			WHERE s.report_id IN ? GROUP BY r.project_id`, ids).Scan(&headCnt)
		h.db.Raw(`SELECT r.project_id, COUNT(*) AS cnt
			FROM report_workers w JOIN daily_reports r ON r.id = w.report_id
			WHERE w.report_id IN ? GROUP BY r.project_id`, ids).Scan(&entryCnt)
	}
	pick := func(id uint) *projStat {
		st := agg[id]
		if st == nil {
			st = &projStat{ID: id, Name: projName[id]}
			agg[id] = st
		}
		return st
	}
	for _, p := range repCnt {
		pick(p.ProjectID).Reports = p.Cnt
	}
	for _, p := range headCnt {
		pick(p.ProjectID).Heads = p.Cnt
	}
	for _, p := range entryCnt {
		pick(p.ProjectID).Entries = p.Cnt
	}
	perProject := make([]projStat, 0, len(agg))
	for _, st := range agg {
		perProject = append(perProject, *st)
	}
	sort.Slice(perProject, func(i, j int) bool { return perProject[i].Reports > perProject[j].Reports })

	c.JSON(http.StatusOK, gin.H{
		"cards": gin.H{
			"reports": len(details), "heads": heads, "entries": entries, "distinctWorkers": distinct,
		},
		"perWorker":  perWorker,
		"perProject": perProject,
		"details":    details,
	})
}
