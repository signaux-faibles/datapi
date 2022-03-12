package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Token type pour faire persister en mémoire les tokens réutilisables
type Token struct {
	Value        string
	CreationDate time.Time
}

// WekanLogin type pour les réponses du service de login
type WekanLogin struct {
	ID           string `json:"id"`
	Token        string `json:"token"`
	TokenExpires string `json:"tokenExpires"`
}

// WekanCreateToken type pour les réponses du service de création de token
type WekanCreateToken struct {
	ID        string `json:"_id"`
	AuthToken string `json:"authToken"`
}

// WekanGetCard type pour les réponses du service de recherche de carte
type WekanGetCard []struct {
	ID          string   `json:"_id"`
	ListID      string   `json:"listId"`
	Description string   `json:"description"`
	Members     []string `json:"members"`
}

// WekanCreateCard type pour les réponses du service de création de carte
type WekanCreateCard struct {
	ID string `json:"_id"`
}

// CustomField type pour les champs personnalisés lors de l'édition de carte
type CustomField struct {
	ID    string `json:"_id" bson:"_id"`
	Value string `json:"value" bson:"value"`
}

// FicheCRP type pour les fiches CRP
type FicheCRP struct {
	Siret       string
	Commentaire string
	Description string
}

// EtablissementData type pour les données de l'établissement à faire figurer sur la carte Wekan
type EtablissementData struct {
	Siret           string
	RaisonSociale   string
	Departement     string
	Region          string
	Effectif        int
	CodeActivite    string
	LibelleActivite string
}

// var tokens sync.Map

func lookupCard(siret string) (WekanCards, error) {
	var cards WekanCards
	var queryOr []bson.M
	for _, boardConfig := range wekanConfig.copy().BoardIds {
		queryOr = append(queryOr,
			bson.M{
				"boardId": boardConfig.BoardID,
				"customFields": bson.D{
					{Key: "_id", Value: boardConfig.CustomFields.SiretField},
					{Key: "value", Value: siret},
				},
			},
		)
	}

	query := bson.M{
		"$or": queryOr,
	}
	cursor, err := mgoDB.Collection("cards").Find(context.Background(), query)
	if err != nil {
		return nil, err
	}
	err = cursor.All(context.Background(), &cards)

	return cards, err
}

