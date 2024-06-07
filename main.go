/*
 * @Author: flwfdd
 * @Date: 2024-06-06 15:47:18
 * @LastEditTime: 2024-06-06 21:51:08
 * @Description:
 * _(:з」∠)_
 */
package main

import (
	"BIT-Helper/database"
	"BIT-Helper/router"
	"BIT-Helper/util/config"
	"fmt"
	"time"

	"github.com/gin-contrib/cors"
	limits "github.com/gin-contrib/size"
	"github.com/gin-gonic/gin"
)

// 服务，启动！
func main() {
	config.Init()
	database.Init()

	if config.Config.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}
	app := gin.Default()
	app.Use(limits.RequestSizeLimiter(config.Config.Saver.MaxSize << 20))
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{"Content-Type", "fake-cookie"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
		// ExposeHeaders:    []string{"Content-Length"},
		// AllowCredentials: true,
		// AllowOriginFunc: func(origin string) bool {
		// 	return true
		// },
		MaxAge: 12 * time.Hour,
	}))
	router.SetRouter(app)
	fmt.Println("BIT101-Helper will run on port " + config.Config.Port)
	app.Run(":" + config.Config.Port)
}
