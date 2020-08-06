package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func importHandler(c *gin.Context) {
	var params struct {
		entreprise    string
		etablissement string
	}
	c.Bind(&params)

	fmt.Println(params)
}
