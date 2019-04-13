package main

import (
	_ "github.com/signaux-faibles/datapi/docs"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	dalib "github.com/signaux-faibles/datapi/lib"

	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
)

func main() {
	dalib.Warmup()

	router := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	config.AddAllowHeaders("Authorization")
	config.AddAllowMethods("GET", "POST")
	router.Use(cors.New(config))

	router.POST("/login", authMiddleware.LoginHandler)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/refresh", authMiddleware.MiddlewareFunc(), authMiddleware.RefreshHandler)
	router.GET("/ws/:token", wsTokenMiddleWare, authMiddleware.MiddlewareFunc(), wsHandler)

	data := router.Group("data")

	data.Use(authMiddleware.MiddlewareFunc())
	data.POST("/get/:bucket", get)
	data.POST("/put/:bucket", put)

	bind := viper.GetString("bind")
	router.Run(bind)
}
