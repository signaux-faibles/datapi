package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	dalib "github.com/signaux-faibles/datapi-lib"
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
	router.PUT("/data/:bucket", put)
	bind := viper.GetString("bind")
	router.Run(bind)
}

func put(c *gin.Context) {
	var o dalib.Object
	a := "success"
	o.Key = dalib.Map{
		"test": &a,
	}

	o.Value = dalib.PropertyMap{
		"test":  "coucou",
		"test2": "recoucou",
	}

	id, err := dalib.Insert(o)
	fmt.Println(id, err)
}

func get(c *gin.Context) {
	r, err := dalib.Query(map[string]*string{})
	fmt.Println(err)
	c.JSON(200, r)
}
