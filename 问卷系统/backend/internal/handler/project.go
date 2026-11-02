package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"lineng/internal/model"
	"lineng/pkg/resp"
)

// fillPrices 批量填充项目默认单价（来自 project_prices 表，作为响应中的派生字段）。
func (h *Handler) fillPrices(ps []model.Project) {
	ids := make([]uint, 0, len(ps))
	for _, p := range ps {
		ids = append(ids, p.ID)
	}
	if len(ids) == 0 {
		return
	}
	var list []model.ProjectPrices
	h.db.Where("project_id IN ?", ids).Find(&list)
	m := make(map[uint]model.ProjectPrices, len(list))
	for _, p := range list {
		m[p.ProjectID] = p
	}
	for i := range ps {
		if p, ok := m[ps[i].ID]; ok {
			ps[i].Prices = p.JSON()
		}
	}
}

func (h *Handler) genProjectNo() string {
	today := time.Now().Format("20060102")
	var n int64
	h.db.Model(&model.Project{}).Where("no LIKE ?", "LN-"+today+"-%").Count(&n)
	for i := n + 1; ; i++ {
		no := fmt.Sprintf("LN-%s-%03d", today, i)
		var c int64
		h.db.Model(&model.Project{}).Where("no = ?", no).Count(&c)
		if c == 0 {
			return no
		}
	}
}

// ListProjects 项目列表（含日报数量）。
func (h *Handler) ListProjects(c *gin.Context) {
	var ps []model.Project
	h.db.Order("id DESC").Find(&ps)
	h.fillPrices(ps)
	type item struct {
		model.Project
		ReportCount int64 `json:"reportCount"`
	}
	out := make([]item, 0, len(ps))
	for _, p := range ps {
		var n int64
		h.db.Model(&model.DailyReport{}).Where("project_id = ?", p.ID).Count(&n)
		out = append(out, item{p, n})
	}
	c.JSON(http.StatusOK, out)
}

// CreateProject 新建项目，自动生成单据编号。
func (h *Handler) CreateProject(c *gin.Context) {
	var p model.Project
	if err := c.ShouldBindJSON(&p); err != nil {
		resp.BadRequest(c, "参数错误")
		return
	}
	if trimStr(p.Unit) == "" || trimStr(p.ActivityName) == "" || trimStr(p.Location) == "" ||
		trimStr(p.PartyAOwner) == "" || p.StartDate == "" || p.EndDate == "" {
		resp.BadRequest(c, "用工单位、活动名称、地点、甲方负责人、起止日期均为必填")
		return
	}
	p.No = h.genProjectNo()
	if err := h.db.Create(&p).Error; err != nil {
		resp.Error(c, http.StatusInternalServerError, "创建失败: "+err.Error())
		return
	}
	// 默认单价独立成表：新建项目即写入一套系统默认价
	prices := model.DefaultProjectPrices(p.ID)
	if err := h.db.Create(&prices).Error; err != nil {
		resp.Error(c, http.StatusInternalServerError, "创建默认单价失败: "+err.Error())
		return
	}
	p.Prices = prices.JSON()
	c.JSON(http.StatusOK, p)
}

// GetProject 项目详情。
func (h *Handler) GetProject(c *gin.Context) {
	var p model.Project
	if err := h.db.First(&p, c.Param("id")).Error; err != nil {
		resp.Error(c, http.StatusNotFound, "项目不存在")
		return
	}
	h.fillPrices([]model.Project{p})
	c.JSON(http.StatusOK, p)
}

// UpdateProject 更新项目（编号不可改）。
func (h *Handler) UpdateProject(c *gin.Context) {
	var p model.Project
	if err := h.db.First(&p, c.Param("id")).Error; err != nil {
		resp.Error(c, http.StatusNotFound, "项目不存在")
		return
	}
	var in model.Project
	if err := c.ShouldBindJSON(&in); err != nil {
		resp.BadRequest(c, "参数错误")
		return
	}
	in.ID = p.ID
	in.No = p.No // 编号不可改
	if err := h.db.Model(&p).Updates(map[string]any{
		"unit": in.Unit, "activity_name": in.ActivityName, "location": in.Location,
		"party_a_owner": in.PartyAOwner, "party_a_phone": in.PartyAPhone,
		"start_date": in.StartDate, "end_date": in.EndDate,
	}).Error; err != nil {
		resp.Error(c, http.StatusInternalServerError, "更新失败")
		return
	}
	h.db.First(&p, p.ID)
	h.fillPrices([]model.Project{p})
	c.JSON(http.StatusOK, p)
}

// DeleteProject 删除项目（存在日报时禁止）。
func (h *Handler) DeleteProject(c *gin.Context) {
	id := c.Param("id")
	var n int64
	h.db.Model(&model.DailyReport{}).Where("project_id = ?", id).Count(&n)
	if n > 0 {
		resp.BadRequest(c, "该项目已有 "+fmt.Sprint(n)+" 张日报，不能删除")
		return
	}
	h.db.Delete(&model.Project{}, id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
