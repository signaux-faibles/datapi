package health

import (
	"runtime"

	"github.com/gin-gonic/gin"
)

type healthResponse struct {
	Goroutines int `json:"goroutines"`
}

func ConfigureEndpoint(endpoint *gin.RouterGroup) {
	endpoint.GET("", getHealth)
}

func getHealth(c *gin.Context) {
	response := healthResponse{
		Goroutines: runtime.NumGoroutine(),
	}

	c.JSON(200, response)
} 