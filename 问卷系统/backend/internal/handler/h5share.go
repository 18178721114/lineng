package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"lineng/internal/model"
	"lineng/pkg/resp"
)

// H5Projects 供手机端 H5 分享链接拉取项目列表（公开接口，靠签名 + 有效期防盗刷）。
// 校验规则与 h5/index.html 的 verifyLink 保持一致：
//
//	sig = HMAC-SHA256(h5Secret, exp|to|ids|path)
//
// ids 为空时返回全部项目（对应签发工具不带项目限定的链接）。
func (h *Handler) H5Projects(c *gin.Context) {
	_, _, idsStr, ok := h.verifyShareLink(c)
	if !ok {
		return
	}

	q := h.db.Order("id DESC")
	if idsStr != "" {
		ids := make([]uint, 0, 8)
		for _, s := range strings.Split(idsStr, ",") {
			if id, err := strconv.ParseUint(strings.TrimSpace(s), 10, 64); err == nil && id > 0 {
				ids = append(ids, uint(id))
			}
		}
		if len(ids) == 0 {
			c.JSON(http.StatusOK, []model.Project{})
			return
		}
		q = q.Where("id IN ?", ids)
	}
	var ps []model.Project
	q.Find(&ps)
	c.JSON(http.StatusOK, ps)
}

// verifyShareLink 校验分享链接签名与有效期（公开 H5 接口共用）。
// sig = HMAC-SHA256(h5Secret, exp|to|ids|path)，与 h5/index.html 的 verifyLink 保持一致。
// 校验通过返回 exp/to/ids 原始字符串；失败时已写入响应。
func (h *Handler) verifyShareLink(c *gin.Context) (expStr, to, ids string, ok bool) {
	expStr = c.Query("exp")
	sig := c.Query("sig")
	to = c.Query("to")
	ids = c.Query("ids")
	path := c.Query("path")

	exp, err := strconv.ParseInt(expStr, 10, 64)
	if err != nil || exp < 0 || sig == "" {
		resp.BadRequest(c, "分享链接参数缺失")
		return "", "", "", false
	}
	// exp == 0 表示「长期有效」，不做过期判断（签名仍校验）
	if exp > 0 && time.Now().UnixMilli() > exp {
		resp.Error(c, http.StatusForbidden, "分享链接已过期")
		return "", "", "", false
	}
	mac := hmac.New(sha256.New, []byte(h.h5Secret))
	mac.Write([]byte(expStr + "|" + to + "|" + ids + "|" + path))
	expect := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expect), []byte(sig)) {
		resp.Error(c, http.StatusForbidden, "分享链接校验失败")
		return "", "", "", false
	}
	return expStr, to, ids, true
}

// H5Workers 供手机端 H5 分享链接拉取人员管理列表（同 H5Projects，签名校验，不走 JWT）。
// 只返回 id/姓名/工种，避免电话、身份证等敏感信息经公开链接外泄。
func (h *Handler) H5Workers(c *gin.Context) {
	if _, _, _, ok := h.verifyShareLink(c); !ok {
		return
	}
	var ws []model.Worker
	h.db.Model(&model.Worker{}).Where("active = ?", true).
		Order("pinyin_full ASC, id ASC").
		Find(&ws)
	type h5Worker struct {
		ID            uint   `json:"id"`
		Name          string `json:"name"`
		JobType       string `json:"jobType"`
		PinyinFull    string `json:"pinyinFull"`    // 拼音全拼（供 H5 下拉搜索）
		PinyinInitial string `json:"pinyinInitial"` // 拼音首字母（供 H5 下拉搜索）
	}
	out := make([]h5Worker, 0, len(ws))
	for _, w := range ws {
		out = append(out, h5Worker{ID: w.ID, Name: w.Name, JobType: w.JobType, PinyinFull: w.PinyinFull, PinyinInitial: w.PinyinInitial})
	}
	c.JSON(http.StatusOK, out)
}
