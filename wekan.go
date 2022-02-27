package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
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
	ID    string `json:"_id"`
	Value string `json:"value"`
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

var tokens sync.Map

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
	type newCard struct {
		BoardID     string `json:"boardId"`
		Description string `json:"description"`
	}

	siret := c.Param("siret")
	var s session
	var param newCard
	s.bind(c)
	err := c.Bind(&param)
	fmt.Println(err)
	if err != nil {
		return
	}

	boardIds := wekanConfig.boardIdsForUser(s.username)
	if !contains(boardIds, param.BoardID) {
		c.JSON(403, "User can't create card on this board "+param.BoardID)
		return
	}

	userID := wekanConfig.userID(s.username)
	if userID == "" {
		c.JSON(403, "not a wekan user")
		return
	}
	var adminToken Token
	adminUsername := viper.GetString("wekanAdminUsername")
	loadedToken, ok := tokens.Load(adminUsername)
	if loadedToken != nil {
		adminToken = loadedToken.(Token)
		ok = isValidToken(adminToken)
	}
	if !ok {
		log.Printf("no valid admin token")
		adminPassword := viper.GetString("wekanAdminPassword")
		body, err := adminLogin(adminUsername, adminPassword)
		if err != nil {
			c.JSON(500, err.Error())
			return
		}
		wekanLogin := WekanLogin{}
		err = json.Unmarshal(body, &wekanLogin)
		if err != nil {
			c.JSON(500, err.Error())
			return
		}
		adminToken.Value = wekanLogin.Token
		now := time.Now()
		adminToken.CreationDate = now
		tokens.Store(adminUsername, adminToken)
	} else {
		log.Printf("existing admin token")
	}
	var userToken Token
	loadedToken, ok = tokens.Load(s.username)
	if loadedToken != nil {
		userToken = loadedToken.(Token)
		ok = isValidToken(userToken)
	}
	if !ok {
		log.Printf("no valid user token")
		body, err := createToken(userID, adminToken.Value)
		if err != nil {
			c.JSON(500, err.Error())
			return
		}
		wekanCreateToken := WekanCreateToken{}
		err = json.Unmarshal(body, &wekanCreateToken)
		if err != nil {
			c.JSON(500, err.Error())
			return
		}
		userToken.Value = wekanCreateToken.AuthToken
		now := time.Now()
		userToken.CreationDate = now
		tokens.Store(s.username, userToken)
	} else {
		log.Printf("existing user token")
	}
	// data enrichment
	etsData, err := getEtablissementDataFromDb(siret)
	if err != nil {
		log.Println(err.Error())
		return
	}

	wekanConfig.mu.Lock()
	board, ok := wekanConfig.BoardIds[param.BoardID]
	wekanConfig.mu.Unlock()
	if !ok {
		c.JSON(500, "missing board in config file")
		return
	}
	boardID := param.BoardID
	listID := board.Lists[0]
	// fields formatting
	departement := etsData.Departement
	swimlaneID, ok := board.Swimlanes[departement]
	if !ok {
		c.JSON(500, "not a departement of this region")
		return
	}
	activite := formatActiviteField(etsData.CodeActivite, etsData.LibelleActivite)
	ficheSF := formatFicheSFField(etsData.Siret)
	effectifIndex := getEffectifIndex(etsData.Effectif)
	// new card
	title := etsData.RaisonSociale
	creationData := map[string]interface{}{
		"authorId":   userID,
		"members":    [1]string{userID},
		"title":      title,
		"swimlaneId": swimlaneID,
	}
	// body, err := ioutil.ReadAll(c.Request.Body)
	// if err != nil {
	// 	c.JSON(500, err.Error())
	// 	return
	// }
	// var requestData map[string]string
	// err = json.Unmarshal(body, &requestData)
	// if err != nil {
	// 	c.JSON(500, err.Error())
	// 	return
	// }

	if param.Description != "" {
		creationData["description"] = param.Description
	}
	body, err := createCard(userToken.Value, boardID, listID, creationData)
	if err != nil {
		fmt.Println("we're here")
		c.JSON(500, err.Error())
		return
	}
	wekanCreateCard := WekanCreateCard{}
	err = json.Unmarshal(body, &wekanCreateCard)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}
	cardID := wekanCreateCard.ID
	// card edition
	var customFields = []CustomField{
		{board.CustomFields.SiretField, siret},
		{board.CustomFields.FicheSFField, ficheSF},
	}
	if activite != "" {
		customFields = append(customFields, CustomField{board.CustomFields.ActiviteField, activite})
	} else {
		customFields = append(customFields, CustomField{board.CustomFields.ActiviteField, ""})
	}
	if effectifIndex >= 0 {
		customFields = append(customFields, CustomField{board.CustomFields.EffectifField.EffectifFieldID, board.CustomFields.EffectifField.EffectifFieldItems[effectifIndex]})
	} else {
		customFields = append(customFields, CustomField{board.CustomFields.EffectifField.EffectifFieldID, ""})
	}
	customFields = append(customFields, CustomField{board.CustomFields.ContactField, ""})
	now := time.Now()
	var editionData = map[string]interface{}{
		"customFields": customFields,
		"startAt":      now,
	}
	_, err = editCard(adminToken.Value, boardID, listID, cardID, editionData)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}
}

