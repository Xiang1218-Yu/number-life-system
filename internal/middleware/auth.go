package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"number-life-system/internal/service"
	"strings"
)

func Auth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "需要登录"})
			return
		}
		id, err := service.ParseUserID(strings.TrimPrefix(header, "Bearer "), secret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "登录已失效"})
			return
		}
		c.Set("user_id", id)
		c.Next()
	}
}
func UserID(c *gin.Context) uint {
	value, _ := c.Get("user_id")
	id, _ := value.(uint)
	if id == 0 {
		return 0
	}
	return id
}
