package main

import (
	_ "github.com/signaux-faibles/datapi/docs"
	dalib "github.com/signaux-faibles/datapi/lib"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
)

func main() {

	router := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	config.AddAllowHeaders("Authorization")
	config.AddAllowMethods("GET", "POST")
	router.Use(cors.New(config))

	router.POST("/login", authMiddleware.LoginHandler)
	router.GET("/debug", debug)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/refresh", authMiddleware.MiddlewareFunc(), authMiddleware.RefreshHandler)
	data := router.Group("data")
	data.Use(authMiddleware.MiddlewareFunc())

	data.POST("/get/:bucket", get)
	data.POST("/put/:bucket", put)

	bind := viper.GetString("bind")
	router.Run(bind)
}

func debug(c *gin.Context) {

	c.JSON(200, dalib.Debug(dalib.CurrentBucketPolicies.SafeRead()))
}
func put(c *gin.Context) {
	bucket := c.Params.ByName("bucket")

	var objects []dalib.Object
	err := c.Bind(&objects)
	if err != nil {
		c.JSON(400, "bad request: "+err.Error())
		return
	}

	u, ok := c.Get("id")
	if !ok {
		c.JSON(400, "malformed token")
	}

	user, ok := u.(*dalib.User)
	if !ok {
		c.JSON(400, "malformed token")
	}

	err = dalib.Insert(objects, bucket, user.Scope, user.Email)
	if err == dalib.ErrInvalidScope {
		c.JSON(403, err.Error())
		return
	} else if err != nil {
		c.JSON(500, "Server error: "+err.Error())
		return
	}
	c.JSON(200, "ok")

}

func get(c *gin.Context) {
	var params dalib.QueryParams

	err := c.ShouldBind(&params)
	if err != nil {
		c.JSON(400, err.Error())
	}

	bucket := c.Params.ByName("bucket")

	u, _ := c.Get("id")
	user := u.(*dalib.User)
	scope := user.Scope

	data, err := dalib.Query(bucket, params, scope, true, nil)

	if err != nil {
		c.JSON(500, err.Error())
		return
	}

	if len(data) == 0 {
		c.JSON(404, "no data")
		return
	}

	c.JSON(200, data)
}
