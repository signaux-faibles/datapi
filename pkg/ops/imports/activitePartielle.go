package imports

import (
	_ "embed"
	"net/http"

	"github.com/gin-gonic/gin"

	"datapi/pkg/db"
	"datapi/pkg/ops/scripts"
)

//go:embed sql/activitePartielle.sql
var sqlRefreshActivitePartielle string

var RefreshActivitePartielle = scripts.Script{
	Label: "rafraîchit les données d'activité partielle",
	SQL:   sqlRefreshActivitePartielle,
}

func refreshActivitePartielleHandler(c *gin.Context) {
	refresh := scripts.StartRefreshScript(c.Request.Context(), db.Get(), RefreshActivitePartielle)
	c.JSON(http.StatusOK, refresh)
}
