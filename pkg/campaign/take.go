package campaign

import (
	"context"
	"datapi/pkg/core"
	"datapi/pkg/db"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/signaux-faibles/libwekan"
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

	zone := zoneForUser(libwekan.Username(s.Username))

	message, err := take(
		c,
		CampaignID(campaignID),
		CampaignEtablissementID(campaignEtablissementID),
		libwekan.Username(s.Username),
		zone,
	)

	if errors.As(err, &PendingNotFoundError{}) {
		c.JSON(http.StatusUnprocessableEntity, "établissement indisponible pour cette action")
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, "erreur imprévue")
	} else {
		stream.Message <- message
		c.JSON(http.StatusOK, "ok")
	}
}

func take(ctx context.Context, campaignID CampaignID, campaignEtablissementID CampaignEtablissementID, username libwekan.Username, zone Zone) (Message, error) {
	var campaignEtablissementActionID *int
	var codeDepartement *string
	dbConn := db.Get()
	row := dbConn.QueryRow(ctx, sqlTakePendingEtablissement, campaignID, campaignEtablissementID, username, zone)
	err := row.Scan(&campaignEtablissementActionID, &codeDepartement)
	if campaignEtablissementActionID == nil || codeDepartement == nil {
		return Message{}, PendingNotFoundError{err: errors.New("aucun id retourné par l'insert")}
	} else if err != nil {
		return Message{}, err
	}
	return Message{
		CampaignID:              campaignID,
		CampaignEtablissementID: campaignEtablissementID,
		Zone:                    []string{*codeDepartement},
		Type:                    "pending",
		Username:                string(username),
	}, nil
}
