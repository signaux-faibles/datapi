package stats

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"datapi/pkg/utils"
)

type API struct {
	db StatsDB
}

func NewAPI(db StatsDB) *API {
	return &API{db: db}
}

func NewAPIFromConfiguration(ctx context.Context, connexionURL string) (*API, error) {
	statsDB, err := createStatsDBFromURL(ctx, connexionURL)
	if err != nil {
		return nil, errors.Wrap(err, "erreur lors de l'initialisation du module stats")
	}
	return NewAPI(statsDB), nil
}

func (api *API) ConfigureEndpoint(endpoint *gin.RouterGroup) {
	endpoint.GET("/since/:n/days", api.sinceDaysHandler)
}

func (api *API) sinceDaysHandler(c *gin.Context) {
	var since time.Time
	param := c.Param("n")
	if len(param) == 0 {
		utils.AbortWithError(c, errors.New("pas de paramètre 'n'"))
		return
	}
	nbDays, err := strconv.Atoi(param)
	if err != nil {
		utils.AbortWithError(c, errors.Wrap(err, "le paramètre 'n' n'est pas un entier"))
		return
	}
	since = time.Now().AddDate(0, 0, -nbDays)
	logs, err := selectLogs(api.db.ctx, api.db.pool, since, time.Now())
	if err != nil {
		utils.AbortWithError(c, errors.Wrap(err, "erreur lors de la recherche des logs en base"))
		return
	}
	data, err := transformLogsToCompressedData(logs)
	if err != nil {
		utils.AbortWithError(c, err)
		return
	}
	c.Header("Content-Disposition", "attachment; filename=statsSince"+param+"days.csv.gz")
	c.Data(http.StatusOK, "application/octet-stream", data)
}

func (l line) getFieldsAsStringArray() []string {
	return []string{
		l.date.Format("20060102150405"),
		l.path,
		l.method,
		l.username,
		strings.Join(l.roles, "-"),
	}
}
