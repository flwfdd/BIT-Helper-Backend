package router

import (
	"BIT-Helper/controller"
	"BIT-Helper/middleware"

	"github.com/gin-gonic/gin"
)

// 配置路由
func SetRouter(router *gin.Engine) {
	router.GET("/", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"msg": "Hello BIT-Helper!"})
	})
	// 用户模块
	user := router.Group("/user")
	{
		user.POST("/login", controller.UserLogin)
		user.GET("/info/:id", middleware.CheckLogin(false), controller.UserGetInfo)
		user.PUT("/info", middleware.CheckLogin(true), controller.UserSetInfo)
	}
	// 上传模块
	upload := router.Group("/upload")
	{
		upload.POST("/image", middleware.CheckLogin(true), controller.ImageUpload)
	}
}
