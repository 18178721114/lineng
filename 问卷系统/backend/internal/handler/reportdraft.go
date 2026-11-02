package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"

	"lineng/internal/model"
	"lineng/pkg/resp"
)

// moduleDraftSaveReq H5 模块暂存入库请求体。
type moduleDraftSaveReq struct {
	ProjectID  uint           `json:"projectId"`
	ReportDate string         `json:"reportDate"`
	ModuleKey  string         `json:"moduleKey"`
	Data       map[string]any `json:"data"`
}

// H5SaveModuleDraft 暂存某模块草稿（签名校验，同 H5Projects）。
// 同项目同日期同模块有则更新、无则新增，保证一个模块只有一条草稿。
func (h *Handler) H5SaveModuleDraft(c *gin.Context) {
	if _, _, _, ok := h.verifyShareLink(c); !ok {
		return
	}
	var in moduleDraftSaveReq
	if err := c.ShouldBindJSON(&in); err != nil {
		resp.BadRequest(c, "参数错误")
		return
	}
	if in.ProjectID == 0 || in.ModuleKey == "" {
		resp.BadRequest(c, "projectId 与 moduleKey 必填")
		return
	}
	if len(in.ReportDate) != 10 || strings.Count(in.ReportDate, "-") != 2 {
		resp.BadRequest(c, "reportDate 格式应为 YYYY-MM-DD")
		return
	}
	var cnt int64
	if err := h.db.Model(&model.Project{}).Where("id = ?", in.ProjectID).Count(&cnt).Error; err != nil || cnt == 0 {
		resp.BadRequest(c, "项目不存在")
		return
	}
	content, err := json.Marshal(in.Data)
	if err != nil {
		resp.BadRequest(c, "暂存内容格式错误")
		return
	}
	if in.Data == nil {
		content = []byte("{}")
	}
	d := model.ReportModuleDraft{
		ProjectID:  in.ProjectID,
		ReportDate: model.Date(in.ReportDate),
		ModuleKey:  in.ModuleKey,
		Content:    string(content),
	}
	if err := h.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "project_id"}, {Name: "report_date"}, {Name: "module_key"}},
		DoUpdates: clause.AssignmentColumns([]string{"content"}),
	}).Create(&d).Error; err != nil {
		resp.Error(c, http.StatusInternalServerError, "暂存失败: "+err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// H5GetModuleDrafts 拉取某项目某天的全部模块草稿（进填写页时回填用）。
func (h *Handler) H5GetModuleDrafts(c *gin.Context) {
	if _, _, _, ok := h.verifyShareLink(c); !ok {
		return
	}
	pid, date := c.Query("projectId"), c.Query("reportDate")
	if pid == "" || date == "" {
		resp.BadRequest(c, "projectId 与 reportDate 必填")
		return
	}
	var ds []model.ReportModuleDraft
	if err := h.db.Where("project_id = ? AND report_date = ?", pid, date).Find(&ds).Error; err != nil {
		resp.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	out := make([]map[string]any, 0, len(ds))
	for _, d := range ds {
		var data map[string]any
		_ = json.Unmarshal([]byte(d.Content), &data)
		out = append(out, map[string]any{
			"moduleKey": d.ModuleKey,
			"data":      data,
			"updatedAt": d.UpdatedAt.Format("2006-01-02 15:04"),
		})
	}
	c.JSON(http.StatusOK, out)
}

// H5DeleteModuleDrafts 提交日报后清空该项目当天的全部模块草稿。
func (h *Handler) H5DeleteModuleDrafts(c *gin.Context) {
	if _, _, _, ok := h.verifyShareLink(c); !ok {
		return
	}
	pid, date := c.Query("projectId"), c.Query("reportDate")
	if pid == "" || date == "" {
		resp.BadRequest(c, "projectId 与 reportDate 必填")
		return
	}
	if err := h.db.Where("project_id = ? AND report_date = ?", pid, date).Delete(&model.ReportModuleDraft{}).Error; err != nil {
		resp.Error(c, http.StatusInternalServerError, "删除失败")
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