func wekanGetCardsHandler(c *gin.Context) {
	type CardData struct {
		CardID          string `json:"cardID,omitempty"`
		List            string `json:"listIndex,omitempty"`
		Archived        bool   `json:"archived,omitempty"`
		Board           string `json:"board,omitempty"`
		CardURL         string `json:"cardURL,omitempty"`
		CardDescription string `json:"cardDescription,omitempty"`
		IsMember        bool   `json:"isMember,omitempty"`
		Creator         string `json:"creator,omitempty"`
		IsBoardMember   bool   `json:"isBoardMember,omitempty"`
	}

	type BoardData struct {
		IsMember bool      `json:"isMember"`
		ID       string    `json:"id,omitempty"`
		Title    string    `json:"title,omitempty"`
		Slug     string    `json:"slug,omitempty"`
		URL      string    `json:"url,omitempty"`
		Card     *CardData `json:"card"`
	}

	var s session
	s.bind(c)
	userID := wekanConfig.userID(s.username)
	siret := c.Param("siret")
	dept, err := getDeptFromSiret(siret)

	if err != nil {
		c.JSON(500, "wekanGetCards/getDeptFromSiret: "+err.Error())
	}
	wc := wekanConfig.copy()

	if userID == "" || !s.hasRole("wekan") {
		c.JSON(403, "not a wekan user")
		return
	}

	cards, err := lookupCard(siret)
	if err != nil {
		c.JSON(500, "wekanGetCardsHandler/lookupCard: "+err.Error())
		return
	}

	wc = wekanConfig.copy()
	wcu := wc.forUser(s.username)

	boardsMap := make(map[*WekanConfigBoard]*WekanCard)
	for _, board := range wcu.Boards {
		if _, ok := board.Swimlanes[dept]; ok {
			boardsMap[board] = nil
		}
	}
	for _, card := range cards {
		c := card
		boardsMap[wc.BoardIds[card.BoardId]] = &c
	}
	wekanURL := viper.GetString("wekanURL")
	var boardDatas []BoardData
	for board, card := range boardsMap {
		var boardData BoardData
		boardData.Title = board.Title

		if _, ok := wcu.BoardIds[board.BoardID]; ok {
			boardData.ID = board.BoardID
			boardData.Slug = board.Slug
			boardData.URL = wekanURL + "b/" + board.BoardID + "/" + wc.BoardIds[board.BoardID].Slug
			boardData.IsMember = true
			if card != nil {
				boardData.Card = &CardData{
					CardID:          card.ID,
					List:            wc.listForListID(card.ListID, card.BoardId),
					Archived:        card.Archived,
					Board:           wc.BoardIds[card.BoardId].Title,
					CardURL:         wekanURL + "b/" + card.BoardId + "/" + wc.BoardIds[card.BoardId].Slug + "/" + card.ID,
					CardDescription: card.Description,
					IsMember:        contains(card.Members, wc.userID(s.username)),
					Creator:         wc.userForUserID(card.UserID),
				}
			}
		} else {
			boardData.IsMember = false
			boardData.Card = &CardData{
				Board:    wc.BoardIds[card.BoardId].Title,
				IsMember: false,
				Creator:  wc.userForUserID(card.UserID),
			}
		}
		boardDatas = append(boardDatas, boardData)
	}

	sort.Slice(boardDatas, func(i, j int) bool {
		return boardDatas[i].Title < boardDatas[j].Title
	})

	if len(boardDatas) > 0 {
		c.JSON(200, boardDatas)
	} else {
		c.JSON(204, nil)
	}
}

func wekanPartCard(userID string, siret string, boardIds []string) error {
	query := bson.M{
		"type":     "cardType-card",
		"boardId":  bson.M{"$in": boardIds},
		"archived": false,
		"$expr": bson.M{
			"$and": bson.A{
				bson.M{"$in": bson.A{userID, "$members"}},
				bson.M{"$in": bson.A{siret, "$customFields.value"}},
			}}}

	update := bson.M{
		"$pull": bson.M{
			"members": userID,
		},
	}
	_, err := mgoDB.Collection("cards").UpdateOne(context.Background(), query, update)
	return err
}

func wekanJoinCardHandler(c *gin.Context) {
	var s session
	s.bind(c)
	cardId := c.Params.ByName("cardId")
	userID := wekanConfig.userID(s.username)

	if userID == "" || !s.hasRole("wekan") {
		c.AbortWithStatusJSON(403, "not a wekan user")
		return
	}
	boardIds := wekanConfig.boardIdsForUser(s.username)
	query := bson.M{
		"_id":      cardId,
		"boardId":  bson.M{"$in": boardIds},
		"archived": false,
		"$expr":    bson.M{"$not": bson.M{"$in": bson.A{userID, "$members"}}},
	}
	update := bson.M{
		"$push": bson.M{
			"members": userID,
		},
	}
	_, err := mgoDB.Collection("cards").UpdateOne(
		context.Background(),
		query,
		update,
	)

	if err != nil {
		c.AbortWithStatusJSON(500, err.Error())
		return
	}
	c.JSON(201, "suivi effectué")
}

