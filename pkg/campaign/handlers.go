package campaign

import (
	"datapi/pkg/core"
	"datapi/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/signaux-faibles/libwekan"
	"strconv"
)

func ConfigureEndpoint(path string, api *gin.Engine) {
	endpoint := api.Group(path, core.AuthMiddleware(), core.LogMiddleware)
	endpoint.GET("/list", listCampaignsHandler) // 1
	endpoint.GET("/get/:campaignID")
}

func listCampaignsHandler(c *gin.Context) {
	var s core.Session
	s.Bind(c)

	boards := core.Kanban.SelectBoardsForUsername(libwekan.Username(s.Username))
	slugs := utils.Convert(boards, libwekan.ConfigBoard.Slug)
	campaigns, err := selectMatchingCampaigns(c, slugs, s.Roles)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}

	c.JSON(200, campaigns)
}

func selectCampaignHandler(c *gin.Context) {
	var s core.Session
	s.Bind(c)
	campaignIDParam := c.Param("id")
	campaignID, err := strconv.Atoi(campaignIDParam)
	campaign, err := selectCampaignDetailsWithCampaignIDAndZone(c, CampaignID(campaignID), []string{"70", "90"})

	if err != nil {
		c.JSON(500, err.Error())
		return
	}

	c.JSON(200, campaign)
}
