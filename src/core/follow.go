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
	userID := oldWekanConfig.userID(s.Username)
	if userID != "" && s.hasRole("wekan") {
		boardIds := oldWekanConfig.boardIdsForUser(s.Username)
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

//func getCardsForCurrentUser(c *gin.Context) {
//	var s session
//	s.Bind(c)
//	var params paramsGetCards
//	err := c.Bind(&params)
//	if err != nil {
//		c.JSON(400, fmt.Sprintf("parametre incorrect: %s", err.Error()))
//		return
//	}
//	types := []string{"no-card", "my-cards", "all-cards"}
//	if !utils.Contains(types, params.Type) {
//		c.AbortWithStatusJSON(400, fmt.Sprintf("`%s` n'est pas un type supporté", params.Type))
//		return
//	}
//	cards, err := getCards(s, params)
//	if err != nil {
//		c.JSON(500, err.Error())
//		return
//	}
//	var follows = make(Follows, 0)
//	for _, s := range cards {
//		var f Follow
//		if s.Summary != nil {
//			if s.Summary.Since == nil {
//				f.Comment = *utils.Coalesce(s.Summary.Comment, &EmptyString)
//				f.Category = *utils.Coalesce(s.Summary.Category, &EmptyString)
//				f.Since = *utils.Coalesce(s.Summary.Since, &time.Time{})
//				f.Active = true
//			}
//			f.EtablissementSummary = s.Summary
//			follows = append(follows, f)
//		}
//	}
//	c.JSON(200, follows)
//}

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

//	func getCards(s session, params paramsGetCards) ([]*Card, error) {
//		var cards []*Card
//		var cardsMap = make(map[string]*Card)
//		var sirets []string
//		var followedSirets []string
//		wcu := oldWekanConfig.forUser(s.Username)
//		userID := oldWekanConfig.userID(s.Username)
//		labelIds := wcu.labelIdsForLabels(params.Labels)
//		labelMode := params.LabelMode
//		if userID != "" && s.hasRole("wekan") && params.Type != "no-card" {
//			var username *string
//			if params.Type == "my-cards" {
//				username = &s.Username
//			}
//
//			if len(params.Boards) == 0 {
//				params.Boards = wcu.boardIds()
//			}
//			boardIds := params.Boards
//			swimlaneIds := wcu.swimlaneIdsForZone(params.Zone)
//			listIds := wcu.listIdsForStatuts(params.Statut)
//			wekanCards, err := selectWekanCards(username, boardIds, swimlaneIds, listIds, labelIds, labelMode, params.Since)
//			if err != nil {
//				return nil, err
//			}
//			for _, w := range wekanCards {
//				siret, err := w.Siret()
//				if err != nil {
//					continue
//				}
//				card := Card{nil, []*WekanCard{w}, nil}
//				cards = append(cards, &card)
//				cardsMap[siret] = &card
//				sirets = append(sirets, siret)
//				if utils.Contains(append(w.Members, w.Assignees...), userID) {
//					followedSirets = append(followedSirets, siret)
//				}
//			}
//			err = followSiretsFromWekan(s.Username, followedSirets)
//			if err != nil {
//				return nil, err
//			}
//   var ss Summaries
//		cursor, err := db.Get().Query(context.Background(), SqlGetCards, s.roles, s.Username, sirets)
//		if err != nil {
//			return nil, err
//		}
//		for cursor.Next() {
//			s := ss.NewSummary()
//			err := cursor.Scan(s...)
//			if err != nil {
//				return nil, err
//			}
//		}
//		for _, s := range ss.summaries {
//			cardsMap[s.Siret].Summary = s
//		}
//	} else {
//		boardIds := wcu.boardIds()
//		wekanCards, err := selectWekanCards(&s.Username, boardIds, nil, nil, nil, labelMode, nil)
//		if err != nil {
//			return nil, err
//		}
//		var excludeSirets = make(map[string]struct{})
//		for _, w := range wekanCards {
//			siret, err := w.Siret()
//			if err != nil {
//				continue
//			}
//			excludeSirets[siret] = struct{}{}
//		}
//		var ss summaries
//		cursor, err := db.Get().Query(context.Background(), SqlGetFollow, s.roles, s.Username, params.Zone)
//		if err != nil {
//			return nil, err
//		}
//		for cursor.Next() {
//			s := ss.NewSummary()
//			err := cursor.Scan(s...)
//			if err != nil {
//				return nil, err
//			}
//		}
//		for _, s := range ss.summaries {
//			if _, ok := excludeSirets[s.Siret]; !ok {
//				card := Card{s, nil, nil}
//				cards = append(cards, &card)
//			}
//		}
//	}
//	return cards, nil
//}