func wekanNewCardHandler(c *gin.Context) {
	type poker struct {
		Question             bool     `bson:"question"`
		One                  []string `bson:"one"`
		Two                  []string `bson:"two"`
		Three                []string `bson:"three"`
		Five                 []string `bson:"five"`
		Eight                []string `bson:"eight"`
		Thirteen             []string `bson:"thirteen"`
		Twenty               []string `bson:"twenty"`
		Forty                []string `bson:"forty"`
		OneHundred           []string `bson:"oneHundred"`
		Unsure               []string `bson:"unsure"`
		End                  *bool    `bson:"end"`
		AllowNonBoardMembers bool     `bson:"allowNonBoardMembers"`
	}
	type vote struct {
		Question             string   `bson:"question"`
		Positive             []string `bson:"positive"`
		Negative             []string `bson:"negative"`
		End                  *bool    `bson:"end"`
		Public               bool     `bson:"public"`
		AllowNonBoardMembers bool     `bson:"allowNonBoardMembers"`
	}
	type newCard struct {
		ID               string        `bson:"_id"`
		Title            string        `bson:"title"`
		Members          []string      `bson:"members"`
		LabelIDs         []string      `bson:"labelIds"`
		CustomFields     []CustomField `bson:"customFields"`
		ListID           string        `bson:"listId"`
		BoardID          string        `bson:"boardId"`
		Sort             float64       `bson:"sort"`
		SwimlaneID       string        `bson:"swimlaneId"`
		Type             string        `bson:"type"`
		Archived         bool          `bson:"archived"`
		ParentID         string        `bson:"parentId"`
		CoverID          string        `bson:"coverId"`
		CreatedAt        time.Time     `bson:"createdAt"`
		ModifiedAt       time.Time     `bson:"modifiedAt"`
		DateLastActivity time.Time     `bson:"dateLastActivity"`
		Description      string        `bson:"description"`
		RequestedBy      string        `bson:"requestedBy"`
		AssignedBy       string        `bson:"assignedBy"`
		Assignees        []string      `bson:"assignees"`
		SpentTime        int           `bson:"spentTime"`
		IsOverTime       bool          `bson:"isOvertime"`
		UserID           string        `bson:"userId"`
		SubtaskSort      int           `bson:"subtaskSort"`
		LinkedID         string        `bson:"linkedId"`
		Vote             vote          `bson:"vote"`
		Poker            poker         `bson:"poker"`
		TargetIDGantt    []string      `bson:"targetId_gantt"`
		LinkTypeGantt    []string      `bson:"linkType_gantt"`
		LinkIDGantt      []string      `bson:"linkId_gantt"`
		StartAt          time.Time     `bson:"startAt"`
	}
	type newActivity struct {
		ID           string    `bson:"_id"`
		UserID       string    `bson:"userId"`
		ActivityType string    `bson:"activityType"`
		BoardID      string    `bson:"boardId"`
		ListName     string    `bson:"listName"`
		ListID       string    `bson:"listId"`
		CardID       string    `bson:"cardId"`
		CardTitle    string    `bson:"cardTitle"`
		SwimlaneName string    `bson:"swimlaneName"`
		SwimlaneID   string    `bson:"swimlaneId"`
		CreatedAt    time.Time `bson:"createdAt"`
		ModifiedAt   time.Time `bson:"modifiedAt"`
	}

	var s session
	var param struct {
		BoardID     string `json:"boardId"`
		Description string `json:"description"`
	}

	s.bind(c)
	siret := c.Param("siret")
	err := c.Bind(&param)
	if err != nil {
		c.JSON(400, "wekanNewCardHandler: "+err.Error())
	}

	wcu := wekanConfig.forUser(s.username)
	if !contains(wekanConfig.boardIdsForUser(s.username), param.BoardID) {
		c.JSON(403, "wekanNewCardHandler: accès refusé")
		return
	}
	userId := wekanConfig.userID(s.username)
	etsData, err := getEtablissementDataFromDb(siret)
	if err != nil {
		c.JSON(500, "wekanNewCardHandler/getEtablissementData: "+err.Error())
		return
	}

	board, ok := wcu.BoardIds[param.BoardID]
	if !ok {
		c.JSON(500, "quelque chose ne va vraiment pas avec wekanConfig")
		return
	}

	listID := board.Lists[0]

	swimlane, ok := board.Swimlanes[etsData.Departement]
	if !ok {
		c.JSON(500, "wekanNewCardHandler: le département ne correspond pas à cette board")
		return
	}

	now := time.Now()
	cardID, err := getNewCardID(0)
	if err != nil {
		c.JSON(500, "wekanNewCardHandler: erreur de génération de la carte -> "+err.Error())
		return
	}
	activityID, err := getNewActivityID(0)
	if err != nil {
		c.JSON(500, "wekanNewCardHandler: erreur de génération de la carte -> "+err.Error())
		return
	}

	customFields := []CustomField{
		board.activiteField(etsData.CodeActivite, etsData.LibelleActivite),
		board.effectifField(etsData.Effectif),
		{
			ID:    board.CustomFields.ContactField,
			Value: "",
		},
		{
			ID:    board.CustomFields.SiretField,
			Value: siret,
		},
		{
			ID:    board.CustomFields.FicheSFField,
			Value: viper.GetString("webBaseURL") + "ets/" + siret,
		},
	}

	card := newCard{
		ID:               cardID,
		Title:            etsData.RaisonSociale,
		BoardID:          board.BoardID,
		ListID:           listID,
		Description:      param.Description,
		UserID:           userId,
		SwimlaneID:       swimlane.ID,
		Sort:             0,
		Members:          []string{userId},
		Archived:         false,
		ParentID:         "",
		CoverID:          "",
		CreatedAt:        now,
		ModifiedAt:       now,
		CustomFields:     customFields,
		DateLastActivity: now,
		RequestedBy:      "",
		AssignedBy:       "",
		LabelIDs:         []string{},
		Assignees:        []string{},
		SpentTime:        0,
		IsOverTime:       false,
		SubtaskSort:      -1,
		Type:             "cardType-card",
		LinkedID:         "",
		Vote: vote{
			Question:             "",
			Positive:             []string{},
			Negative:             []string{},
			End:                  nil,
			Public:               false,
			AllowNonBoardMembers: false,
		},
		Poker: poker{
			Question:             false,
			One:                  []string{},
			Two:                  []string{},
			Three:                []string{},
			Five:                 []string{},
			Eight:                []string{},
			Thirteen:             []string{},
			Twenty:               []string{},
			Forty:                []string{},
			OneHundred:           []string{},
			Unsure:               []string{},
			End:                  nil,
			AllowNonBoardMembers: false,
		},
		TargetIDGantt: []string{},
		LinkTypeGantt: []string{},
		LinkIDGantt:   []string{},
		StartAt:       now,
	}

	activity := newActivity{
		ID:           activityID,
		UserID:       userId,
		ActivityType: "createCard",
		BoardID:      board.BoardID,
		ListName:     "À définir",
		ListID:       listID,
		CardID:       cardID,
		CardTitle:    etsData.RaisonSociale,
		SwimlaneName: swimlane.Title,
		SwimlaneID:   swimlane.ID,
		CreatedAt:    now,
		ModifiedAt:   now,
	}

	_, err = mgoDB.Collection("activities").InsertOne(context.Background(), activity)
	if err != nil {
		c.JSON(500, "wekanNewCardHandler:"+err.Error())
		return
	}

	_, err = mgoDB.Collection("cards").InsertOne(context.Background(), card)
	if err != nil {
		c.JSON(500, "wekanNewCardHandler:"+err.Error())
		return
	}
}

