package stats

import (
	"bytes"
	"context"
	"encoding/csv"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"

	"datapi/pkg/utils"
)

type StatsHandler struct {
	ctx    context.Context
	dbPool *pgxpool.Pool
}

type line struct {
	date     time.Time
	method   string
	username string
	roles    []string
}

// statusHandler : point d'entrée de l'API qui retourne les infos d'un `Refresh` depuis son `UUID`
func (sh *StatsHandler) sinceMonthsHandler(c *gin.Context) {
	var since time.Time
	param := c.Param("n")
	if len(param) == 0 {
		utils.AbortWithError(c, errors.New("pas de paramètre 'n'"))
		return
	}
	index, err := strconv.Atoi(param)
	if err != nil {
		utils.AbortWithError(c, errors.Wrap(err, "le parametre 'n' n'est pas un entier"))
		return
	}
	since = time.Now().Add(time.Duration(index*-24) * time.Hour)
	logs, err := selectLogs(sh.ctx, sh.dbPool, since)
	if err != nil {
		utils.AbortWithError(c, errors.Wrap(err, "erreur lors de la recherche des logs en base"))
		return
	}
	data := make([]byte, 0)
	w := bytes.NewBuffer(data)
	csvW := csv.NewWriter(w)
	utils.Apply(logs, func(l line) {
		wErr := csvW.Write(l.getFieldsAsStringArray())
		if wErr != nil {
			utils.AbortWithError(c, errors.Wrap(wErr, "erreur lors de la transformation de la ligne de log "+l.String()))
			return
		}
	})
	c.Data(http.StatusOK, "text/csv", data)
}

func selectLogs(ctx context.Context, dbPool *pgxpool.Pool, since time.Time) ([]line, error) {
	sql := `select date_add,
       path,
       method,
--        body,
       tokencontent ->> 'preferred_username'::text as username,
       translate(((tokencontent ->> 'resource_access')::json ->> 'signauxfaibles')::json ->> 'roles', '[]', '{}')::text[]  as roles
    from v_log
    where date_add > $1`
	stats := []line{}
	rows, err := dbPool.Query(ctx, sql, since)
	if err != nil {
		return stats, err
	}
	err = rows.Scan(stats)
	if err != nil {
		return stats, err
	}
	return stats, nil
}

func (l line) getFieldsAsStringArray() []string {
	return []string{
		l.date.Format("20060102150405"),
		l.method,
		l.username,
		strings.Join(l.roles, "-"),
	}
}

func (l line) String() string {
	return strings.Join(l.getFieldsAsStringArray(), ";")
}
