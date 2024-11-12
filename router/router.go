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
	// 商品模块
	goods := router.Group("/goods")
	{
		goods.GET("", controller.GoodsList)
		goods.POST("", middleware.CheckLogin(true), controller.GoodsPost)
		goods.GET("/:id", controller.GoodsGet)
		goods.PUT("/:id", middleware.CheckLogin(true), controller.GoodsPut)
		goods.DELETE("/:id", middleware.CheckLogin(true), controller.GoodsDelete)
	}
	// 话题模块
	topic := router.Group("/topic")
	{
		topic.GET("/:type", controller.TopicList)
		topic.POST("/create", middleware.CheckLogin(true), controller.TopicPost)
		topic.GET("/:type/:id", controller.TopicGet)
		topic.PUT("/update/:id", middleware.CheckLogin(true), controller.TopicPut)
		topic.DELETE("/delete/:id", middleware.CheckLogin(true), controller.TopicDelete)
		topic.POST("/like/:id", middleware.CheckLogin(true), controller.TopicLike) // 点赞话题
	}
	// 订单模块
	order := router.Group("/orders")
	{
		order.GET("", middleware.CheckLogin(true), controller.OrderList)
		order.POST("", middleware.CheckLogin(true), controller.OrderPost)
		order.GET("/:id", middleware.CheckLogin(true), controller.OrderGet)
		order.PUT("/:id", middleware.CheckLogin(true), controller.OrderPut)
		order.POST("/review/:id", middleware.CheckLogin(true), controller.OrderReview)
	}
	// 消息模块
	chat := router.Group("/chats")
	{
		chat.GET("", middleware.CheckLogin(true), controller.GetAllChats)
		chat.POST("/:id", middleware.CheckLogin(true), controller.ChatsPost)
		chat.GET("/:id", middleware.CheckLogin(true), controller.GetChatsById)
	}
	// 操作反馈模块
	reaction := router.Group("/reaction")
	{
		reaction.POST("/like", middleware.CheckLogin(true), controller.ReactionLike)
		reaction.POST("/comments", middleware.CheckLogin(true), controller.ReactionComment)
		reaction.GET("/comments", middleware.CheckLogin(false), controller.ReactionCommentList)
		reaction.DELETE("/comments/:id", middleware.CheckLogin(true), controller.ReactionCommentDelete)
	}
	// 上传模块
	upload := router.Group("/upload")
	{
		upload.POST("/image", middleware.CheckLogin(true), controller.ImageUpload)
	}
}
