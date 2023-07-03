package campaign

import (
	"github.com/gin-gonic/gin"
)

func ConfigureEndpoint(campaignRoute *gin.RouterGroup) {
	campaignRoute.GET("/list", listCampaignsHandler) // 1
	campaignRoute.GET("/actions/pending/:campaignID", pendingHandler)
	campaignRoute.GET("/actions/mine/:campaignID", myActionsHandler)
	campaignRoute.GET("/actions/taken/:campaignID", takenActionsHandler)
	campaignRoute.GET("/take/:campaignID/:campaignEtablissementID", takeHandler)
	campaignRoute.POST("/success/:campaignID/:campaignEtablissementID", actionHandlerFunc("success"))
	campaignRoute.POST("/cancel/:campaignID/:campaignEtablissementID", actionHandlerFunc("cancel"))
	campaignRoute.GET("/stream", HeadersMiddleware(), stream.serveHTTP(), streamHandler)
}
