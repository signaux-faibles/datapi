package stats

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"datapi/pkg/utils"
)

const MAX = 999999
const DATE_FORMAT = "2006-01-02"

type API struct {
	db                 StatsDB
	maxResultsToReturn int
	dateFormat         string
}

func NewAPI(db StatsDB) *API {
	return &API{db: db, maxResultsToReturn: MAX, dateFormat: DATE_FORMAT}
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
	endpoint.GET("/since/:n/months", api.sinceMonthsHandler)
	endpoint.GET("/from/:start/for/:n/days", api.fromStartForDaysHandler)
	endpoint.GET("/from/:start/for/:n/months", api.fromStartForMonthsHandler)
	endpoint.GET("/from/:start/to/:end", api.fromStartToHandler)
}

func (api *API) sinceMonthsHandler(c *gin.Context) {
	var start time.Time
	var end time.Time
	var err error

	end = time.Now()
	nbMonths, err := utils.GetIntHTTPParameter(c, "n")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"erreur": err.Error()})
		return
	}
	start = end.AddDate(0, -nbMonths, 0)
	api.handleStatsInInterval(c, start, end)
}

func (api *API) sinceDaysHandler(c *gin.Context) {
	var start time.Time
	var end time.Time
	var err error

	end = time.Now()
	nbDays, err := utils.GetIntHTTPParameter(c, "n")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"erreur": err.Error()})
		return
	}
	start = end.AddDate(0, 0, -nbDays)
	api.handleStatsInInterval(c, start, end)
}

func (api *API) handleStatsInInterval(c *gin.Context, since time.Time, to time.Time) {
	logs, err := api.fetchLogs(since, to)
	if err != nil {
		utils.AbortWithError(c, errors.Wrap(err, "erreur lors de la recherche des logs en base"))
		return
	}
	data, err := transformLogsToCompressedData(logs)
	if err != nil {
		utils.AbortWithError(c, err)
		return
	}
	c.Header("Content-Disposition", "attachment; filename=statsSince"+since.String()+".csv.gz")
	c.Data(http.StatusOK, "application/octet-stream", data)
}

func (api *API) fetchLogs(since time.Time, to time.Time) ([]line, error) {
	logs, err := selectLogs(api.db.ctx, api.db.pool, since, to)
	if err != nil {
		return nil, errors.Wrap(err, "erreur pendant la récupération des logs")
	}
	nbLogs := len(logs)
	if nbLogs == 0 {
		return nil, utils.NewJSONerror(http.StatusBadRequest, "aucune logs trouvée, les critères sont trop restrictifs")
	}
	if nbLogs > api.maxResultsToReturn {
		return nil, utils.NewJSONerror(http.StatusBadRequest, "trop de logs trouvées, les critères ne sont pas assez restrictifs")
	}
	return logs, nil
}

func (api *API) fromStartForDaysHandler(c *gin.Context) {
	var from time.Time
	var to time.Time
	var err error
	from, err = utils.GetDateHTTPParameter(c, "start", api.dateFormat)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"erreur": err.Error()})
		return
	}

	nbDays, err := utils.GetIntHTTPParameter(c, "n")
	to = from.AddDate(0, 0, nbDays)

	api.handleStatsInInterval(c, from, to)
}

func (api *API) fromStartForMonthsHandler(c *gin.Context) {
	var start time.Time
	var end time.Time
	var err error
	start, err = utils.GetDateHTTPParameter(c, "start", api.dateFormat)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"erreur": err.Error()})
		return
	}

	nbDays, err := utils.GetIntHTTPParameter(c, "n")
	end = start.AddDate(0, 0, nbDays)

	api.handleStatsInInterval(c, start, end)
}

func (api *API) fromStartToHandler(c *gin.Context) {
	var start time.Time
	var end time.Time
	var err error
	start, err = utils.GetDateHTTPParameter(c, "start", api.dateFormat)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"erreur": err.Error()})
		return
	}
	end, err = utils.GetDateHTTPParameter(c, "end", api.dateFormat)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"erreur": err.Error()})
		return
	}

	api.handleStatsInInterval(c, start, end)
}
