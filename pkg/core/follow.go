package core

import (
	"context"
	"datapi/pkg/db"
	"datapi/pkg/utils"
	"github.com/signaux-faibles/libwekan"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Follow type follow pour l'API
type Follow struct {
	Siret                *string   `json:"siret,omitempty"`
	Username             *string   `json:"username,omitempty"`
	Active               bool      `json:"active"`
	Since                time.Time `json:"since"`
	Comment              string    `json:"comment"`
	Category             string    `json:"category"`
	UnfollowComment      string    `json:"unfollowComment,omitempty"`
	UnfollowCategory     string    `json:"unfollowCategory,omitempty"`
	EtablissementSummary *Summary  `json:"etablissementSummary,omitempty"`
}

func followEtablissement(c *gin.Context) {
	siret := c.Param("siret")
	username := c.GetString("username")

	var param struct {
		Comment  string `json:"comment"`
		Category string `json:"category"`
	}
	paramErr := c.ShouldBind(&param)
	if paramErr != nil {
		utils.AbortWithError(c, paramErr)
		return
	}
	if param.Category == "" {
		c.AbortWithStatusJSON(
			http.StatusBadRequest,
			gin.H{"message": "la propriété `category` est obligatoire"},
		)
		return
	}

	follow := Follow{
		Siret:    &siret,
		Username: &username,
		Comment:  param.Comment,
		Category: param.Category,
	}

	err := follow.load()
	if err != nil && err.Error() != "no rows in result set" {
		utils.AbortWithError(c, err)
		return
	}
	if follow.Active {
		c.JSON(http.StatusNoContent, follow)
		return
	}

	err = follow.activate()
	if err != nil {
		if err.Error() == "no rows in result set" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "établissement inconnu"})
			return
		}
		utils.AbortWithError(c, err)
		return
	}
	c.JSON(http.StatusCreated, follow)
	return
}

func unfollowEtablissement(c *gin.Context) {
	var s session
	s.Bind(c)
	siret := c.Param("siret")

	var param struct {
		UnfollowComment  string `json:"unfollowComment"`
		UnfollowCategory string `json:"unfollowCategory"`
	}
	paramErr := c.ShouldBind(&param)
	if paramErr != nil {
		utils.AbortWithError(c, paramErr)
		return
	}
	if param.UnfollowCategory == "" {
		c.JSON(400, "mandatory non-empty `unfollowCategory` property")
		return
	}

	follow := Follow{
		Siret:            &siret,
		Username:         &s.Username,
		UnfollowComment:  param.UnfollowComment,
		UnfollowCategory: param.UnfollowCategory,
	}

	if s.hasRole("wekan") {
		user, ok := kanban.GetUser(libwekan.Username(s.Username))
		if !ok {
			c.JSON(500, "l'utilisateur a le rôle wekan mais n'est pas présent dans l'application")
		}
		cards, err := kanban.SelectCardsFromSiret(c, siret, libwekan.Username(s.Username))
		if err != nil {
			c.JSON(500, err.Error())
		}
		for _, card := range cards {
			err := kanban.PartCard(c, card.ID, user.ID)
			if err != nil {
				c.JSON(500, err.Error())
				return
			}
		}
	}

	err := follow.deactivate()
	if err != nil {
		c.JSON(err.Code(), err.Error())
		return
	}
	c.JSON(200, "this establishment is no longer followed")
}

func (f *Follow) load() error {
	sqlFollow := `select
        active, since, comment, category
        from etablissement_follow
        where
        username = $1 and
        siret = $2 and
        active`

	row := db.Get().QueryRow(
		context.Background(),
		sqlFollow,
		f.Username,
		f.Siret,
	)
	return row.Scan(
		&f.Active,
		&f.Since,
		&f.Comment,
		&f.Category,
	)
}

func (f *Follow) activate() error {
	f.Active = true
	sqlActivate := `insert into etablissement_follow
    (siret, siren, username, active, since, comment, category)
    select $1, $2, $3, true, current_timestamp, $4, $5
    from etablissement0 e
    where siret = $6
    returning since, true`

	return db.Get().QueryRow(context.Background(),
		sqlActivate,
		*f.Siret,
		(*f.Siret)[0:9],
		f.Username,
		f.Comment,
		f.Category,
		f.Siret,
	).Scan(&f.Since, &f.Active)
}

func (f *Follow) deactivate() utils.Jerror {
	sqlUnactivate := `update etablissement_follow set active = false, until = current_timestamp, unfollow_comment = $3, unfollow_category = $4
        where siret = $1 and username = $2 and active = true`

	commandTag, err := db.Get().Exec(context.Background(),
		sqlUnactivate, f.Siret, f.Username, f.UnfollowComment, f.UnfollowCategory)

	if err != nil {
		return utils.ErrorToJSON(500, err)
	}

	if commandTag.RowsAffected() == 0 {
		return utils.NewJSONerror(204, "this establishment is already not followed")
	}

	return nil
}

func (f *Follow) list(roles Scope) (Follows, utils.Jerror) {
	liste, err := findAllListes()
	if err != nil {
		return nil, utils.ErrorToJSON(500, err)
	}

	params := summaryParams{roles, nil, nil, &liste[0].ID, false, nil,
		&True, &True, *f.Username, false, "follow", &False, nil,
		nil, &True, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil}

	sms, err := getSummaries(params)
	if err != nil {
		return nil, utils.ErrorToJSON(500, err)
	}
	var follows Follows

	for _, s := range sms.Summaries {
		var f Follow
		f.Comment = *utils.Coalesce(s.Comment, &EmptyString)
		f.Category = *utils.Coalesce(s.Category, &EmptyString)
		f.Since = *utils.Coalesce(s.Since, &time.Time{})
		f.EtablissementSummary = s
		f.Active = true
		follows = append(follows, f)
	}

	if len(follows) == 0 {
		return nil, utils.NewJSONerror(204, "aucun suivi")
	}

	return follows, nil
}

// Follows is a slice of Follows
type Follows []Follow
