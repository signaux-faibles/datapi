package core

import (
	"context"
	"errors"
	"fmt"
	"github.com/signaux-faibles/datapi/src/db"
	"github.com/signaux-faibles/datapi/src/utils"
	"github.com/signaux-faibles/libwekan"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"math/rand"
	"net/http"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// TODO libwekan
func connectWekanDB() *mongo.Database {
	mongoURI := viper.GetString("wekanMgoURL")
	if mongoURI != "" {
		client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))
		if err != nil {
			log.Fatal(err)
		}
		err = client.Connect(context.Background())
		if err != nil {
			log.Fatal(err)
		}
		mgoDb := client.Database(viper.GetString("wekanMgoDB"))
		return mgoDb
	}
	return nil
}

// CustomField type pour les champs personnalisés lors de l'édition de carte
type CustomField struct {
	ID    string `json:"_id" bson:"_id"`
	Value string `json:"value" bson:"value"`
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

// TODO libwekan
func wekanUnarchiveCardHandler(c *gin.Context) {
	var s session
	s.bind(c)
	userID := oldWekanConfig.userID(s.username)
	if userID == "" || !s.hasRole("wekan") {
		c.JSON(http.StatusForbidden, "not a wekan user")
		return
	}

	cardID := c.Param("cardID")
	boardIDs := oldWekanConfig.boardIdsForUser(s.username)
	result, err := mgoDB.Collection("cards").UpdateOne(context.Background(),
		bson.M{
			"boardId": bson.M{
				"$in": boardIDs,
			},
			"_id": cardID,
		},
		bson.M{
			"$set": bson.M{
				"archived": false,
			},
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, "wekanUnarchiveCardHandler: "+err.Error())
		return
	}
	if result.MatchedCount == 0 {
		c.JSON(http.StatusBadRequest, "wekanUnarchiveCardHandler: carte non existante ou non accessible")
		return
	}
	if result.ModifiedCount == 0 {
		c.JSON(http.StatusNoContent, "wekanUnarchiveCardHandler: la carte n'est pas archivée")
		return
	}
}

// TODO libwekan
func wekanGetCardsHandler(c *gin.Context) {
	type CardData struct {
		CardID          string     `json:"cardId,omitempty"`
		List            string     `json:"listIndex,omitempty"`
		Archived        bool       `json:"archived,omitempty"`
		Board           string     `json:"board,omitempty"`
		CardURL         string     `json:"cardURL,omitempty"`
		CardDescription string     `json:"cardDescription,omitempty"`
		IsMember        bool       `json:"isMember,omitempty"`
		Creator         string     `json:"creator,omitempty"`
		IsBoardMember   bool       `json:"isBoardMember,omitempty"`
		LastActivity    time.Time  `json:"lastActivity,omitempty"`
		StartAt         time.Time  `json:"startAt,omitempty"`
		EndAt           *time.Time `json:"endAt,omitempty"`
		LabelIDs        []string   `json:"labelIds,omitempty"`
	}

	type BoardData struct {
		IsMember bool        `json:"isMember"`
		ID       string      `json:"id,omitempty"`
		Title    string      `json:"title,omitempty"`
		Slug     string      `json:"slug,omitempty"`
		URL      string      `json:"url,omitempty"`
		Cards    []*CardData `json:"cards"`
	}

	var s session
	s.bind(c)
	_, ok := wekanConfig.GetUserByUsername(libwekan.Username(s.username))

	if !ok || !s.hasRole("wekan") {
		c.JSON(http.StatusForbidden, "not a wekan user")
		return
	}

	siret := c.Param("siret")
	dept, err := getDeptFromSiret(siret)

	if err != nil {
		c.JSON(500, "wekanGetCards/getDeptFromSiret: "+err.Error())
	}
	wc := oldWekanConfig.copy()

	cards, err := selectWekanCardsFromSiret(s.username, siret)
	if err != nil {
		c.JSON(500, "wekanGetCardsHandler/lookupCard: "+err.Error())
		return
	}

	wc = oldWekanConfig.copy()
	wcu := wc.forUser(s.username)

	boardsMap := make(map[*WekanConfigBoard][]*WekanCard)
	for _, board := range wcu.Boards {
		if _, ok := board.Swimlanes[dept]; ok {
			boardsMap[board] = nil
		}
	}
	for _, card := range cards {
		c := card
		boardsMap[wc.BoardIds[card.BoardId]] = append(boardsMap[wc.BoardIds[card.BoardId]], c)
	}
	wekanURL := viper.GetString("wekanURL")
	var boardDatas []BoardData
	for board, cards := range boardsMap {
		var boardData BoardData
		boardData.Title = board.Title
		if _, ok := wcu.BoardIds[board.BoardID]; ok {
			boardData.ID = board.BoardID
			boardData.Slug = board.Slug
			boardData.URL = wekanURL + "b/" + board.BoardID + "/" + wc.BoardIds[board.BoardID].Slug
			boardData.IsMember = true
			for _, card := range cards {
				if card != nil {
					boardData.Cards = append(boardData.Cards, &CardData{
						CardID:          card.ID,
						List:            wc.listForListID(card.ListID, card.BoardId),
						Archived:        card.Archived,
						Board:           wc.BoardIds[card.BoardId].Title,
						CardURL:         wekanURL + "b/" + card.BoardId + "/" + wc.BoardIds[card.BoardId].Slug + "/" + card.ID,
						CardDescription: card.Description,
						IsMember:        utils.Contains(card.Members, wc.userID(s.username)),
						Creator:         wc.userForUserID(card.UserID),
						LabelIDs:        card.LabelIds,
						StartAt:         card.StartAt,
						EndAt:           card.EndAt,
						LastActivity:    card.LastActivity,
					})
				}
			}
		} else {
			boardData.IsMember = false
			for _, card := range cards {
				boardData.Cards = append(boardData.Cards, &CardData{
					Board:        wc.BoardIds[card.BoardId].Title,
					IsMember:     false,
					Archived:     card.Archived,
					Creator:      wc.userForUserID(card.UserID),
					StartAt:      card.StartAt,
					EndAt:        card.EndAt,
					LastActivity: card.LastActivity,
				})
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

// TODO libwekan
func wekanJoinCardHandler(c *gin.Context) {
	var s session
	s.bind(c)
	cardId := c.Params.ByName("cardId")
	userID := oldWekanConfig.userID(s.username)

	if userID == "" || !s.hasRole("wekan") {
		c.AbortWithStatusJSON(403, "not a wekan user")
		return
	}
	boardIds := oldWekanConfig.boardIdsForUser(s.username)
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

// TODO libwekan
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

	wcu := oldWekanConfig.forUser(s.username)
	if !utils.Contains(oldWekanConfig.boardIdsForUser(s.username), param.BoardID) {
		c.JSON(403, "wekanNewCardHandler: accès refusé")
		return
	}
	userID := oldWekanConfig.userID(s.username)
	etsData, err := getEtablissementDataFromDb(siret)
	if err != nil {
		c.JSON(500, "wekanNewCardHandler/getEtablissementData: "+err.Error())
		return
	}

	board, ok := wcu.BoardIds[param.BoardID]
	if !ok {
		c.JSON(500, "quelque chose ne va vraiment pas avec oldWekanConfig")
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
		UserID:           userID,
		SwimlaneID:       swimlane.ID,
		Sort:             0,
		Members:          []string{userID},
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
		UserID:       userID,
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

// TODO libwekan
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
	}
	return "", errors.New("getNewCardID: pas de nouvel ID disponible au 4° essais")
}

// TODO libwekan
// Génère un ID de la même façon de Meteor et teste s'il est disponible dans la collection activities
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
	}
	return "", errors.New("getNewActivityID: pas de nouvel ID disponible au 4° essais")
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
	rows, err := db.Get().Query(context.Background(), sql, siret)
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
	}
	return CustomField{wcb.CustomFields.EffectifField.EffectifFieldID, ""}
}

// TODO libwekan
// WekanCards est une liste de WekanCard
type WekanCards []WekanCard

// TODO libwekan
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
	LastActivity time.Time  `bson:"lastActivity"`
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
	return oldWekanConfig.BoardForID(c.BoardId)
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
	oldWekanConfig.mu.Lock()
	defer oldWekanConfig.mu.Unlock()
	board, ok := oldWekanConfig.BoardIds[c.BoardId]
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
	oldWekanConfig.mu.Lock()
	defer oldWekanConfig.mu.Unlock()
	board, ok := oldWekanConfig.BoardIds[c.BoardId]
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
	oldWekanConfig.mu.Lock()
	board, ok := oldWekanConfig.BoardIds[c.BoardId]
	defer oldWekanConfig.mu.Unlock()

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
	oldWekanConfig.mu.Lock()
	defer oldWekanConfig.mu.Unlock()

	board, ok := oldWekanConfig.BoardIds[c.BoardId]
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

// TODO libwekan
func selectWekanCardsFromSiret(username string, siret string) ([]*WekanCard, error) {
	wcu := oldWekanConfig.forUser(username)
	userID := oldWekanConfig.userID(username)
	if userID == "" {
		return nil, fmt.Errorf("selectWekanCardFromSiret() -> utilisateur wekan inconnu: %s", username)
	}

	var wekanCards []*WekanCard

	pipeline := cardFromSiretPipeline(wcu, siret)
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

// TODO libwekan
func cardFromSiretPipeline(wcu WekanConfig, siret string) bson.A {
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
			"$lookup": bson.M{
				"from":         "card_comments",
				"localField":   "_id",
				"foreignField": "cardId",
				"as":           "lastActivity",
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
				"lastActivity": bson.M{
					"$max": bson.A{
						bson.M{"$max": "$lastActivity.modifiedAt"},
						"$modifiedAt",
					},
				},
			},
		},
	}
	return pipeline
}

// TODO libwekan
func selectWekanCards(username *string, boardIds []string, swimlaneIds []string, listIds []string, labelIds [][]labelID, labelMode bool, since *time.Time) ([]*WekanCard, error) {
	if username != nil {
		userID := oldWekanConfig.userID(*username)
		if userID == "" {
			return nil, fmt.Errorf("selectWekanCards() -> utilisateur wekan inconnu: %s", *username)
		}
	}

	pipeline := cardsPipeline(username, boardIds, swimlaneIds, listIds, labelIds, labelMode, since)

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

func cardsPipeline(username *string, boardIds []string, swimlaneIds []string, listIds []string, labelIds [][]labelID, labelMode bool, since *time.Time) bson.A {
	query := bson.M{
		"type": "cardType-card",
	}

	if username != nil {
		userID := oldWekanConfig.userID(*username)
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
			oldWekanConfig.mu.Lock()
			if _, ok := oldWekanConfig.BoardIds[id]; ok {
				ids = append(ids, id)
			}
			oldWekanConfig.mu.Unlock()
		}
		query["boardId"] = bson.M{"$in": ids}
	} else { // limiter au périmètre des boards de la configuration
		query["boardId"] = bson.M{"$in": oldWekanConfig.boardIds()}
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
		if labelMode {
			query["$and"] = labelQueryAnd
		} else {
			query["$or"] = labelQueryAnd
		}
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
			"$lookup": bson.M{
				"from":         "card_comments",
				"localField":   "_id",
				"foreignField": "cardId",
				"as":           "lastActivity",
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
				"lastActivity": bson.M{
					"$max": bson.A{
						bson.M{"$max": "$lastActivity.modifiedAt"},
						"$dateLastActivity",
					},
				},
			},
		},
	}

	if since != nil {
		pipeline = append(pipeline, bson.M{
			"$match": bson.M{
				"lastActivity": bson.M{"$gte": since},
			},
		})
	}
	return pipeline
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

// TODO libwekan
func meteorID() string {
	b := make([]rune, 17)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
