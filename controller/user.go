package controller

import (
	"BIT-Helper/database"
	"BIT-Helper/util/config"
	"BIT-Helper/util/jwt"
	"BIT-Helper/util/webvpn"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

// 转换用户信息
func CleanUser(old_user database.User) database.User {
	user := old_user
	// 转换头像链接
	user.Avatar = GetImageUrl(user.Avatar)
	return user
}

// 获取用户信息
func GetUserAPI(uid int) database.User {
	return GetUserAPIMap(map[int]bool{uid: true})[uid]
}

func GetUserAPIList(uid_list []int) []database.User {
	uid_map := make(map[int]bool)
	for _, uid := range uid_list {
		uid_map[uid] = true
	}
	user_api_map := GetUserAPIMap(uid_map)
	user_api_list := make([]database.User, 0, len(uid_list))
	for _, uid := range uid_list {
		user_api_list = append(user_api_list, user_api_map[uid])
	}
	return user_api_list
}

// 批量获取用户信息
func GetUserAPIMap(uid_map map[int]bool) map[int]database.User {
	out := make(map[int]database.User)
	uid_list := make([]int, 0)
	for uid := range uid_map {
		uid_list = append(uid_list, uid)
	}

	var users []database.User
	database.DB.Where("id IN ?", uid_list).Find(&users)
	for _, user := range users {
		out[int(user.ID)] = CleanUser(user)
	}
	return out
}

// 登录请求结构
type UserLoginQuery struct {
	Sid      string `json:"sid" binding:"required"`      // 学号
	Password string `json:"password" binding:"required"` // 密码
}

// 注册或登录
func UserLogin(c *gin.Context) {
	var query UserLoginQuery
	if err := c.ShouldBind(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}

	// 初始化Webvpn
	data, err := webvpn.InitLogin()
	if err != nil {
		c.JSON(500, gin.H{"msg": "初始化登陆失败Orz"})
		return
	}

	// 登录Webvpn
	encryptedPassword, err := webvpn.EncryptPassword(query.Password, data.Salt)
	if err != nil {
		c.JSON(500, gin.H{"msg": "密码加密失败Orz"})
		return
	}
	err = webvpn.Login(query.Sid, encryptedPassword, data.Execution, data.Cookie, "")
	if err != nil {
		c.JSON(500, gin.H{"msg": "统一身份认证失败Orz"})
		return
	}

	// 未注册过则注册
	var user database.User
	if err := database.DB.Limit(1).Find(&user, "sid = ?", query.Sid).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		return
	}
	if user.ID == 0 {
		// 未注册过
		user.Sid = query.Sid
		user.Nickname = query.Sid
		user.Intro = "BITer " + query.Sid
		user.Score = 100
		user.Identity = database.Identity_Normal
		if err := database.DB.Create(&user).Error; err != nil {
			c.JSON(500, gin.H{"msg": "数据库错误Orz"})
			return
		}
	}

	token := jwt.GetUserToken(fmt.Sprint(user.ID), config.Config.LoginExpire, config.Config.Key, int(user.Identity))
	c.JSON(200, gin.H{"msg": "登录成功OvO", "fake_cookie": token})
}

// UserGetInfo 获取用户信息
func UserGetInfo(c *gin.Context) {
	id_str := c.Param("id")
	var uid uint
	if id_str == "" || id_str == "0" {
		// 获取自己的信息
		uid = c.GetUint("uid_uint")
		if uid == 0 {
			c.JSON(401, gin.H{"msg": "请先登录awa"})
			return
		}
	} else {
		uid_, err := strconv.ParseUint(id_str, 10, 32)
		if err != nil {
			c.JSON(400, gin.H{"msg": "参数错误awa"})
			return
		}
		var user database.User
		database.DB.Limit(1).Find(&user, "id = ?", uid_)
		if user.ID == 0 {
			c.JSON(404, gin.H{"msg": "用户不存在Orz"})
			return
		}
		uid = uint(uid_)
	}
	c.JSON(200, GetUserAPI(int(uid)))
}

// 修改用户信息请求结构
type UserSetInfoQuery struct {
	Nickname  string `json:"nickname"`   // 昵称
	AvatarMid string `json:"avatar_mid"` // 头像
	Intro     string `json:"intro"`      // 格言 简介
	Phone     int    `json:"phone"`      // 电话
}

// 修改用户信息
func UserSetInfo(c *gin.Context) {
	var query UserSetInfoQuery
	if err := c.ShouldBind(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}
	uid := c.GetString("uid")
	var user database.User
	if err := database.DB.Limit(1).Find(&user, "id = ?", uid).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		return
	}
	if user.ID == 0 {
		c.JSON(500, gin.H{"msg": "用户不存在Orz"})
		return
	}

	if query.Nickname != "" {
		user_ := database.User{}
		if err := database.DB.Limit(1).Find(&user_, "nickname = ?", query.Nickname).Error; err != nil {
			c.JSON(500, gin.H{"msg": "数据库错误Orz"})
			return
		}
		if user_.ID != 0 && user_.ID != user.ID {
			c.JSON(500, gin.H{"msg": "昵称冲突Orz"})
			return
		}
		user.Nickname = query.Nickname
	}
	if query.AvatarMid != "" {
		// 验证图片是否存在
		avatar := database.Image{}
		if err := database.DB.Limit(1).Find(&avatar, "mid = ?", query.AvatarMid).Error; err != nil {
			c.JSON(500, gin.H{"msg": "数据库错误Orz"})
			return
		}
		if avatar.ID == 0 {
			c.JSON(500, gin.H{"msg": "头像图片无效Orz"})
			return
		}
		user.Avatar = query.AvatarMid
	}
	if query.Intro != "" {
		user.Intro = query.Intro
	}
	if query.Phone != 0 {
		user.Phone = query.Phone
	}
	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		return
	}
	c.JSON(200, gin.H{"msg": "修改成功OvO"})
}
