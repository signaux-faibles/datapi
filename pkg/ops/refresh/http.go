// Package refresh : contient tout le code qui concerne l'exécution d'un `Refresh` Datapi,
// c'est-à-dire l'exécution du script sql configuré
package refresh

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"datapi/pkg/db"
	"datapi/pkg/utils"
)

// ConfigureEndpoint configure le endpoint du package `refresh`
func ConfigureEndpoint(refreshRoute *gin.RouterGroup) {
	refreshRoute.GET("/start", startHandler)
	refreshRoute.GET("/status/:uuid", statusHandler)
	refreshRoute.GET("/last", lastHandler)
	refreshRoute.GET("/list/:status", listHandler)
}

// startHandler : point d'entrée de l'API qui démarre un nouveau `Refresh` et retourne son `UUID`
func startHandler(c *gin.Context) {
	refresh := StartRefreshScript(context.Background(), db.Get())
	c.JSON(http.StatusOK, refresh)
}

// statusHandler : point d'entrée de l'API qui retourne les infos d'un `Refresh` depuis son `UUID`
func statusHandler(c *gin.Context) {
	param := c.Param("uuid")
	if len(param) <= 0 {
		c.JSON(http.StatusBadRequest, "il manque le paramètre 'uuid'")
		return
	}
	id, err := uuid.Parse(param)
	if err != nil {
		utils.AbortWithError(c, err) // nolint: errcheck
		return
	}
	refresh, err := Fetch(id)
	if err != nil {
		utils.AbortWithError(c, err) // nolint: errcheck
		return
	}
	c.JSON(http.StatusOK, refresh)
}

// lastHandler : point d'entrée de l'API qui retourne le dernier `Refresh` démarré
func lastHandler(c *gin.Context) {
	last := FetchLast()
	c.JSON(http.StatusOK, last)
}

// listHandler : point d'entrée de l'API qui retourne les `Refresh` selon le `status` passé en paramètre
func listHandler(c *gin.Context) {
	param := c.Param("status")
	if len(param) <= 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "il manque le paramètre 'status'"})
		return
	}
	last := FetchRefreshsWithState(Status(param))
	c.JSON(http.StatusOK, last)
}
