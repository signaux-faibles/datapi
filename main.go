package main

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"

	dalib "./lib"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func main() {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	router := gin.Default()

	router.POST("/login", authMiddleware.LoginHandler)

	data := router.Group("data")
	data.Use(authMiddleware.MiddlewareFunc())

	data.POST("/get", get)
	router.POST("/put", put)
	router.GET("/hash/:password", hash)
	router.GET("/jwt", readJwt)
	bind := viper.GetString("bind")
	router.Run(bind)
}

func put(c *gin.Context) {
	var objects []dalib.Object
	err := c.Bind(&objects)
	if err != nil {
		c.JSON(400, "Bad request: "+err.Error())
		return
	}

	err = dalib.Insert(objects)
	if err != nil {
		c.JSON(500, "Server error: "+err.Error())
		return
	}
	c.JSON(200, "ok")

}

func get(c *gin.Context) {
	r, err := dalib.Query(dalib.Map{}, dalib.Tags{})
	fmt.Println(err)
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
