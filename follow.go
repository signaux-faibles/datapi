package main

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

// Follow type follow pour l'API
type Follow struct {
	Siret                *string               `json:"siret"`
	Username             *string               `json:"username"`
	Active               bool                  `json:"active"`
	Since                time.Time             `json:"since"`
	Comment              string                `json:"comment"`
	EtablissementSummary *EtablissementSummary `json:"etablissementSummary,omitempty"`
}

func (f *Follow) load() error {
	sqlFollow := `select 
		active, since, comment		
		from etablissement_follow 
		where
		username = $1 and
		siret = $2 and
		active`

	return db.QueryRow(context.Background(), sqlFollow, f.Username, f.Siret).Scan(&f.Active, &f.Since, &f.Comment)
}

func (f *Follow) activate() error {
	f.Active = true
	sqlActivate := `insert into etablissement_follow
	(siret, siren, username, active, since, comment)
	select $1, $2, $3, true, current_timestamp, $4
	from etablissement0 e
	where siret = $5
	returning since, true`

	return db.QueryRow(context.Background(),
		sqlActivate,
		*f.Siret,
		(*f.Siret)[0:9],
		f.Username,
		f.Comment,
		f.Siret,
	).Scan(&f.Since, &f.Active)
}

func (f *Follow) deactivate() Jerror {
	sqlUnactivate := `update etablissement_follow set active = false, until = current_timestamp
		where siret = $1 and username = $2 and active = true`

	commandTag, err := db.Exec(context.Background(),
		sqlUnactivate, f.Siret, f.Username)

	if err != nil {
		return errorToJSON(500, err)
	}

	if commandTag.RowsAffected() == 0 {
		return newJSONerror(204, "this establishment is already not followed")
	}

	return nil
}

func (f *Follow) list() ([]Follow, Jerror) {
	sqlFollow := `select f.siret, f.username, f.since, f.comment,
	d.libelle, d.code, en.raison_sociale, e.commune,
	r.libelle, e.ape, n.code_n1, n.libelle_n5, n.libelle_n1, ef.effectif
	from etablissement_follow f
	inner join etablissement0 e on e.siret = f.siret
	inner join departements d on d.code = e.departement
	inner join regions r on r.id = d.id_region
	inner join entreprise0 en on en.siren = f.siren
	inner join v_naf n on e.ape = n.code_n5
	left join v_last_effectif ef on ef.siret = f.siret
	where ((f.siret = $1 or $1 is null) and (f.username = $2 or $2 is null)) and f.active`

	rows, err := db.Query(context.Background(), sqlFollow, f.Siret, f.Username)
	if err != nil {
		return nil, errorToJSON(500, err)
	}

	var follows []Follow
	for rows.Next() {
		var f Follow
		var e EtablissementSummary
		err := rows.Scan(
			&f.Siret, &f.Username, &f.Since, &f.Comment,
			&e.LibelleDepartement, &e.Departement, &e.RaisonSociale, &e.Commune,
			&e.Region, &e.CodeActivite, &e.CodeSecteur, &e.Activite, &e.Secteur, &e.DernierEffectif,
		)
		if err != nil {
			return nil, errorToJSON(500, err)
		}
		f.EtablissementSummary = &e
		follows = append(follows, f)
	}

	return follows, nil
}

func followEtablissement(c *gin.Context) {
	siret := c.Param("siret")
	username := c.GetString("username")
	var param struct {
		Comment string `json:"comment"`
	}
	err := c.ShouldBind(&param)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	if param.Comment == "" {
		c.JSON(400, "mandatory non-empty `comment` property")
		return
	}

	follow := Follow{
		Siret:    &siret,
		Username: &username,
		Comment:  param.Comment,
	}

	err = follow.load()
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
	follow := Follow{
		Siret:    &siret,
		Username: &username,
	}

	err := follow.deactivate()
	if err != nil {
		c.JSON(err.Code(), err.Error())
		return
	}
	c.JSON(200, "this establishment is no longer followed")
}
