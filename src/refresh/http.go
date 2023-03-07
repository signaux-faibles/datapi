package refresh

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/signaux-faibles/datapi/src/db"
	"github.com/spf13/viper"
	"net/http"
)

func StartHandler(c *gin.Context) {
	refreshScriptPath := viper.GetString("refreshScript")
	id := StartRefreshScript(context.Background(), db.Get(), refreshScriptPath)
	c.JSON(http.StatusOK, fmt.Sprintf("Le refresh a l'UUID %s", id))
}

func StatusHandler(c *gin.Context) {
	param := c.Params.ByName("uuid")
	if len(param) <= 0 {
		c.JSON(http.StatusBadRequest, "il manque le paramètre 'uuid'")
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
	param := c.Params.ByName("status")
	//param := c.Query("status")
	if len(param) <= 0 {
		c.JSON(http.StatusBadRequest, "il manque le paramètre 'status'")
	}
	last := FetchRefreshWithState(Status(param))
	c.JSON(http.StatusOK, last)
}
