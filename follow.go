package main

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// Follow type follow pour l'API
type Follow struct {
	Siret          *string   `json:"siret,omitempty"`
	Username       *string   `json:"username,omitempty"`
	Active         bool      `json:"active"`
	Since          time.Time `json:"since"`
	Comment        string    `json:"comment"`
	Etablissements *Score    `json:"etablissements,omitempty"`
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
	fmt.Println("coucou")
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

func (f *Follow) list(roles scope) ([]Follow, Jerror) {
	liste, err := findAllListes()
	if err != nil {
		return nil, errorToJSON(500, err)
	}

	sqlFollow := `select 
	f.comment,
	f.since,
	et.siret,
	et.siren,
	en.raison_sociale, 
	et.commune, 
	d.libelle, 
	d.code,
	s.score,
	s.diff,
	di.chiffre_affaire,
	di.arrete_bilan,
	di.variation_ca,
	di.resultat_expl,
	ef.effectif,
	n.libelle,
	n1.libelle,
	et.code_activite,
	coalesce(ep.last_procol, 'in_bonis') as last_procol,
	case when 'dgefp' = any($1) then coalesce(ap.ap, false) else null end as activite_partielle ,
	case when 'urssaf' = any($1) then 
		case when u.dette[0] > u.dette[1] or u.dette[1] > u.dette[2] then true else false end 
	else null end as hausseUrssaf,
	s.alert,
	r.roles && $1 as visible,
	et.departement = any($1) as in_zone,
	true as followed
	from etablissement_follow f
	inner join v_roles r on r.siren = f.siren
	inner join etablissement0 et on et.siret = f.siret
	inner join entreprise0 en on en.siren = f.siren
	inner join departements d on d.code = et.departement
	left join v_naf n on n.code_n5 = et.code_activite
	left join v_last_effectif ef on ef.siret = f.siret
	left join v_hausse_urssaf u on u.siret = f.siret
	left join v_apdemande ap on ap.siret = f.siret
	left join v_last_procol ep on ep.siret = f.siret
	left join v_diane_variation_ca di on di.siren = f.siren
	left join score s on s.siret = f.siret and s.libelle_liste = $2
	where f.username = $3 and f.active
	`

	rows, err := db.Query(context.Background(), sqlFollow, roles, liste[0].ID, f.Username)
	if err != nil {
		return nil, errorToJSON(500, err)
	}

	var follows []Follow
	for rows.Next() {
		var f Follow
		var e Score
		err := rows.Scan(
			&f.Comment,
			&f.Since,
			&e.Siret,
			&e.Siren,
			&e.RaisonSociale,
			&e.Commune,
			&e.LibelleDepartement,
			&e.Departement,
			&e.Score,
			&e.Diff,
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
			&e.Visible,
			&e.InZone,
			&e.Followed,
		)
		if err != nil {
			return nil, errorToJSON(500, err)
		}
		f.Etablissements = &e
		f.Active = true
		follows = append(follows, f)
	}

	if len(follows) == 0 {
		return nil, newJSONerror(204, "aucun suivi")
	}

	return follows, nil
}
