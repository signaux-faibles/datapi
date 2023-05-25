package campaigns

import (
	"datapi/pkg/core"
	"github.com/gin-gonic/gin"
)

func ConfigureEndpoint(path string, api *gin.Engine) {
	endpoint := api.Group(path, core.AdminAuthMiddleware)
	endpoint.GET("/list", getCampaigns) // 1
	endpoint.GET("/get/:campaignID")
}

func getCampaigns(c *gin.Context) {
	var s core.Session
	s.Bind(c)

}
