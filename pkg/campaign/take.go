package campaign

import (
	"context"
	"datapi/pkg/core"
	"datapi/pkg/db"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func takeHandler(c *gin.Context) {
	var s core.Session
	s.Bind(c)

	campaignID, err := strconv.Atoi(c.Param("campaignID"))
	if err != nil {
		c.JSON(400, `/campaign/take/:campaignID/:campaignEtablissementID: le parametre campaignID doit être un entier`)
		return
	}
	campaignEtablissementID, err := strconv.Atoi(c.Param("campaignEtablissementID"))
	if err != nil {
		c.JSON(400, `/campaign/take/:campaignID/:campaignEtablissementID: le parametre campaignEtablissementID doit être un entier`)
		return
	}

	ids := IDs{
		CampaignID(campaignID),
		CampaignEtablissementID(campaignEtablissementID),
	}

	message, err := take(c, ids, s.Username)

	if errors.As(err, &PendingNotFoundError{}) {
		c.JSON(http.StatusUnprocessableEntity, "établissement indisponible pour cette action")
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, "erreur imprévue")
	} else {
		stream.Message <- message
		c.JSON(http.StatusOK, "ok")
	}
}

func take(ctx context.Context, ids IDs, username string) (Message, error) {
	var campaignEtablissementActionID *int
	var codeDepartement *string
	zone := zoneForUser(username)

	dbConn := db.Get()
	row := dbConn.QueryRow(ctx, sqlTakePendingEtablissement, ids.CampaignID, ids.CampaignEtablissementID, username, zone)
	err := row.Scan(&campaignEtablissementActionID, &codeDepartement)

	if campaignEtablissementActionID == nil || codeDepartement == nil {
		return Message{}, PendingNotFoundError{err: errors.New("aucun id retourné par l'insert")}
	} else if err != nil {
		return Message{}, err
	}

	return Message{
		CampaignID:              ids.CampaignID,
		CampaignEtablissementID: ids.CampaignEtablissementID,
		Zone:                    []string{*codeDepartement},
		Type:                    "pending",
		Username:                username,
	}, nil
}