// Génère un ID de la même façon de Meteor et teste s'il est disponible dans la collection cards
func getNewCardID(try int) (string, error) {
	if try < 4 {
		id := meteorID()
		res := mgoDB.Collection("cards").FindOne(context.Background(), bson.M{
			"_id": id,
		})
		if res.Err() == mongo.ErrNoDocuments {
			return id, nil
		}
		return getNewCardID(try + 1)
	} else {
		return "", errors.New("getNewCardID: pas de nouvel ID disponible au 4° essais")
	}
}

// Génère un ID de la même façon de Meteor et teste s'il est disponible dans la collection cards
func getNewActivityID(try int) (string, error) {
	if try < 4 {
		id := meteorID()
		res := mgoDB.Collection("activities").FindOne(context.Background(), bson.M{
			"_id": id,
		})
		if res.Err() == mongo.ErrNoDocuments {
			return id, nil
		}
		return getNewCardID(try + 1)
	} else {
		return "", errors.New("getNewActivityID: pas de nouvel ID disponible au 4° essais")
	}
}

func getEtablissementDataFromDb(siret string) (EtablissementData, error) {
	sql := `select s.siret,
	coalesce(s.raison_sociale, ''), 
	coalesce(s.code_departement, ''), 
	coalesce(r.libelle, ''),
	coalesce(s.effectif, 0),
	coalesce(s.code_activite, ''), 
	coalesce(s.libelle_n5, '') 
	from v_summaries s
	inner join departements d on d.code = s.code_departement
	inner join regions r on r.id = d.id_region
	where s.siret = $1`
	rows, err := db.Query(context.Background(), sql, siret)
	var etsData EtablissementData
	if err != nil {
		return etsData, err
	}
	for rows.Next() {
		err := rows.Scan(&etsData.Siret, &etsData.RaisonSociale, &etsData.Departement, &etsData.Region, &etsData.Effectif, &etsData.CodeActivite, &etsData.LibelleActivite)
		if err != nil {
			return etsData, err
		}
	}
	return etsData, nil
}

