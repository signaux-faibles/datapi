package main

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"

	dalib "./lib"

	jwt "github.com/appleboy/gin-jwt"
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

	data := router.Group("data")
	data.Use(authMiddleware.MiddlewareFunc())

	router.POST("/get/:bucket", get)
	router.POST("/put/:bucket", put)

	router.GET("/hash/:password", hash)
	router.GET("/jwt", readJwt)
	bind := viper.GetString("bind")
	router.Run(bind)
}

func put(c *gin.Context) {
	bucket := c.Params.ByName("bucket")

	var objects []dalib.Object
	err := c.Bind(&objects)

	if err != nil {
		c.JSON(400, "Bad request: "+err.Error())
		return
	}

	err = dalib.Insert(objects, bucket)
	if err != nil {
		c.JSON(500, "Server error: "+err.Error())
		return
	}
	c.JSON(200, "ok")
}

func get(c *gin.Context) {
	bucket := c.Params.ByName("bucket")
	fmt.Println(bucket)
	claims := jwt.ExtractClaims(c)
	fmt.Println(claims)
	scope := dalib.Tags{"admin", "user"}
	key := dalib.Map{
		"type":   "testObject",
		"bucket": bucket,
	}

	fmt.Println(key)
	r, err := dalib.Query(key, scope)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}
	c.JSON(200, r)
}

func hash(c *gin.Context) {
	password := c.Params.ByName("password")
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		c.JSON(500, err.Error())
	}
	c.JSON(200, string(hash[:]))
}

func readJwt(c *gin.Context) {
	fmt.Println(identityHandler(c))
}
