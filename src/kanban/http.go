package kanban

import (
	"github.com/gin-gonic/gin"
	"github.com/signaux-faibles/datapi/src/core"
	"github.com/signaux-faibles/libwekan"
	"net/http"
)

// ConfigureEndpoint configure l'endpoint du package `ops`
func ConfigureEndpoint(path string, api *gin.Engine) {
	kanban := api.Group(path, core.AuthMiddleware(), core.LogMiddleware)
	kanban.GET("/config", kanbanConfigHandler)
}

func kanbanConfigHandler(c *gin.Context) {
	var s core.Session
	s.Bind(c)
	kanbanConfig := kanbanConfigForUser(libwekan.Username(s.Username))
	c.JSON(http.StatusOK, kanbanConfig)
}