func (wcb *WekanConfigBoard) activiteField(codeActivite string, libelleActivite string) CustomField {
	var activite string
	if len(libelleActivite) > 0 && len(codeActivite) > 0 {
		activite = libelleActivite + " (" + codeActivite + ")"
	}
	return CustomField{
		ID:    wcb.CustomFields.ActiviteField,
		Value: activite,
	}
}

func (wcb *WekanConfigBoard) effectifField(effectif int) CustomField {
	var effectifIndex = -1
	effectifClass := [4]int{10, 20, 50, 100}
	for i := len(effectifClass) - 1; i >= 0; i-- {
		if effectif >= effectifClass[i] {
			effectifIndex = i
			break
		}
	}
	if effectifIndex >= 0 {
		return CustomField{wcb.CustomFields.EffectifField.EffectifFieldID, wcb.CustomFields.EffectifField.EffectifFieldItems[effectifIndex]}
	} else {
		return CustomField{wcb.CustomFields.EffectifField.EffectifFieldID, ""}
	}
}

// WekanCards est une liste de WekanCard
type WekanCards []WekanCard

// WekanCard embarque la sélection de champs Wekan nécessaires au traitement d'export
type WekanCard struct {
	ID           string     `bson:"_id"`
	Title        string     `bson:"title"`
	ListID       string     `bson:"listId"`
	BoardId      string     `bson:"boardId"`
	SwimlaneID   string     `bson:"swimlaneId"`
	Members      []string   `bson:"members"`
	Assignees    []string   `bson:"assignees"`
	Rank         float64    `bson:"sort"`
	Description  string     `bson:"description"`
	StartAt      time.Time  `bson:"startAt"`
	EndAt        *time.Time `bson:"endAt"`
	LabelIds     []string   `bson:"labelIds"`
	Archived     bool       `bson:"archived"`
	UserID       string     `bson:"userId"`
	Comments     []string   `bson:"comments"`
	CustomFields []struct {
		ID    string `bson:"_id"`
		Value string `bson:"value"`
	}
}

// Board retourne le nom de la board où est la carte
func (c WekanCard) Board() string {
	return wekanConfig.BoardForID(c.BoardId)
}

func (cs WekanCards) Sirets() []string {
	var sirets []string
	for _, c := range cs {
		siret, err := c.Siret()
		if err != nil {
			fmt.Printf("WekanCards.Sirets(): %s\n", err.Error())
			continue
		}
		sirets = append(sirets, siret)
	}
	return sirets
}

func (c WekanCard) Siret() (string, error) {
	wekanConfig.mu.Lock()
	defer wekanConfig.mu.Unlock()
	board, ok := wekanConfig.BoardIds[c.BoardId]
	if !ok {
		return "", fmt.Errorf("WekanCard.Siret(), pas de board: %s", c.BoardId)
	}
	for _, field := range c.CustomFields {
		if field.ID == board.CustomFields.SiretField {
			return field.Value, nil
		}
	}
	return "", fmt.Errorf("pas de propriété SIRET pour cette carte: %s", c.ID)
}