func adminLogin(username string, password string) ([]byte, error) {
	wekanURL := viper.GetString("wekanURL")
	data := url.Values{}
	data.Set("username", username)
	data.Set("password", password)
	req, err := http.NewRequest("POST", wekanURL+"users/login", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	client := &http.Client{}
	resp, err := client.Do(req)
	if resp.StatusCode != http.StatusOK {
		err = errors.New("adminLogin: unexpected wekan API response")
	}
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func createToken(userID string, adminToken string) ([]byte, error) {
	wekanURL := viper.GetString("wekanURL")
	req, err := http.NewRequest("POST", wekanURL+"api/createtoken/"+userID, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+adminToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if resp.StatusCode != http.StatusOK {
		err = errors.New("createToken: unexpected wekan API response")
	}
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func createCard(userToken string, boardID string, listID string, creationData map[string]interface{}) ([]byte, error) {
	wekanURL := viper.GetString("wekanURL")
	json, err := json.Marshal(creationData)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", wekanURL+"api/boards/"+boardID+"/lists/"+listID+"/cards", bytes.NewBuffer(json))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+userToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		err = fmt.Errorf("createCard: unexpected wekan API response: %d\nBody: %s", resp.StatusCode, body)
	}
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func editCard(userToken string, boardID string, listID string, cardID string, editionData map[string]interface{}) ([]byte, error) {
	wekanURL := viper.GetString("wekanURL")
	json, err := json.Marshal(editionData)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("PUT", wekanURL+"api/boards/"+boardID+"/lists/"+listID+"/cards/"+cardID, bytes.NewBuffer(json))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+userToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		err = fmt.Errorf("editCard: unexpected wekan API response: %d\nBody: %s", resp.StatusCode, body)
	}
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func commentCard(userToken string, boardID string, cardID string, commentData map[string]string) ([]byte, error) {
	wekanURL := viper.GetString("wekanURL")
	json, err := json.Marshal(commentData)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", wekanURL+"api/boards/"+boardID+"/cards/"+cardID+"/comments", bytes.NewBuffer(json))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+userToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		err = fmt.Errorf("commentCard: unexpected wekan API response: %d\nBody: %s", resp.StatusCode, body)
	}
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func isValidToken(token Token) bool {
	duration := time.Since(token.CreationDate)
	hours := duration.Hours()
	days := hours / 24
	return days < float64(viper.GetInt("wekanTokenDays"))
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

func formatActiviteField(codeActivite string, libelleActivite string) string {
	var activite string
	if len(libelleActivite) > 0 && len(codeActivite) > 0 {
		activite = libelleActivite + " (" + codeActivite + ")"
	}
	return activite
}

func getEffectifIndex(effectif int) int {
	var effectifIndex = -1
	effectifClass := [4]int{10, 20, 50, 100}
	for i := len(effectifClass) - 1; i >= 0; i-- {
		if effectif >= effectifClass[i] {
			effectifIndex = i
			break
		}
	}
	return effectifIndex
}

func formatFicheSFField(siret string) string {
	return viper.GetString("webBaseURL") + "ets/" + siret
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

func (c WekanCard) Board() string {
	return wekanConfig.BoardForId(c.BoardId)
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
