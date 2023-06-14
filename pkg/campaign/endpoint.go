package campaign

import (
	"github.com/gin-gonic/gin"
)

func ConfigureEndpoint(campaignRoute *gin.RouterGroup) {
	campaignRoute.GET("/list", listCampaignsHandler) // 1
	campaignRoute.GET("/pending/:campaignID", pendingHandler)
	campaignRoute.GET("/myactions/:campaignID", myActionsHandler)
	campaignRoute.GET("/allactions/:campaignID", allActionsHandler)
	campaignRoute.GET("/take/:campaignID/:campaignEtablissementID", takePendingHandler)
	campaignRoute.POST("/success/:campaignID/:campaignEtablissementID", actionHandlerFunc("success"))
	campaignRoute.POST("/cancel/:campaignID/:campaignEtablissementID", actionHandlerFunc("cancel"))
	campaignRoute.GET("/stream/:campaignID", HeadersMiddleware(), stream.serveHTTP(), streamHandler)
}