func (c WekanCard) Activite() (string, error) {
	wekanConfig.mu.Lock()
	defer wekanConfig.mu.Unlock()
	board, ok := wekanConfig.BoardIds[c.BoardId]
	if !ok {
		return "", fmt.Errorf("WekanCard.Activite(), pas de board: %s", c.BoardId)
	}
	for _, field := range c.CustomFields {
		if field.ID == board.CustomFields.ActiviteField {
			return field.Value, nil
		}
	}
	return "", fmt.Errorf("WekanCard.Activite(), pas de propriété Activité pour cette carte: %s", c.ID)
}

func (c WekanCard) Effectif() (int, error) {
	wekanConfig.mu.Lock()
	board, ok := wekanConfig.BoardIds[c.BoardId]
	defer wekanConfig.mu.Unlock()

	if !ok {
		return -1, fmt.Errorf("WekanCard.Effectif(), pas de board: %s", c.BoardId)
	}
	for _, field := range c.CustomFields {
		if field.ID == board.CustomFields.EffectifField.EffectifFieldID {
			for k, v := range board.CustomFields.EffectifField.EffectifFieldItems {
				if v == field.Value {
					return k, nil
				}
			}
		}
	}
	return -1, fmt.Errorf("WekanCard.Effectif(), pas de propriété Effectif pour cette carte: %s", c.ID)
}

func (c WekanCard) FicheSF() (string, error) {
	wekanConfig.mu.Lock()
	defer wekanConfig.mu.Unlock()

	board, ok := wekanConfig.BoardIds[c.BoardId]
	if !ok {
		return "", fmt.Errorf("pas de board: %s", c.BoardId)
	}
	for _, field := range c.CustomFields {
		if field.ID == board.CustomFields.FicheSFField {
			return field.Value, nil
		}
	}
	return "", fmt.Errorf("pas de propriété FicheSF pour cette carte: %s", c.ID)
}

func selectWekanCardsFromSiret(username string, siret string) ([]*WekanCard, error) {
	wcu := wekanConfig.forUser(username)
	userID := wekanConfig.userID(username)
	if userID == "" {
		return nil, fmt.Errorf("selectWekanCardFromSiret() -> utilisateur wekan inconnu: %s", username)
	}

	var wekanCards []*WekanCard

	pipeline := cardPipeline(wcu, siret)
	cursor, err := mgoDB.Collection("cards").Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}

	err = cursor.All(context.Background(), &wekanCards)
	if err != nil {
		return nil, err
	}

	return wekanCards, nil
}

func cardPipeline(wcu WekanConfig, siret string) bson.A {
	var queryOr []bson.M
	for _, boardConfig := range wcu.Boards {
		queryOr = append(queryOr,
			bson.M{
				"boardId": boardConfig.BoardID,
				"customFields": bson.D{
					{Key: "_id", Value: boardConfig.CustomFields.SiretField},
					{Key: "value", Value: siret},
				},
			},
		)
	}

	pipeline := bson.A{
		bson.M{
			"$match": bson.M{
				"$or": queryOr,
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from": "card_comments",
				"let":  bson.M{"card": "$_id"},
				"pipeline": bson.A{
					bson.M{
						"$match": bson.M{
							"$expr": bson.M{"$eq": bson.A{"$cardId", "$$card"}},
							"text":  bson.M{"$regex": "#export"},
						},
					}, bson.M{
						"$sort": bson.M{
							"createdAt": 1,
						},
					}, bson.M{
						"$project": bson.M{
							"text": 1,
							"_id":  0,
						},
					},
				},
				"as": "comments",
			},
		},
		bson.M{
			"$sort": bson.D{
				bson.E{Key: "archived", Value: 1},
				bson.E{Key: "startAt", Value: 1},
			},
		},
		bson.M{
			"$project": bson.M{
				"archived":     1,
				"title":        1,
				"listId":       1,
				"boardId":      1,
				"members":      1,
				"swimlaneId":   1,
				"customFields": 1,
				"sort":         1,
				"labelIds":     1,
				"startAt":      1,
				"endAt":        1,
				"description":  1,
				"comments":     "$comments.text",
			},
		},
	}
	return pipeline
}

