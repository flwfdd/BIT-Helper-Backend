package middleware

import (
	"BIT-Helper/util/config"
	"BIT-Helper/util/jwt"
	"strconv"

	"github.com/gin-gonic/gin"
)

// 验证用户是否登录
func CheckLogin(strict bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("fake-cookie")
		uid, ok, admin := jwt.VeirifyUserToken(token, config.Config.Key)
		if ok {
			uid_uint, err := strconv.ParseUint(uid, 10, 32)
			if err != nil {
				c.JSON(500, gin.H{"msg": "获取用户ID错误Orz"})
				c.Abort()
				return
			}
			// if controller.CheckBan(uint(uid_uint)) {
			// 	t, _ := controller.ParseTime(database.BanMap[uint(uid_uint)].Time)
			// 	c.JSON(403, gin.H{"msg": "您已被关小黑屋Orz,解封时间：" + t.Format("2006-01-02 15:04:05")})
			// 	c.Abort()
			// 	return
			// }
			c.Set("uid", uid)
			c.Set("uid_uint", uint(uid_uint))
			c.Set("admin", admin)
		} else if strict {
			c.JSON(401, gin.H{"msg": "请先登录awa"})
			c.Abort()
		}
	}
}

// CheckAdmin 验证用户是否为admin/super
func CheckAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !c.GetBool("admin") && !c.GetBool("super") {
			c.JSON(403, gin.H{"msg": "权限不足awa"})
			c.Abort()
		}
	}
}
