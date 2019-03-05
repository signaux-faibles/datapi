package main

import (
	"fmt"
	"time"

	dalib "github.com/signaux-faibles/datapi/lib"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func main() {

	router := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	config.AddAllowHeaders("Authorization")
	config.AddAllowMethods("GET", "POST")
	router.Use(cors.New(config))

	router.POST("/login", authMiddleware.LoginHandler)
	router.GET("/policies", debug)
	data := router.Group("data")
	data.Use(authMiddleware.MiddlewareFunc())

	data.POST("/get/:bucket", get)
	data.POST("/put/:bucket", put)

	bind := viper.GetString("bind")
	router.Run(bind)
}

func debug(c *gin.Context) {
	dalib.CurrentBucketPolicies = dalib.LoadPolicies(time.Now())

	fmt.Println(dalib.CurrentBucketPolicies)
}

func put(c *gin.Context) {
	bucket := c.Params.ByName("bucket")

	var objects []dalib.Object
	err := c.Bind(&objects)
	if err != nil {
		c.JSON(400, "Bad request: "+err.Error())
		return
	}

	u, _ := c.Get("id")
	user := u.(*dalib.User)

	err = dalib.Insert(objects, bucket, *user)
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
	var key = make(map[string]string)
	err := c.ShouldBind(&key)
	if err != nil {
		c.JSON(400, err.Error())
	}

	bucket := c.Params.ByName("bucket")

	u, _ := c.Get("id")
	user := u.(*dalib.User)

	scope := user.Scope

	dateQuery := time.Now()

	data, err := dalib.Query(bucket, key, scope, dateQuery)

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
