// Package router 负责组装 Gin 引擎与路由表。
package router

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"lineng/internal/handler"
	"lineng/pkg/jwt"
	"lineng/pkg/middleware"
)

// Setup 构建并返回配置好中间件与路由的 Gin 引擎。
func Setup(h *handler.Handler) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORS())

	api := r.Group("/api")
	{
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"pong": time.Now().Unix()})
		})
		api.POST("/login", h.Login)
		api.GET("/h5/projects", h.H5Projects) // H5 分享链接拉取项目（签名校验，不走 JWT）
		api.GET("/h5/workers", h.H5Workers)   // H5 分享链接拉取人员管理列表（签名校验，不走 JWT）
		// H5 模块暂存草稿（签名校验，同项目同日期同模块有则更新、无则新增）
		api.POST("/h5/report-draft", h.H5SaveModuleDraft)
		api.GET("/h5/report-drafts", h.H5GetModuleDrafts)
		api.DELETE("/h5/report-drafts", h.H5DeleteModuleDrafts)
		// H5 分享链接提交日报（签名校验，与 PC /api/report 复写入库逻辑）
		api.POST("/h5/report", h.H5SubmitReport)
		// H5 分享链接读取/删除日报（签名校验，不走 JWT）：展示以服务器为准
		api.GET("/h5/report", h.H5GetReport)
		api.GET("/h5/reports", h.H5ListReports)
		api.DELETE("/h5/report/:id", h.H5DeleteReport)
	}

	auth := api.Group("", jwt.Middleware())
	{
		auth.GET("/me", h.Me)

		// 项目
		auth.GET("/projects", h.ListProjects)
		auth.POST("/projects", h.CreateProject)
		auth.GET("/projects/:id", h.GetProject)
		auth.PUT("/projects/:id", h.UpdateProject)
		auth.DELETE("/projects/:id", h.DeleteProject)

		// 人员
		auth.GET("/workers", h.ListWorkers)
		auth.POST("/workers", h.CreateWorker)
		auth.PUT("/workers/:id", h.UpdateWorker)
		auth.DELETE("/workers/:id", h.DeleteWorker)

		// 日报
		auth.GET("/reports", h.ListReports)
		auth.GET("/report/:projectId/:date", h.GetReport)
		auth.POST("/report", h.UpsertReport)
		auth.DELETE("/report/:id", h.DeleteReport)

		// 统计
		auth.GET("/stats", h.GetStats)

		// 导出
		auth.GET("/export/settlement", h.ExportSettlement)
		auth.GET("/export/attendance", h.ExportAttendance)
	}

	return r
}
