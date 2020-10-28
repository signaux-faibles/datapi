package main

import (
	"context"
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
	UnfollowComment      string    `json:"unfollowComment"`
	UnfollowCategory     string    `json:"unfollowCategory"`
	EtablissementSummary *Summary  `json:"etablissementSummary,omitempty"`
}

func getEtablissementsFollowedByCurrentUser(c *gin.Context) {
	username := c.GetString("username")
	scope := scopeFromContext(c)
	follow := Follow{Username: &username}
	follows, err := follow.list(scope)

	if err != nil {
		c.JSON(err.Code(), err.Error())
		return
	}

	c.JSON(200, follows)
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
		c.AbortWithError(500, paramErr)
		return
	}
	if param.Category == "" {
		c.JSON(400, "mandatory non-empty `category` property")
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
		c.AbortWithError(500, err)
		return
	}

	if follow.Active {
		c.JSON(204, follow)
	} else {
		err := follow.activate()
		if err != nil && err.Error() != "no rows in result set" {
			c.AbortWithError(500, err)
			return
		} else if err != nil && err.Error() == "no rows in result set" {
			c.JSON(403, "unknown establishment")
			return
		}
		c.JSON(201, follow)
	}
}

func unfollowEtablissement(c *gin.Context) {
	siret := c.Param("siret")
	username := c.GetString("username")

	var param struct {
		UnfollowComment  string `json:"unfollowComment"`
		UnfollowCategory string `json:"unfollowCategory"`
	}
	paramErr := c.ShouldBind(&param)
	if paramErr != nil {
		c.AbortWithError(500, paramErr)
		return
	}
	if param.UnfollowCategory == "" {
		c.JSON(400, "mandatory non-empty `unfollowCategory` property")
		return
	}

	follow := Follow{
		Siret:            &siret,
		Username:         &username,
		UnfollowComment:  param.UnfollowComment,
		UnfollowCategory: param.UnfollowCategory,
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

	return db.QueryRow(context.Background(), sqlFollow, f.Username, f.Siret).Scan(&f.Active, &f.Since, &f.Comment, &f.Category)
}

func (f *Follow) activate() error {
	f.Active = true
	sqlActivate := `insert into etablissement_follow
    (siret, siren, username, active, since, comment, category)
    select $1, $2, $3, true, current_timestamp, $4, $5
    from etablissement0 e
    where siret = $6
    returning since, true`

	return db.QueryRow(context.Background(),
		sqlActivate,
		*f.Siret,
		(*f.Siret)[0:9],
		f.Username,
		f.Comment,
		f.Category,
		f.Siret,
	).Scan(&f.Since, &f.Active)
}

func (f *Follow) deactivate() Jerror {
	sqlUnactivate := `update etablissement_follow set active = false, until = current_timestamp, unfollow_comment = $3, unfollow_category = $4
        where siret = $1 and username = $2 and active = true`

	commandTag, err := db.Exec(context.Background(),
		sqlUnactivate, f.Siret, f.Username, f.UnfollowComment, f.UnfollowCategory)

	if err != nil {
		return errorToJSON(500, err)
	}

	if commandTag.RowsAffected() == 0 {
		return newJSONerror(204, "this establishment is already not followed")
	}

	return nil
}

func (f *Follow) list(roles scope) ([]Follow, Jerror) {
	liste, err := findAllListes()
	if err != nil {
		return nil, errorToJSON(500, err)
	}

	sqlFollow := `select * from get_summary($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18) as follow;`

	rows, err := db.Query(context.Background(), sqlFollow,
		roles.zoneGeo(), nil, nil, liste[0].ID, nil, nil,
		true, true, f.Username, false, "follow", false, nil,
		nil, true, nil, nil, nil,
	)
	if err != nil {
		return nil, errorToJSON(500, err)
	}

	var follows []Follow
	var throwAway interface{}
	for rows.Next() {
		var f Follow
		var e Summary
		err := rows.Scan(
			&e.Siret,
			&e.Siren,
			&e.RaisonSociale,
			&e.Commune,
			&e.LibelleDepartement,
			&e.Departement,
			&e.Score,
			&e.Detail,
			&e.FirstAlert,
			&e.DernierCA,
			&e.ArreteBilan,
			&e.VariationCA,
			&e.DernierREXP,
			&e.DernierEffectif,
			&e.LibelleActivite,
			&e.LibelleActiviteN1,
			&e.CodeActivite,
			&e.EtatProcol,
			&e.ActivitePartielle,
			&e.HausseUrssaf,
			&e.Alert,
			&throwAway,
			&throwAway,
			&throwAway,
			&e.Visible,
			&e.InZone,
			&e.Followed,
			&e.FollowedEntreprise,
			&e.Siege,
			&e.Groupe,
			&e.TerrInd,
			&f.Comment,
			&f.Category,
			&f.Since,
			&e.PermUrssaf,
			&e.PermDGEFP,
			&e.PermScore,
			&e.PermBDF,
		)
		if err != nil {
			return nil, errorToJSON(500, err)
		}
		f.EtablissementSummary = &e
		f.Active = true
		follows = append(follows, f)
	}

	if len(follows) == 0 {
		return nil, newJSONerror(204, "aucun suivi")
	}

	return follows, nil
}
