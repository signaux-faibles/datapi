package campaign

import (
	"datapi/pkg/core"
	"github.com/gin-gonic/gin"
)

func ConfigureEndpoint(kanbanService core.KanbanService) func(campaignRoute *gin.RouterGroup) {
	return func(campaignRoute *gin.RouterGroup) {
		campaignRoute.GET("/list", listCampaignsHandler) // 1
		campaignRoute.GET("/actions/pending/:campaignID", pendingHandler(kanbanService))
		campaignRoute.GET("/actions/mine/:campaignID", myActionsHandler(kanbanService))
		campaignRoute.GET("/actions/taken/:campaignID", takenActionsHandler(kanbanService))
		campaignRoute.GET("/take/:campaignID/:campaignEtablissementID", takeHandler)
		campaignRoute.POST("/success/:campaignID/:campaignEtablissementID", actionHandlerFunc("success"))
		campaignRoute.POST("/cancel/:campaignID/:campaignEtablissementID", actionHandlerFunc("cancel"))
		campaignRoute.GET("/stream", HeadersMiddleware(), stream.serveHTTP(), streamHandler)
		campaignRoute.POST("/upsertcard", upsertCardHandler(kanbanService))
		campaignRoute.POST("/withdraw/:campaignID/:campaignEtablissementID", withdrawHandler)
		campaignRoute.POST("/checksirets/:campaignID", checkSiretsHandler)
	}
}
