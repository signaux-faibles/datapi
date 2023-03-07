package refresh

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/signaux-faibles/datapi/src/db"
	"github.com/spf13/viper"
)

func StartHandler(c *gin.Context) {
	refreshScriptPath := viper.GetString("refreshScript")
	id := StartRefreshScript(context.Background(), db.Db(), refreshScriptPath)
	c.JSON(200, fmt.Sprintf("Le refresh a l'UUID %s", id))
}

func StatusHandler(c *gin.Context) {
	param := c.Params.ByName("uuid")
	id, err := uuid.Parse(param)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	state, err := Fetch(id)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	c.JSON(200, state)
}

func LastHandler(c *gin.Context) {
	last := FetchLastRefreshState()
	c.JSON(200, last)
}

func ListHandler(c *gin.Context) {
	param := c.Params.ByName("status")
	last := FetchRefreshWithState(Status(param))
	c.JSON(200, last)
}
