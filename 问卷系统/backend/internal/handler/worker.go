package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"lineng/internal/model"
	"lineng/pkg/resp"
)

// ListWorkers 人员列表，支持汉字/拼音/首字母搜索。
func (h *Handler) ListWorkers(c *gin.Context) {
	q := trimStr(c.Query("q"))
	query := h.db.Model(&model.Worker{}).Order("pinyin_full ASC, id ASC")
	if q != "" {
		lq := strings.ToLower(q)
		query = query.Where(
			"name LIKE ? OR pinyin_full LIKE ? OR pinyin_initial LIKE ?",
			"%"+q+"%", "%"+lq+"%", "%"+lq+"%")
	}
	var ws []model.Worker
	query.Find(&ws)
	c.JSON(http.StatusOK, ws)
}

// CreateWorker 新增人员，自动生成拼音全拼与首字母。
func (h *Handler) CreateWorker(c *gin.Context) {
	var w model.Worker
	if err := c.ShouldBindJSON(&w); err != nil {
		resp.BadRequest(c, "参数错误")
		return
	}
	w.Name = trimStr(w.Name)
	if w.Name == "" {
		resp.BadRequest(c, "姓名必填")
		return
	}
	var n int64
	h.db.Model(&model.Worker{}).Where("name = ?", w.Name).Count(&n)
	if n > 0 {
		resp.BadRequest(c, fmt.Sprintf("人员「%s」已存在", w.Name))
		return
	}
	w.PinyinFull, w.PinyinInitial = toPinyin(w.Name)
	w.Active = true
	if err := h.db.Create(&w).Error; err != nil {
		resp.Error(c, http.StatusInternalServerError, "创建失败: "+err.Error())
		return
	}
	c.JSON(http.StatusOK, w)
}

// UpdateWorker 更新人员信息并重算拼音。
func (h *Handler) UpdateWorker(c *gin.Context) {
	var w model.Worker
	if err := h.db.First(&w, c.Param("id")).Error; err != nil {
		resp.Error(c, http.StatusNotFound, "人员不存在")
		return
	}
	var in model.Worker
	if err := c.ShouldBindJSON(&in); err != nil {
		resp.BadRequest(c, "参数错误")
		return
	}
	in.Name = trimStr(in.Name)
	if in.Name == "" {
		resp.BadRequest(c, "姓名必填")
		return
	}
	var n int64
	h.db.Model(&model.Worker{}).Where("name = ? AND id <> ?", in.Name, w.ID).Count(&n)
	if n > 0 {
		resp.BadRequest(c, fmt.Sprintf("人员「%s」已存在", in.Name))
		return
	}
	full, ini := toPinyin(in.Name)
	if err := h.db.Model(&w).Updates(map[string]any{
		"name": in.Name, "phone": in.Phone, "job_type": in.JobType, "note": in.Note,
		"pinyin_full": full, "pinyin_initial": ini, "active": in.Active,
	}).Error; err != nil {
		resp.Error(c, http.StatusInternalServerError, "更新失败")
		return
	}
	h.db.First(&w, w.ID)
	c.JSON(http.StatusOK, w)
}

// DeleteWorker 删除人员。
func (h *Handler) DeleteWorker(c *gin.Context) {
	h.db.Delete(&model.Worker{}, c.Param("id"))
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
