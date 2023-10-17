package stats

import (
	"archive/zip"
	"context"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"datapi/pkg/utils"
)

const STATS_FILENAME = "stats_datapi"

type API struct {
	db         StatsDB
	dateFormat string
	filename   string
}

func NewAPI(db StatsDB) *API {
	return &API{db: db, dateFormat: time.DateOnly, filename: STATS_FILENAME}
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
	results := make(chan row[accessLog])
	activitesUtilisateur := make(chan row[activiteParUtilisateur])
	activitesJour := make(chan row[activiteParJour])
	go api.fetchLogs(since, to, results)
	go api.fetchActivitesParUtilisateur(since, to, activitesUtilisateur)
	go api.fetchActivitesParJour(since, to, activitesJour)
	// Create a zip archive.
	c.Header("Content-type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename="+api.filename+".zip")
	c.Stream(func(w io.Writer) bool {
		archive := zip.NewWriter(w)
		err := createCSV(archive, api.filename, results)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return false
		}
		err = createExcel(archive, api.filename, activitesUtilisateur, activitesJour)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return false
		}
		err = archive.Close()
		if err != nil {
			slog.Error("erreur pendant la fermeture de l'archive", slog.Any("error", err))
			c.Status(http.StatusInternalServerError)
		}
		return false
	})
}

func (api *API) fetchLogs(since time.Time, to time.Time, result chan row[accessLog]) {
	selectLogs(api.db.ctx, api.db.pool, since, to, result)
}

func (api *API) fetchActivitesParUtilisateur(since time.Time, to time.Time, activites chan row[activiteParUtilisateur]) {
	selectActiviteParUtilisateur(api.db.ctx, api.db.pool, since, to, activites)
}

func (api *API) fetchActivitesParJour(since time.Time, to time.Time, result chan row[activiteParJour]) {
	selectActiviteParJour(api.db.ctx, api.db.pool, since, to, result)
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
	// end est inclus dans les résultats, donc on va jusqu'au lendemain exclus
	api.handleStatsInInterval(c, start, end.AddDate(0, 0, 1))
}
