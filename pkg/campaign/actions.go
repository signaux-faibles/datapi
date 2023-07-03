package campaign

import (
	"context"
	"datapi/pkg/core"
	"datapi/pkg/db"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/signaux-faibles/libwekan"
	"net/http"
	"strconv"
)

type CampaignEtablissementActionID int

func (c CampaignEtablissementActionID) Tuple() []interface{} {
	return []interface{}{&c}
}

func doAction(ctx context.Context,
	campaignID CampaignID,
	campaignEtablissementID CampaignEtablissementID,
	username libwekan.Username,
	zone Zone,
	action string,
	detail string,
) (Message, error) {
	var campaignEtablissementActionID *int
	var codeDepartement *string
	dbConn := db.Get()
	row := dbConn.QueryRow(ctx, sqlActionMyEtablissement, campaignID,
		campaignEtablissementID, username, zone, action, detail)
	err := row.Scan(&campaignEtablissementActionID, &codeDepartement)
	fmt.Println(err)
	if campaignEtablissementActionID == nil || codeDepartement == nil {
		return Message{}, TakeNotFoundError{err: errors.New("aucun id retourné par l'insert")}
	} else if err != nil {
		return Message{}, err
	}
	message := Message{
		campaignID,
		campaignEtablissementID,
		[]string{*codeDepartement},
		action,
		string(username),
	}
	return message, nil
}

type actionParam struct {
	Detail string `json:"detail"`
}

func actionHandlerFunc(action string) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		var s core.Session
		s.Bind(ctx)

		var params actionParam
		err := ctx.Bind(&params)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, "décodage de la requete impossible: "+err.Error())
		}
		if params.Detail == "" {
			ctx.JSON(http.StatusBadRequest, "le paramètre detail doit contenir au moins 1 caractère")
		}

		campaignID, err := strconv.Atoi(ctx.Param("campaignID"))
		if err != nil {
			ctx.JSON(400, `/campaign/take/:campaignID/:campaignEtablissementID: le parametre campaignID doit être un entier`)
			return
		}

		campaignEtablissementID, err := strconv.Atoi(ctx.Param("campaignEtablissementID"))
		if err != nil {
			ctx.JSON(400, `/campaign/take/:campaignID/:campaignEtablissementID: le parametre campaignEtablissementID doit être un entier`)
			return
		}

		zone := zoneForUser(libwekan.Username(s.Username))

		if err != nil {
			ctx.JSON(http.StatusBadRequest, "/"+action+"/:campaignID, :campaignID n'est pas entier")
		}
		message, err := doAction(ctx,
			CampaignID(campaignID),
			CampaignEtablissementID(campaignEtablissementID),
			libwekan.Username(s.Username),
			zone,
			action,
			params.Detail)

		if errors.As(err, &TakeNotFoundError{}) {
			ctx.JSON(http.StatusUnprocessableEntity, "traitement indisponible pour cet établissement")
			return
		} else if err != nil {
			ctx.JSON(http.StatusInternalServerError, "erreur innattendue")
			return
		}

		stream.Message <- message
		ctx.JSON(http.StatusOK, "ok")
	}
}
