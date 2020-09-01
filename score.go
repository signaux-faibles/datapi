package main

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

type paramsListeScores struct {
	Departements []string `json:"zone,omitempty"`
	EtatsProcol  []string `json:"procol,omitempty"`
	Activites    []string `json:"activite,omitempty"`
	EffectifMin  *int     `json:"effectifMin"`
	EffectifMax  *int     `json:"effectifMax"`
	VueFrance    *bool    `json:"vueFrance"`
}

// Liste de détection
type Liste struct {
	ID     string            `json:"id"`
	Batch  string            `json:"batch"`
	Algo   string            `json:"algo"`
	Query  paramsListeScores `json:"-"`
	Scores []Score           `json:"scores,omitempty"`
}

// Score d'une liste de détection
type Score struct {
	Siren              string     `json:"siren"`
	Siret              string     `json:"siret"`
	Score              float64    `json:"-"`
	Diff               *float64   `json:"-"`
	RaisonSociale      *string    `json:"raison_sociale"`
	Commune            *string    `json:"commune"`
	LibelleActivite    *string    `json:"libelle_activite"`
	LibelleActiviteN1  *string    `json:"libelle_activite_n1"`
	CodeActivite       *string    `json:"code_activite"`
	Departement        *string    `json:"departement"`
	LibelleDepartement *string    `json:"libelleDepartement"`
	DernierEffectif    *int       `json:"dernier_effectif"`
	HausseUrssaf       *bool      `json:"urssaf"`
	ActivitePartielle  *bool      `json:"activite_partielle"`
	DernierCA          *int       `json:"ca"`
	VariationCA        *float64   `json:"variation_ca"`
	ArreteBilan        *time.Time `json:"arrete_bilan"`
	DernierREXP        *int       `json:"resultat_expl"`
	EtatProcol         *string    `json:"etat_procol"`
	Alert              *string    `json:"alert"`
}

func getEtablissementsFollowedByCurrentUser(c *gin.Context) {
	username := c.GetString("username")
	follow := Follow{Username: &username}
	follows, err := follow.list()

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

func getListes(c *gin.Context) {
	listes, err := findAllListes()
	if err != nil {
		c.JSON(500, err.Error())
	}
	c.JSON(200, listes)
}

func getLastListeScores(c *gin.Context) {
	roles := scopeFromContext(c)
	listes, err := findAllListes()
	if err != nil || len(listes) == 0 {
		c.AbortWithStatus(204)
		return
	}

	var params paramsListeScores
	err = c.Bind(&params)

	if err != nil || len(listes) == 0 {
		c.AbortWithStatus(400)
		return
	}

	liste := Liste{
		ID:    listes[0].ID,
		Query: params,
	}

	err = liste.getScores(roles)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}
	c.JSON(200, liste)
}

func getListeScores(c *gin.Context) {
	roles := scopeFromContext(c)

	var params paramsListeScores
	err := c.Bind(&params)

	liste := Liste{
		ID:    c.Param("id"),
		Query: params,
	}

	if err != nil || liste.ID == "" {
		c.AbortWithStatus(400)
		return
	}

	err = liste.getScores(roles)

	if err != nil {
		if err.Error() == "no rows in result set" {
			c.AbortWithStatus(204)
			return
		}
		c.JSON(500, err.Error())
		return
	}
	c.JSON(200, liste)
}

func (liste *Liste) getScores(roles scope) error {
	if liste.Batch == "" {
		err := liste.load()
		if err != nil {
			return err
		}
	}
	sqlScores := `	select 
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
		et.ape,
		coalesce(ep.last_procol, 'in_bonis') as last_procol,
		coalesce(ap.ap, false) as activite_partielle,
		case when u.dette[0] > u.dette[1] or u.dette[1] > u.dette[2] then true else false end as hausseUrssaf,
		s.alert
	from score0 s
	inner join v_roles r on r.roles && $1 and r.siren = s.siren
	inner join etablissement0 et on et.siret = s.siret
	inner join entreprise0 en on en.siren = s.siren
	inner join departements d on d.code = et.departement
	inner join naf n on n.code = et.ape
	inner join naf n1 on n.id_n1 = n1.id
	left join v_last_effectif ef on ef.siret = s.siret
	left join v_hausse_urssaf u on u.siret = s.siret
	left join v_apdemande ap on ap.siret = s.siret
	left join v_last_procol ep on ep.siret = s.siret
	left join v_diane_variation_ca di on di.siren = s.siren
	where s.libelle_liste = $2
	and (coalesce(ep.last_procol, 'in_bonis')=any($3) or $3 is null)
	and (et.departement=any($4) or $4 is null)
	and (n1.code=any($5) or $5 is null)
	and (ef.effectif >= $6 or $6 is null)
	and (ef.effectif <= $7 or $7 is null)
	and (et.departement=any($1) or $8 = true)
	and s.alert != 'Pas d''alerte'
	order by s.score desc;`
	rows, err := db.Query(context.Background(), sqlScores, roles.zoneGeo(), liste.ID, liste.Query.EtatsProcol,
		liste.Query.Departements, liste.Query.Activites, liste.Query.EffectifMin, liste.Query.EffectifMax, liste.Query.VueFrance)
	if err != nil {
		return err
	}

	var scores []Score
	for rows.Next() {
		var score Score
		err := rows.Scan(
			&score.Siret,
			&score.Siren,
			&score.RaisonSociale,
			&score.Commune,
			&score.LibelleDepartement,
			&score.Departement,
			&score.Score,
			&score.Diff,
			&score.DernierCA,
			&score.ArreteBilan,
			&score.VariationCA,
			&score.DernierREXP,
			&score.DernierEffectif,
			&score.LibelleActivite,
			&score.LibelleActiviteN1,
			&score.CodeActivite,
			&score.EtatProcol,
			&score.ActivitePartielle,
			&score.HausseUrssaf,
			&score.Alert,
		)
		if err != nil {
			return err
		}
		scores = append(scores, score)
	}

	liste.Scores = scores
	return nil
}

func findAllListes() ([]Liste, error) {
	var listes []Liste
	rows, err := db.Query(context.Background(), "select algo, batch, libelle from liste order by batch desc, algo asc")
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var l Liste
		err := rows.Scan(&l.Algo, &l.Batch, &l.ID)
		if err != nil {
			return nil, err
		}
		listes = append(listes, l)
	}

	return listes, nil
}

func (liste *Liste) load() error {
	sqlListe := `select batch, algo from liste where libelle=$1 and version=0`
	row := db.QueryRow(context.Background(), sqlListe, liste.ID)
	batch, algo := "", ""
	err := row.Scan(&batch, &algo)
	if err != nil {
		return err
	}
	liste.Batch = batch
	liste.Algo = algo
	return nil
}
