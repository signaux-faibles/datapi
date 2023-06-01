package campaign

import (
	"datapi/pkg/core"
	"datapi/pkg/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/signaux-faibles/libwekan"
	"strconv"
)

func ConfigureEndpoint(path string, api *gin.Engine) {
	endpoint := api.Group(path, core.AuthMiddleware(), core.LogMiddleware)
	endpoint.GET("/list", listCampaignsHandler) // 1
	endpoint.GET("/get/:campaignID", selectCampaignHandler)
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
	campaignIDParam := c.Param("campaignID")
	campaignID, err := strconv.Atoi(campaignIDParam)
	campaign, err := selectCampaignDetailsWithCampaignIDAndZone(c, CampaignID(campaignID), []string{"79", "91"})
	fmt.Println("hello")
	if err != nil {
		c.JSON(500, err.Error())
		return
	}

	c.JSON(200, campaign)
}
