package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"lineng/internal/model"
	"lineng/pkg/resp"
)

// H5SubmitReport 分享链接手机端提交日报（签名校验，不走 JWT）。
// 请求体即与 PC 端一致的 ReportDetail，由 H5 页面在提交时把表单字段转换后上传。
func (h *Handler) H5SubmitReport(c *gin.Context) {
	if _, _, _, ok := h.verifyShareLink(c); !ok {
		return
	}
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
	// 校验项目存在：签名即使被复用也只能写入合法项目，不能任意写库。
	var cnt int64
	if err := h.db.Model(&model.Project{}).Where("id = ?", in.ProjectID).Count(&cnt).Error; err != nil || cnt == 0 {
		resp.BadRequest(c, "项目不存在")
		return
	}
	d, err := h.saveReportDetail(in)
	if err != nil {
		if errors.Is(err, errOutsourceNameEmpty) {
			resp.BadRequest(c, err.Error())
			return
		}
		resp.Error(c, http.StatusInternalServerError, "提交失败: "+err.Error())
		return
	}
	c.JSON(http.StatusOK, d)
}

// H5GetReport 供手机端 H5 分享链接拉取某天已提交的日报（签名校验，不走 JWT）。
// 不存在时返回 404，H5 据此展示空白表单（不再读本机缓存）。
func (h *Handler) H5GetReport(c *gin.Context) {
	if _, _, _, ok := h.verifyShareLink(c); !ok {
		return
	}
	pidStr, date := c.Query("projectId"), c.Query("reportDate")
	pid, err := strconv.ParseUint(strings.TrimSpace(pidStr), 10, 64)
	if pid == 0 || err != nil || date == "" {
		resp.BadRequest(c, "projectId 与 reportDate 必填")
		return
	}
	if len(date) != 10 || strings.Count(date, "-") != 2 {
		resp.BadRequest(c, "reportDate 格式应为 YYYY-MM-DD")
		return
	}
	var r model.DailyReport
	if err := h.db.Where("project_id = ? AND report_date = ?", pid, date).First(&r).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "未找到日报"})
			return
		}
		resp.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	c.JSON(http.StatusOK, h.loadDetails([]model.DailyReport{r})[0])
}

// H5ListReports 供手机端 H5 分享链接拉取某项目全部日报（签名校验，不走 JWT）。
// 用于首页/项目页/统计页聚合展示，使后台修改后 H5 立即可见。
func (h *Handler) H5ListReports(c *gin.Context) {
	if _, _, _, ok := h.verifyShareLink(c); !ok {
		return
	}
	pidStr := c.Query("projectId")
	pid, err := strconv.ParseUint(strings.TrimSpace(pidStr), 10, 64)
	if pid == 0 || err != nil {
		resp.BadRequest(c, "projectId 必填")
		return
	}
	var rs []model.DailyReport
	h.db.Where("project_id = ?", pid).Order("report_date DESC").Find(&rs)
	c.JSON(http.StatusOK, h.loadDetails(rs))
}

// H5DeleteReport 供手机端 H5 分享链接删除日报（签名校验，不走 JWT）。
// 与 PC 端 DeleteReport 同逻辑：清理主表 + 五个模块子表。
func (h *Handler) H5DeleteReport(c *gin.Context) {
	if _, _, _, ok := h.verifyShareLink(c); !ok {
		return
	}
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
