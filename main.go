package main

import (
	"fmt"

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
	router.GET("/data/:bucket/:id", get)
	router.POST("/data/:bucket", put)
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
