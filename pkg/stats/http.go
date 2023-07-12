package stats

import (
	"archive/zip"
	"context"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"datapi/pkg/utils"
)

const MAX = 9999999
const DATE_FORMAT = "2006-01-02"
const STATS_FILENAME = "stats_datapi"

type API struct {
	db                 StatsDB
	maxResultsToReturn int
	dateFormat         string
	filename           string
}

func NewAPI(db StatsDB) *API {
	return &API{db: db, maxResultsToReturn: MAX, dateFormat: DATE_FORMAT, filename: STATS_FILENAME}
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
	results := make(chan accessLog)
	go api.fetchLogs(since, to, results)
	// Create a zip archive.
	archive := zip.NewWriter(c.Writer)
	entry, err := archive.CreateHeader(&zip.FileHeader{
		Name:     api.filename + ".csv",
		Comment:  "fourni par Datapi avec amour",
		NonUTF8:  false,
		Modified: time.Now(),
	})
	c.Header("Content-type", "application/zip")
	c.Header("Content-Disposition", "attachment; filename="+api.filename+".zip")
	c.Stream(func(w io.Writer) bool {
		if err != nil {
			utils.AbortWithError(c, err)
		}
		err = writeLinesToCSV(results, api.maxResultsToReturn, entry)
		if err != nil {
			utils.AbortWithError(c, err)
		}
		return false
	})
	err = archive.Flush()
	if err != nil {
		utils.AbortWithError(c, err)
	}
	if err := archive.Close(); err != nil {
		utils.AbortWithError(c, err)
	}
}

func (api *API) fetchLogs(since time.Time, to time.Time, result chan accessLog) {
	selectLogs(api.db.ctx, api.db.pool, since, to, result)
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
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"erreur": err.Error()})
		return
	}
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
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"erreur": err.Error()})
		return
	}

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
