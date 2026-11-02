// Package jwt 负责 JWT 令牌的签发与校验中间件。
package jwt

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	jwtlib "github.com/golang-jwt/jwt/v5"

	"lineng/pkg/resp"
)

var secret = []byte(envOr("JWT_SECRET", "lineng-jwt-secret-2026"))

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// Claims 自定义 JWT 载荷。
type Claims struct {
	UserID   uint   `json:"uid"`
	Username string `json:"username"`
	jwtlib.RegisteredClaims
}

// MakeToken 为用户签发 7 天有效期的令牌。
func MakeToken(userID uint, username string) (string, error) {
	claims := &Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwtlib.RegisteredClaims{
			ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwtlib.NewNumericDate(time.Now()),
		},
	}
	return jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims).SignedString(secret)
}

// Middleware 鉴权中间件：校验 Bearer Token 并把 uid / username 注入上下文。
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			resp.Error(c, http.StatusUnauthorized, "未登录")
			return
		}
		claims := &Claims{}
		tok, err := jwtlib.ParseWithClaims(strings.TrimPrefix(h, "Bearer "), claims, func(t *jwtlib.Token) (interface{}, error) {
			return secret, nil
		})
		if err != nil || !tok.Valid {
			resp.Error(c, http.StatusUnauthorized, "登录已过期，请重新登录")
			return
		}
		c.Set("uid", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}
