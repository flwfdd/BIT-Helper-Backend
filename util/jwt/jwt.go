/*
 * @Author: flwfdd
 * @Date: 2024-06-06 16:04:35
 * @LastEditTime: 2024-06-06 17:12:19
 * @Description:
 * _(:з」∠)_
 */
package jwt

import (
	"BIT-Helper/database"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// 用于验证用户登录状态
type UserClaims struct {
	Id     string `json:"id"`
	Expire int64  `json:"exp"`
	Admin  bool   `json:"admin"`
	jwt.RegisteredClaims
}

// 生成用户token 也可将验证码作为key来验证
func GetUserToken(id string, expire int64, key string, identity int) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, UserClaims{
		id,
		time.Now().Unix() + expire,
		identity == database.Identity_Admin,
		jwt.RegisteredClaims{},
	})
	tokenString, err := token.SignedString([]byte(key))
	if err != nil {
		panic(err)
	}
	return tokenString
}

// 验证用户 返回用户id 是否成功 是否是管理员
func VeirifyUserToken(tokenString string, key string) (string, bool, bool) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(key), nil
	})
	if err != nil {
		println(err.Error())
		return "", false, false
	}

	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid && claims.Expire > time.Now().Unix() {
		return claims.Id, true, claims.Admin
	} else {
		return "", false, false
	}
}
