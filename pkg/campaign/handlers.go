package campaign

import (
	"datapi/pkg/core"
	"github.com/gin-gonic/gin"
	"github.com/signaux-faibles/libwekan"
	"regexp"
)

func ConfigureEndpoint(campaignRoute *gin.RouterGroup) {
	campaignRoute.GET("/list", listCampaignsHandler) // 1
	campaignRoute.GET("/pending/:campaignID", pendingHandler)
}

func listCampaignsHandler(c *gin.Context) {
	var s core.Session
	s.Bind(c)
	boards := core.Kanban.SelectBoardsForUsername(libwekan.Username(s.Username))
	zone := zonesFromBoards(boards)
	boardIDs := idsFromBoards(boards)
	campaigns, err := selectMatchingCampaigns(c, zone, boardIDs)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}
	c.JSON(200, campaigns)
}

func matchConfigBoardSlugFilter(wekanDomainRegexp regexp.Regexp) func(libwekan.ConfigBoard) bool {
	return func(board libwekan.ConfigBoard) bool {
		return wekanDomainRegexp.MatchString(string(board.Slug()))
	}
}
