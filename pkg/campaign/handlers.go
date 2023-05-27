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

	boards := core.Kanban.SelectBoardsForUsername(libwekan.Username(s.Username))
	slugs := utils.Convert(boards, libwekan.ConfigBoard.Slug)
	campaigns, err := selectMatchingCampaigns(c, slugs)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}

	c.JSON(200, campaigns)
}
