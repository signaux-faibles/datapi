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
	logger := slog.Default().With(slog.Any("since", since), slog.Any("to", to))
	//activitesUtilisateur := make(chan row[activiteParUtilisateur])
	//activitesJour := make(chan row[activiteParJour])
	accessLogs := fetchAny(api, accessLog{}, newAccessLogsSelector(since, to))
	activitesUtilisateur := fetchAny(api, activiteParUtilisateur{}, newActiviteParUtilisateurSelector(since, to))
	activitesJour := fetchAny(api, activiteParJour{}, newActiviteParJourSelector(since, to))
	//accessLogs := api.fetchLogs(since, to)
	//go api.fetchActivitesParUtilisateur(since, to, activitesUtilisateur)
	//go api.fetchActivitesParJour(since, to, activitesJour)
	// Create a zip archive.
	c.Header("Content-type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename="+api.filename+".zip")
	c.Stream(func(w io.Writer) bool {
		archive := zip.NewWriter(w)
		filename := api.filename + ".xlsx"
		logger = slog.Default().With("filename", filename)
		xlsInZip, err := createEntryInZip(archive, filename)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return false
		}
		// crée le fichier excel
		xls := newExcel()
		defer func() {
			if err := xls.Close(); err != nil {
				logger.Error("erreur à la fermeture du fichier", slog.Any("error", err))
			}
		}()
		// on écrit les données
		err = writeOneSheetToExcel(xls, accessLogSheetConfig(), accessLogs)
		if err != nil {
			logger.Error("erreur pendant l'écriture des accessLogs", slog.Any("error", err))
			c.Status(http.StatusInternalServerError)
			return false
		}
		err = writeOneSheetToExcel(xls, activitesUtilisateurSheetConfig(), activitesUtilisateur)
		if err != nil {
			logger.Error("erreur pendant l'écriture des activités utilisateurs", slog.Any("error", err))
			c.Status(http.StatusInternalServerError)
			return false
		}
		err = writeOneSheetToExcel(xls, activitesJourSheetConfig(), activitesJour)
		if err != nil {
			logger.Error("erreur pendant l'écriture des activités jour", slog.Any("error", err))
			c.Status(http.StatusInternalServerError)
			return false
		}
		// on termine le boulot proprement
		err = exportTo(xlsInZip, xls)
		if err != nil {
			logger.Error("erreur pendant l'écriture du excel dans le zip", slog.Any("error", err))
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

func (api *API) fetchLogs(since time.Time, to time.Time) chan row[accessLog] {
	return fetchAny(api, accessLog{}, newAccessLogsSelector(since.Truncate(day), to.Truncate(day)))
}

func (api *API) fetchActivitesParUtilisateur(since time.Time, to time.Time) chan row[activiteParUtilisateur] {
	return fetchAny(api, activiteParUtilisateur{}, newActiviteParUtilisateurSelector(since.Truncate(day), to.Truncate(day)))
}

func (api *API) fetchActivitesParJour(since time.Time, to time.Time) chan row[activiteParJour] {
	return fetchAny(api, activiteParJour{}, newActiviteParJourSelector(since.Truncate(day), to.Truncate(day)))
	//selectActiviteParJour(api.db.ctx, api.db.pool, since, to, result)
}

func createEntryInZip(archive *zip.Writer, filename string) (io.Writer, error) {
	xlsEntry, err := archive.CreateHeader(&zip.FileHeader{
		Name:     filename,
		Comment:  "fourni par Datapi avec dégoût",
		NonUTF8:  false,
		Modified: time.Now(),
	})
	if err != nil {
		slog.Error("erreur pendant la création du xls dans le zip", slog.Any("error", err))
		return xlsEntry, err
	}
	return xlsEntry, nil
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
