package main

import (
	"context"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type searchParams struct {
	search      string
	page        int
	ignoreRoles bool
	ignoreZone  bool
	roles       scope
	username    string
}

type searchResult struct {
	From    int       `json:"from"`
	To      int       `json:"to"`
	Total   int       `json:"total"`
	PageMax int       `json:"pageMax"`
	Page    int       `json:"page"`
	Results []Summary `json:"results"`
}

func searchEtablissementHandler(c *gin.Context) {
	var params searchParams
	var err error

	params.username = c.GetString("username")

	if search := c.Param("search"); len(search) >= 3 {
		params.search = search
	} else {
		c.JSON(400, "search string length < 3")
	}

	if page, ok := c.GetQuery("page"); ok {
		params.page, err = strconv.Atoi(page)
		if err != nil {
			c.JSON(400, "page has to be integer >= 0")
			return
		}
	}

	if ignoreRoles, ok := c.GetQuery("ignoreroles"); ok {
		params.ignoreRoles, err = strconv.ParseBool(ignoreRoles)
		if err != nil {
			c.JSON(400, "france is either `true` or `false`")
			return
		}
	}

	if ignoreZone, ok := c.GetQuery("ignorezone"); ok {
		params.ignoreZone, err = strconv.ParseBool(ignoreZone)
		if err != nil {
			c.JSON(400, "roles is either `true` or `false`")
			return
		}
	}

	params.roles = scopeFromContext(c)

	result, Jerr := searchEtablissement(params)
	if Jerr != nil {
		c.JSON(Jerr.Code(), Jerr.Error())
		return
	}
	c.JSON(200, result)

}

func searchEtablissement(params searchParams) (searchResult, Jerror) {
	sqlSearch := `select 
		et.siret,
		et.siren,
		en.raison_sociale, 
		et.commune, 
		d.libelle, 
		d.code,
		case when (r.roles && $1 and vs.siret is not null) or f.id is not null then s.score else null end,
		case when (r.roles && $1 and vs.siret is not null) or f.id is not null then s.diff else null end,
		di.chiffre_affaire,
		di.arrete_bilan,
		di.variation_ca,
		di.resultat_expl,
		ef.effectif,
		n.libelle_n5,
		n.libelle_n1,
		et.code_activite,
		coalesce(ep.last_procol, 'in_bonis') as last_procol,
		case when 'dgefp' = any($1) and ((r.roles && $1 and vs.siret is not null) or f.id is not null) then coalesce(ap.ap, false) else null end as activite_partielle ,
		case when 'urssaf' = any($1) and ((r.roles && $1 and vs.siret is not null) or f.id is not null) then 
			case when u.dette[0] > u.dette[1] or u.dette[1] > u.dette[2] then true else false end 
		else null end as hausseUrssaf,
		case when 'detection' = any($1) and ((r.roles && $1 and vs.siret is not null) or f.id is not null) then s.alert else null end,
		count(*) over (),
		r.roles && $1 as visible,
		coalesce(et.departement = any($2), false) as in_zone,
		f.id is not null as followed,
		et.siege,
		g.siren is not null,
		ti.code_commune is not null
		from etablissement0 et
		inner join v_roles r on et.siren = r.siren
		inner join entreprise0 en on en.siren = r.siren
		inner join departements d on d.code = et.departement
		left join v_naf n on n.code_n5 = et.code_activite
		left join score s on et.siret = s.siret
		left join v_alert_etablissement vs on vs.siret = et.siret
		left join v_last_effectif ef on ef.siret = et.siret
		left join v_hausse_urssaf u on u.siret = et.siret
		left join v_apdemande ap on ap.siret = et.siret
		left join v_last_procol ep on ep.siret = et.siret
		left join v_diane_variation_ca di on di.siren = s.siren
		left join etablissement_follow f on f.siret = et.siret and f.active and f.username = $10
		left join entreprise_ellisphere0 g on g.siren = et.siren
		left join terrind ti on ti.code_commune = et.code_commune
		where (et.siret ilike $6 or en.raison_sociale ilike $7)
		and (r.roles && $1 or $8)
		and (et.departement=any($2) or $9)
		and coalesce(s.libelle_liste, $5) = $5
		order by en.raison_sociale, siret
		limit $3 offset $4
		;`

	liste, err := findAllListes()
	if err != nil {
		return searchResult{}, errorToJSON(500, err)
	}
	zoneGeo := params.roles.zoneGeo()
	limit := viper.GetInt("searchPageLength")

	rows, err := db.Query(context.Background(), sqlSearch,

		zoneGeo,
		zoneGeo,
		limit,
		params.page*limit,
		liste[0].ID,
		params.search+"%",
		"%"+params.search+"%",
		params.ignoreRoles,
		params.ignoreZone,
		params.username,
	)

	if err != nil {
		return searchResult{}, errorToJSON(500, err)
	}

	var search searchResult
	for rows.Next() {
		var r Summary
		err := rows.Scan(&r.Siret,
			&r.Siren,
			&r.RaisonSociale,
			&r.Commune,
			&r.LibelleDepartement,
			&r.Departement,
			&r.Score,
			&r.Diff,
			&r.DernierCA,
			&r.ArreteBilan,
			&r.VariationCA,
			&r.DernierREXP,
			&r.DernierEffectif,
			&r.LibelleActivite,
			&r.LibelleActiviteN1,
			&r.CodeActivite,
			&r.EtatProcol,
			&r.ActivitePartielle,
			&r.HausseUrssaf,
			&r.Alert,
			&search.Total,
			&r.Visible,
			&r.InZone,
			&r.Followed,
			&r.Siege,
			&r.Groupe,
			&r.TerrInd,
		)
		if err != nil {
			return searchResult{}, errorToJSON(500, err)
		}
		search.Results = append(search.Results, r)
	}

	if viper.GetInt("searchPageLength")*params.page >= search.Total {
		return searchResult{}, newJSONerror(204, "empty page")
	}

	search.From = limit*params.page + 1
	search.To = limit*params.page + len(search.Results)
	search.Page = params.page
	search.PageMax = (search.Total - 1) / limit

	return search, nil
}
