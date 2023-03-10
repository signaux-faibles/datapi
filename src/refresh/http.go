// Copyright 2023 The Signaux Faibles team
// license that can be found in the LICENSE file.
//
// ce package contient tout le code qui concerne l'exécution d'un `Refresh` Datapi,
// c'est à dire l'exécution du script sql configuré

package refresh

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/signaux-faibles/datapi/src/db"
	"github.com/spf13/viper"
	"net/http"
)

// StartHandler : point d'entrée de l'API qui démarre un nouveau `Refresh` et retourne son `UUID`
func StartHandler(c *gin.Context) {
	refreshScriptPath := viper.GetString("refreshScript")
	id := StartRefreshScript(context.Background(), db.Get(), refreshScriptPath)
	c.JSON(http.StatusOK, gin.H{"refreshUuid": id.String()})
}

// StatusHandler : point d'entrée de l'API qui retourne les infos d'un `Refresh` depuis son `UUID`
func StatusHandler(c *gin.Context) {
	param := c.Param("uuid")
	if len(param) <= 0 {
		c.JSON(http.StatusBadRequest, "il manque le paramètre 'uuid'")
		return
	}
	id, err := uuid.Parse(param)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	state, err := Fetch(id)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, state)
}

// LastHandler : point d'entrée de l'API qui retourne le dernier `Refresh` démarré
func LastHandler(c *gin.Context) {
	last := FetchLast()
	c.JSON(http.StatusOK, last)
}

// ListHandler : point d'entrée de l'API qui retourne les `Refresh` selon le `status` passé en paramètre
func ListHandler(c *gin.Context) {
	param := c.Param("status")
	if len(param) <= 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "il manque le paramètre 'status'"})
		return
	}
	last := FetchRefreshsWithState(Status(param))
	c.JSON(http.StatusOK, last)
}
