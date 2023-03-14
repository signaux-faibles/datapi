package core

import (
	"context"
	"fmt"
	"github.com/signaux-faibles/datapi/src/db"
	"github.com/signaux-faibles/datapi/src/utils"
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
	s.bind(c)
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
		Username:         &s.username,
		UnfollowComment:  param.UnfollowComment,
		UnfollowCategory: param.UnfollowCategory,
	}
	userID := wekanConfig.userID(s.username)
	if userID != "" && s.hasRole("wekan") {
		boardIds := wekanConfig.boardIdsForUser(s.username)
		err := wekanPartCard(userID, siret, boardIds)
		if err != nil {
			fmt.Println(err)
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

	params := summaryParams{roles.zoneGeo(), nil, nil, &liste[0].ID, false, nil,
		&True, &True, *f.Username, false, "follow", &False, nil,
		nil, &True, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil}

	sms, err := getSummaries(params)
	if err != nil {
		return nil, utils.ErrorToJSON(500, err)
	}
	var follows Follows

	for _, s := range sms.summaries {
		var f Follow
		f.Comment = *coalescepString(s.Comment, &EmptyString)
		f.Category = *coalescepString(s.Category, &EmptyString)
		f.Since = *coalescepTime(s.Since, &time.Time{})
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

func getCardsForCurrentUser(c *gin.Context) {
	var s session
	s.bind(c)
	var params paramsGetCards
	err := c.Bind(&params)
	if err != nil {
		c.JSON(400, fmt.Sprintf("parametre incorrect: %s", err.Error()))
		return
	}
	types := []string{"no-card", "my-cards", "all-cards"}
	if !utils.Contains(types, params.Type) {
		c.AbortWithStatusJSON(400, fmt.Sprintf("`%s` n'est pas un type supporté", params.Type))
		return
	}
	cards, err := getCards(s, params)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}
	var follows = make(Follows, 0)
	for _, s := range cards {
		var f Follow
		if s.Summary != nil {
			if s.Summary.Since == nil {
				f.Comment = *coalescepString(s.Summary.Comment, &EmptyString)
				f.Category = *coalescepString(s.Summary.Category, &EmptyString)
				f.Since = *coalescepTime(s.Summary.Since, &time.Time{})
				f.Active = true
			}
			f.EtablissementSummary = s.Summary
			follows = append(follows, f)
		}
	}
	c.JSON(200, follows)
}

type paramsGetCards struct {
	Type   string     `json:"type"`
	Statut []string   `json:"statut"`
	Zone   []string   `json:"zone"`
	Labels []string   `json:"labels"`
	Since  *time.Time `json:"since"`
}

type Card struct {
	Summary    *Summary     `json:"summary"`
	WekanCards []*WekanCard `json:"wekanCard"`
	dbExport   *dbExport
}

type Cards []*Card

func (cards Cards) dbExportsOnly() Cards {
	var filtered Cards
	for _, c := range cards {
		if c.dbExport != nil {
			filtered = append(filtered, c)
		}
	}
	return filtered
}

func getCards(s session, params paramsGetCards) ([]*Card, error) {
	var cards []*Card
	var cardsMap = make(map[string]*Card)
	var sirets []string
	var followedSirets []string
	wcu := wekanConfig.forUser(s.username)
	userID := wekanConfig.userID(s.username)
	labelIds := wcu.labelIdsForLabels(params.Labels)
	if userID != "" && s.hasRole("wekan") && params.Type != "no-card" {
		var username *string
		if params.Type == "my-cards" {
			username = &s.username
		}
		boardIds := wcu.boardIds()
		swimlaneIds := wcu.swimlaneIdsForZone(params.Zone)
		listIds := wcu.listIdsForStatuts(params.Statut)
		wekanCards, err := selectWekanCards(username, boardIds, swimlaneIds, listIds, labelIds, params.Since)
		if err != nil {
			return nil, err
		}
		for _, w := range wekanCards {
			siret, err := w.Siret()
			if err != nil {
				continue
			}
			card := Card{nil, []*WekanCard{w}, nil}
			cards = append(cards, &card)
			cardsMap[siret] = &card
			sirets = append(sirets, siret)
			if utils.Contains(append(w.Members, w.Assignees...), userID) {
				followedSirets = append(followedSirets, siret)
			}
		}
		err = followSiretsFromWekan(s.username, followedSirets)
		if err != nil {
			return nil, err
		}
		var ss summaries
		cursor, err := db.Get().Query(context.Background(), sqlGetCards, s.roles.zoneGeo(), s.username, sirets)
		if err != nil {
			return nil, err
		}
		for cursor.Next() {
			s := ss.newSummary()
			cursor.Scan(s...)
		}
		for _, s := range ss.summaries {
			cardsMap[s.Siret].Summary = s
		}
	} else {
		boardIds := wcu.boardIds()
		wekanCards, err := selectWekanCards(&s.username, boardIds, nil, nil, nil, nil)
		if err != nil {
			return nil, err
		}
		var excludeSirets = make(map[string]struct{})
		for _, w := range wekanCards {
			siret, err := w.Siret()
			if err != nil {
				continue
			}
			excludeSirets[siret] = struct{}{}
		}
		var ss summaries
		cursor, err := db.Get().Query(context.Background(), sqlGetFollow, s.roles.zoneGeo(), s.username, params.Zone)
		if err != nil {
			return nil, err
		}
		for cursor.Next() {
			s := ss.newSummary()
			cursor.Scan(s...)
		}
		for _, s := range ss.summaries {
			if _, ok := excludeSirets[s.Siret]; !ok {
				card := Card{s, nil, nil}
				cards = append(cards, &card)
			}
		}
	}
	return cards, nil
}

func followSiretsFromWekan(username string, sirets []string) error {
	tx, err := db.Get().Begin(context.Background())
	if err != nil {
		return err
	}
	if _, err := tx.Exec(context.Background(), sqlCreateTmpFollowWekan, sirets, username); err != nil {
		return err
	}
	if _, err := tx.Exec(context.Background(), sqlFollowFromTmp, username); err != nil {
		return err
	}
	return tx.Commit(context.Background())
}

// sqlCreateTmpFollowWekan
const sqlCreateTmpFollowWekan = `create temporary table tmp_follow_wekan on commit drop as
	with sirets as (select unnest($1::text[]) as siret),
	follow as (select siret from etablissement_follow f where f.username = $2 and active)
	select case when f.siret is null then 'follow' else 'unfollow' end as todo,
	coalesce(f.siret, s.siret) as siret
	from follow f
	full join sirets s on s.siret = f.siret 
	where f.siret is null or s.siret is null;
`

// sqlFollowFromTmp $1 = username
const sqlFollowFromTmp = `insert into etablissement_follow
	(siret, siren, username, active, since, comment, category)
	select t.siret, substring(t.siret from 1 for 9), $1, 
	true, current_timestamp, 'participe à la carte wekan', 'wekan'
	from tmp_follow_wekan t
	inner join etablissement0 e on e.siret = t.siret
	where todo = 'follow'`

// sqlGetCards: $1 = roles.ZoneGeo, $2 = username, $3 = sirets
const sqlGetCards = `select 
s.siret, s.siren, s.raison_sociale, s.commune,
s.libelle_departement, s.code_departement,
case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score then s.valeur_score end as valeur_score,
case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score then s.detail_score end as detail_score,
case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score then s.first_alert end as first_alert,
s.chiffre_affaire, s.arrete_bilan, s.exercice_diane, s.variation_ca, s.resultat_expl, s.effectif, s.effectif_entreprise,
s.libelle_n5, s.libelle_n1, s.code_activite, s.last_procol,
case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).dgefp then s.activite_partielle end as activite_partielle,
case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).dgefp then s.apconso_heure_consomme end as apconso_heure_consomme,
case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).dgefp then s.apconso_montant end as apconso_montant,
case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).urssaf then s.hausse_urssaf end as hausse_urssaf,
case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).urssaf then s.dette_urssaf end as dette_urssaf,
case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score then s.alert end,
count(*) over () as nb_total,
count(case when s.alert='Alerte seuil F1' and (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score then 1 end) over () as nb_f1,
count(case when s.alert='Alerte seuil F2' and (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score then 1 end) over () as nb_f2,
(permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).visible, 
(permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).in_zone, 
f.id is not null as followed_etablissement,
fe.siren is not null as followed_entreprise,
s.siege, s.raison_sociale_groupe, territoire_industrie,
f.comment, f.category, f.since,
(permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).urssaf,
(permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).dgefp,
(permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score,
(permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).bdf,
s.secteur_covid, s.excedent_brut_d_exploitation, s.etat_administratif, s.etat_administratif_entreprise
from v_summaries s
left join etablissement_follow f on f.active and f.siret = s.siret and f.username = $2
left join v_entreprise_follow fe on fe.siren = s.siren and fe.username = $2
where s.siret = any($3)`

const sqlGetFollow = `select 
s.siret, s.siren, s.raison_sociale, s.commune,
s.libelle_departement, s.code_departement,
case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score then s.valeur_score end as valeur_score,
case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score then s.detail_score end as detail_score,
case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score then s.first_alert end as first_alert,
s.chiffre_affaire, s.arrete_bilan, s.exercice_diane, s.variation_ca, s.resultat_expl, s.effectif, s.effectif_entreprise,
s.libelle_n5, s.libelle_n1, s.code_activite, s.last_procol,
case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).dgefp then s.activite_partielle end as activite_partielle,
case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).dgefp then s.apconso_heure_consomme end as apconso_heure_consomme,
case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).dgefp then s.apconso_montant end as apconso_montant,
case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).urssaf then s.hausse_urssaf end as hausse_urssaf,
case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).urssaf then s.dette_urssaf end as dette_urssaf,
case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score then s.alert end,
count(*) over () as nb_total,
count(case when s.alert='Alerte seuil F1' and (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score then 1 end) over () as nb_f1,
count(case when s.alert='Alerte seuil F2' and (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score then 1 end) over () as nb_f2,
(permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).visible, 
(permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).in_zone, 
f.id is not null as followed_etablissement,
fe.siren is not null as followed_entreprise,
s.siege, s.raison_sociale_groupe, territoire_industrie,
f.comment, f.category, f.since,
(permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).urssaf,
(permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).dgefp,
(permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score,
(permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).bdf,
s.secteur_covid, s.excedent_brut_d_exploitation, s.etat_administratif, s.etat_administratif_entreprise
from v_summaries s
inner join etablissement_follow f on f.active and f.siret = s.siret and f.username = $2
inner join v_entreprise_follow fe on fe.siren = s.siren and fe.username = $2
where s.code_departement = any($3) or $3 is null`
