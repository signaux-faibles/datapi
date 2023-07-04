package stats

import (
	"context"
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

func (sh *StatsHandler) sinceDaysHandler(c *gin.Context) {
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
	logs, err := selectLogs(sh.ctx, sh.dbPool, since, time.Now())
	if err != nil {
		utils.AbortWithError(c, errors.Wrap(err, "erreur lors de la recherche des logs en base"))
		return
	}
	data, err := transformLogsToData(logs)
	if err != nil {
		utils.AbortWithError(c, err)
		return
	}
	c.Data(http.StatusOK, "text/csv", data)
}

func (l line) getFieldsAsStringArray() []string {
	return []string{
		l.date.Format("20060102150405"),
		l.method,
		l.username,
		strings.Join(l.roles, "-"),
	}
}
