package campaign

import (
	"datapi/pkg/core"
	"datapi/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/signaux-faibles/libwekan"
)

func ConfigureEndpoint(path string, api *gin.Engine) {
	endpoint := api.Group(path, core.AuthMiddleware(), core.LogMiddleware)
	endpoint.GET("/list", listeCampaignsHandler) // 1
	endpoint.GET("/get/:campaignID")
}

func listeCampaignsHandler(c *gin.Context) {
	var s core.Session
	s.Bind(c)

	campaigns, err := selectAllCampaignsFromDB(c)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}

	kanbanConfig := core.Kanban.LoadConfigForUser(libwekan.Username(s.Username))
	c.JSON(200, campaignsForUser(kanbanConfig.Boards, campaigns))
}
