package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"lineng/internal/model"
	"lineng/pkg/jwt"
	"lineng/pkg/resp"
)

// Login 账号密码登录，签发 JWT。
func (h *Handler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "参数错误")
		return
	}
	var u model.User
	if err := h.db.Where("username = ?", trimStr(req.Username)).First(&u).Error; err != nil {
		resp.Error(c, http.StatusUnauthorized, "账号或密码错误")
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password)) != nil {
		resp.Error(c, http.StatusUnauthorized, "账号或密码错误")
		return
	}
	tok, err := jwt.MakeToken(u.ID, u.Username)
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, "签发令牌失败")
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": tok, "username": u.Username})
}

// Me 返回当前登录用户信息。
func (h *Handler) Me(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"id": c.GetUint("uid"), "username": c.GetString("username")})
}
