package refresh

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/signaux-faibles/datapi/src/db"
	"github.com/spf13/viper"
	"net/http"
)

func StartHandler(c *gin.Context) {
	refreshScriptPath := viper.GetString("refreshScript")
	id := StartRefreshScript(context.Background(), db.Get(), refreshScriptPath)
	c.JSON(http.StatusOK, gin.H{"refreshUuid": id.String()})
}

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

func LastHandler(c *gin.Context) {
	last := FetchLastRefreshState()
	c.JSON(http.StatusOK, last)
}

func ListHandler(c *gin.Context) {
	param := c.Param("status")
	if len(param) <= 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "il manque le paramètre 'status'"})
		return
	}
	last := FetchRefreshWithState(Status(param))
	c.JSON(http.StatusOK, last)
}