func selectWekanCards(username *string, boardIds []string, swimlaneIds []string, listIds []string, labelIds [][]labelID) ([]*WekanCard, error) {
	if username != nil {
		userID := wekanConfig.userID(*username)
		if userID == "" {
			return nil, fmt.Errorf("selectWekanCards() -> utilisateur wekan inconnu: %s", *username)
		}
	}

	pipeline := cardsPipeline(username, boardIds, swimlaneIds, listIds, labelIds)

	cardsCursor, err := mgoDB.Collection("cards").Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	var wekanCards []*WekanCard
	err = cardsCursor.All(context.Background(), &wekanCards)
	if err != nil {
		return nil, err
	}
	return wekanCards, nil
}

func cardsPipeline(username *string, boardIds []string, swimlaneIds []string, listIds []string, labelIds [][]labelID) bson.A {
	query := bson.M{
		"type": "cardType-card",
	}

	if username != nil {
		userID := wekanConfig.userID(*username)
		query["$expr"] = bson.M{
			"$or": bson.A{
				bson.M{"$in": bson.A{userID, "$members"}},
				bson.M{"$in": bson.A{userID, "$assignees"}},
			},
		}
	}

	if len(boardIds) > 0 { // rejet des ids non connus dans la configuration
		var ids []string
		for _, id := range boardIds {
			wekanConfig.mu.Lock()
			if _, ok := wekanConfig.BoardIds[id]; ok {
				ids = append(ids, id)
			}
			wekanConfig.mu.Unlock()
		}
		query["boardId"] = bson.M{"$in": ids}
	} else { // limiter au périmètre des boards de la configuration
		query["boardId"] = bson.M{"$in": wekanConfig.boardIds()}
	}

	if len(swimlaneIds) > 0 {
		query["swimlaneId"] = bson.M{"$in": swimlaneIds}
	}
	if len(listIds) > 0 {
		query["listId"] = bson.M{"$in": listIds}
	}

	if len(labelIds) > 0 {
		labelQueryAnd := bson.A{}
		for _, l := range labelIds {
			labelQueryOr := bson.A{}
			for _, i := range l {
				labelQueryOr = append(labelQueryOr, bson.M{
					"labelIds": i.labelID,
					"boardId":  i.boardID,
				})
			}
			labelQueryAnd = append(labelQueryAnd, bson.M{
				"$or": labelQueryOr,
			})
		}
		query["$and"] = labelQueryAnd
	}

	pipeline := bson.A{
		bson.M{
			"$match": query,
		},
		bson.M{
			"$lookup": bson.M{
				"from": "card_comments",
				"let":  bson.M{"card": "$_id"},
				"pipeline": bson.A{
					bson.M{
						"$match": bson.M{
							"$expr": bson.M{"$eq": bson.A{"$cardId", "$$card"}},
							"text":  bson.M{"$regex": "#export"},
						},
					}, bson.M{
						"$sort": bson.M{
							"createdAt": -1,
						},
					}, bson.M{
						"$project": bson.M{
							"text": 1,
							"_id":  0,
						},
					},
				},
				"as": "comments",
			},
		},
		bson.M{
			"$sort": bson.D{
				bson.E{Key: "archived", Value: 1},
				bson.E{Key: "startAt", Value: 1},
			},
		},
		bson.M{
			"$project": bson.M{
				"archived":     1,
				"title":        1,
				"listId":       1,
				"boardId":      1,
				"members":      1,
				"swimlaneId":   1,
				"customFields": 1,
				"sort":         1,
				"labelIds":     1,
				"startAt":      1,
				"endAt":        1,
				"description":  1,
				"assignees":    1,
				"comments":     "$comments.text",
			},
		},
	}

	return pipeline
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func meteorID() string {
	rand.Seed(time.Now().UTC().UnixNano())
	b := make([]rune, 17)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
