package campaign

import (
	"github.com/gin-gonic/gin"
)

func ConfigureEndpoint(campaignRoute *gin.RouterGroup) {
	campaignRoute.GET("/list", listCampaignsHandler) // 1
	campaignRoute.GET("/pending/:campaignID", pendingHandler)
	campaignRoute.GET("/take/:campaignID/:campaignEtablissementID", takePendingHandler)
}
