package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type wekanListDetails struct {
	ID    string `bson:"_id"`
	Title string `bson:"title"`
}

type wekanLabel struct {
	Color string `bson:"color" json:"color"`
	ID    string `bson:"_id" json:"_id"`
	Name  string `bson:"name" json:"name"`
}

type WekanConfigBoard struct {
	BoardID      string             `bson:"boardId" json:"boardId"`
	Title        string             `bson:"title" json:"title"`
	Slug         string             `bson:"slug" json:"slug"`
	Labels       []wekanLabel       `bson:"labels" json:"labels"`
	Swimlanes    map[string]string  `bson:"swimlanes" json:"swimlanes"`
	Members      []string           `bson:"members" json:"members"`
	Lists        []string           `bson:"lists" json:"lists"`
	ListDetails  []wekanListDetails `bson:"listDetails" json:"listDetails"`
	CustomFields struct {
		SiretField    string `bson:"siretField" json:"siretField"`
		ActiviteField string `bson:"activiteField" json:"activiteField"`
		ContactField  string `bson:"contactField" json:"contactField"`
		EffectifField struct {
			EffectifFieldID    string   `bson:"effectifFieldId" json:"effectifFieldId"`
			EffectifFieldItems []string `bson:"effectifFieldItems" json:"effectifFieldItems"`
		} `bson:"effectifField" json:"effectifField"`
		FicheSFField string `bson:"ficheSFField" json:"ficheSFField"`
	} `bson:"customFields" json:"customFields"`
}

// WekanConfig type pour le fichier de configuration de Wekan
type WekanConfig struct {
	Slugs    map[string]*WekanConfigBoard `bson:"slugs" json:"slugs,omitempty"`
	Boards   map[string]*WekanConfigBoard `bson:"boards" json:"boards"`
	Users    map[string]string            `bson:"users" json:"users,omitempty"`
	BoardIds map[string]*WekanConfigBoard `json:"-"`
	Regions  map[string]string            `json:"regions,omitempty"`
	mu       *sync.Mutex                  `json:"-"`
}

func (wc WekanConfig) copy() WekanConfig {
	var newc WekanConfig
	newc.BoardIds = make(map[string]*WekanConfigBoard)
	newc.mu = &sync.Mutex{}
	wc.mu.Lock()
	for boardID, wekanConfigBoard := range wc.BoardIds {
		newc.BoardIds[boardID] = wekanConfigBoard
	}
	wc.mu.Unlock()
	return newc
}

func (wc WekanConfig) labelForLabelsIDs(labelIDs []string, boardID string) []string {
	var labels []string
	if wc.BoardIds != nil {
		board := wc.BoardIds[boardID]
		for _, l := range board.Labels {
			if contains(labelIDs, l.ID) && l.Name != "" {
				labels = append(labels, l.Name)
			}
		}
		return labels
	}
	return nil
}

func (wc WekanConfig) forUser(username string) WekanConfig {
	wc.mu.Lock()
	defer wc.mu.Unlock()

	userID, ok := wc.Users[username]
	if !ok {
		return WekanConfig{}
	}

	var userWc WekanConfig
	userWc.Boards = make(map[string]*WekanConfigBoard)
	for board, configBoard := range wc.Boards {
		if contains(configBoard.Members, userID) {
			userWc.Boards[board] = configBoard
		}
	}
	userWc.mu = &sync.Mutex{}
	return userWc
}

func (wc WekanConfig) userID(username string) string {
	wc.mu.Lock()
	userID := wc.Users[username]
	wc.mu.Unlock()
	return userID
}

func (wc WekanConfig) boardIdsForUser(username string) []string {
	wc.mu.Lock()
	defer wc.mu.Unlock()

	if userID, ok := wc.Users[username]; ok {
		var boards []string
		for _, board := range wekanConfig.Boards {
			if contains(board.Members, userID) {
				boards = append(boards, board.BoardID)
			}
		}
		return boards
	}
	return nil
}

func (wc WekanConfig) swimlaneIdsForZone(zone []string) []string {
	wc.mu.Lock()
	var swimlaneIds []string
	for _, board := range wc.Boards {
		for k, s := range board.Swimlanes {
			if contains(zone, k) {
				swimlaneIds = append(swimlaneIds, s)
			}
		}
	}
	wc.mu.Unlock()
	return swimlaneIds
}

type labelID struct {
	boardID string
	labelID string
}

func (wc WekanConfig) labelIdsForLabels(labels []string) [][]labelID {
	wc.mu.Lock()
	defer wc.mu.Unlock()
	var labelIDs [][]labelID
	for _, label := range labels {
		var ids = make([]labelID, 0)
		for _, board := range wc.Boards {
			for _, l := range board.Labels {
				if l.Name == label {
					ids = append(ids, labelID{board.BoardID, l.ID})
				}
			}
		}
		labelIDs = append(labelIDs, ids)
	}
	return labelIDs
}

func (wc WekanConfig) listIdsForStatuts(statuts []string) []string {
	wc.mu.Lock()
	defer wc.mu.Unlock()

	var listIds []string
	for _, board := range wc.Boards {
		for _, s := range board.ListDetails {
			if contains(statuts, s.Title) {
				listIds = append(listIds, s.ID)
			}
		}
	}
	return listIds
}

func (wc WekanConfig) boardIds() []string {
	wc.mu.Lock()
	var ids []string
	for _, board := range wc.Boards {
		ids = append(ids, board.BoardID)
	}
	wc.mu.Unlock()
	return ids
}

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

//TODO: factorisation
func wekanGetCardHandler(c *gin.Context) {
	var s session
	s.bind(c)
	userID := wekanConfig.userID(s.username)
	if userID == "" {
		c.JSON(403, "not a wekan user")
		return
	}

	siret := c.Param("siret")

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
	etsData, err := getEtablissementDataFromDb(siret)
	if err != nil {
		log.Println(err.Error())
		return
	}

	wekanConfig.mu.Lock()
	board := wekanConfig.Boards[etsData.Region]
	wekanConfig.mu.Unlock()

	if board == nil {
		board = &WekanConfigBoard{}
	}

	boardID := board.BoardID
	siretField := board.CustomFields.SiretField
	body, err := getCard(userToken.Value, boardID, siretField, siret)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}
	var wekanGetCard WekanGetCard
	err = json.Unmarshal(body, &wekanGetCard)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}
	if len(wekanGetCard) == 0 {
		c.JSON(404, "card not found")
		return
	}
	card := wekanGetCard[0]
	cardID := card.ID
	cardDescription := card.Description
	lists := board.Lists
	listIndex := indexOf(card.ListID, lists)
	wekanURL := viper.GetString("wekanURL")
	cardURL := wekanURL + "b/" + boardID + "/" + board.Slug + "/" + cardID
	isMember := contains(card.Members, userID)
	cardData := map[string]interface{}{
		"cardId":          cardID,
		"listIndex":       listIndex,
		"cardURL":         cardURL,
		"cardDescription": cardDescription,
		"isMember":        isMember,
	}
	c.JSON(200, cardData)
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
	siret := c.Param("siret")
	var s session
	s.bind(c)

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
	board := wekanConfig.Boards[etsData.Region]
	wekanConfig.mu.Unlock()
	if !ok {
		c.JSON(500, "missing board in config file")
		return
	}
	boardID := board.BoardID
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
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}
	var requestData map[string]string
	err = json.Unmarshal(body, &requestData)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}
	description := requestData["description"]
	if description != "" {
		creationData["description"] = description
	}
	body, err = createCard(userToken.Value, boardID, listID, creationData)
	if err != nil {
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

func wekanImportHandler(c *gin.Context) {
	region := c.Query("region")
	if region == "" {
		c.JSON(400, "region query string missing")
		return
	}
	username := c.Query("username")
	if username == "" {
		c.JSON(400, "username query string missing")
		return
	}
	log.Printf("1. csv file parsing")
	importFile := viper.GetString("wekanImportFile")
	csvFile, err := os.Open(importFile)
	if err != nil {
		log.Println(err.Error())
		return
	}
	defer csvFile.Close()
	csvLines, err := csv.NewReader(csvFile).ReadAll()
	if err != nil {
		log.Println(err.Error())
		return
	}
	var tableauCRP []FicheCRP
	var invalidLines []string
	reSiret, _ := regexp.Compile("^[0-9]{14}$")
	for i, line := range csvLines {
		if i > 1 {
			siret := strings.ReplaceAll(line[1], " ", "")
			match := reSiret.MatchString(siret)
			if !match {
				invalidLines = append(invalidLines, "#"+strconv.Itoa(i+1))
			} else {
				commentaire := "**Difficultés rencontrées par l'entreprise**\n" + line[16] + "\n\n**Etat du climat social**\n" + line[19] + "\n\n**Etat du dossier**\n" + line[22] + "\n\n**Actions et échéances à venir**\n" + line[24] + "\n\n**Commentaires sur l'entreprise**\n" + line[25] + "\n\n**Commentaires sur la situation**\n" + line[26] + "\n\n"
				description := line[27]
				ficheCRP := FicheCRP{siret, commentaire, description}
				tableauCRP = append(tableauCRP, ficheCRP)
			}
		}
	}
	log.Printf("%d lines to import", len(tableauCRP))
	if len(invalidLines) > 0 {
		log.Printf("%d invalid lines : %s", len(invalidLines), strings.Join(invalidLines, ", "))
	} else {
		log.Printf("no invalid lines")
	}
	log.Printf("2. config file reading")
	configFile := viper.GetString("wekanConfigFile")
	fileContent, err := ioutil.ReadFile(configFile)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}
	var config WekanConfig
	err = json.Unmarshal(fileContent, &config)
	if err != nil {
		c.AbortWithError(500, err)
	}
	board, ok := config.Boards[region]
	if !ok {
		c.JSON(500, "missing board in config file")
		return
	}
	boardID := board.BoardID
	listID := board.Lists[0]
	userID, ok := config.Users[username]
	if !ok {
		c.JSON(500, "not a wekan user")
		return
	}
	log.Printf("3. admin login")
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
	log.Printf("4. user token creation")
	var userToken Token
	loadedToken, ok = tokens.Load(username)
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
		tokens.Store(username, userToken)
	} else {
		log.Printf("existing user token")
	}
	log.Printf("5. wekan cards creation")
	for _, ficheCRP := range tableauCRP {
		siret := ficheCRP.Siret
		log.Printf("ets %s to import...", siret)
		// data enrichment
		etsData, err := getEtablissementDataFromDb(siret)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		// fields formatting
		if siret != etsData.Siret {
			log.Printf("unknown ets")
			continue
		}
		if region != etsData.Region {
			log.Printf("incorrect region")
			continue
		}
		departement := etsData.Departement
		if departement == "" {
			log.Printf("unknown departement")
			continue
		}
		swimlaneID, ok := board.Swimlanes[departement]
		if !ok {
			log.Printf("not a departement of this region")
			continue
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
		description := ficheCRP.Description
		if description != "" {
			creationData["description"] = description
		}
		body, err := createCard(userToken.Value, boardID, listID, creationData)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		wekanCreateCard := WekanCreateCard{}
		err = json.Unmarshal(body, &wekanCreateCard)
		if err != nil {
			log.Println(err.Error())
			continue
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
			log.Printf("unknown activite")
			customFields = append(customFields, CustomField{board.CustomFields.ActiviteField, ""})
		}
		if effectifIndex >= 0 {
			customFields = append(customFields, CustomField{board.CustomFields.EffectifField.EffectifFieldID, board.CustomFields.EffectifField.EffectifFieldItems[effectifIndex]})
		} else {
			log.Printf("unknown effectif class")
			customFields = append(customFields, CustomField{board.CustomFields.EffectifField.EffectifFieldID, ""})
		}
		customFields = append(customFields, CustomField{board.CustomFields.ContactField, ""})
		now := time.Now()
		var editionData = map[string]interface{}{
			"customFields": customFields,
			"startAt":      now,
		}
		_, err = editCard(userToken.Value, boardID, listID, cardID, editionData)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		// card comment
		comment := ficheCRP.Commentaire
		commentData := map[string]string{
			"authorId": userID,
			"comment":  comment,
		}
		_, err = commentCard(userToken.Value, boardID, cardID, commentData)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		log.Printf("ets %s imported", siret)
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

func getCard(userToken string, boardID string, siretField string, siret string) ([]byte, error) {
	wekanURL := viper.GetString("wekanURL")
	req, err := http.NewRequest("GET", wekanURL+"api/boards/"+boardID+"/cardsByCustomField/"+siretField+"/"+siret, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+userToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		err = fmt.Errorf("getCard: unexpected wekan API response: %d\nBody: %s", resp.StatusCode, body)
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

func indexOf(element string, array []string) int {
	for i, v := range array {
		if element == v {
			return i
		}
	}
	return -1
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

func wekanConfigHandler(c *gin.Context) {
	roles := scopeFromContext(c)
	if contains(roles, "wekan") {
		username := c.GetString("username")
		c.JSON(200, wekanConfig.forUser(username))
		return
	}
	c.AbortWithStatus(403)
}

func wekanReloadConfigHandler(c *gin.Context) {
	var err error
	wekanConfig, err = lookupWekanConfig()
	if err != nil {
		c.JSON(500, fmt.Sprintf("wekanReloadConfig: %s", err))
		return
	}
	c.JSON(200, true)
}

// WekanCards est une liste de WekanCard
type WekanCards []WekanCard

// WekanCard embarque la sélection de champs Wekan nécessaires au traitement d'export
type WekanCard struct {
	ID           string    `bson:"_id"`
	Title        string    `bson:"title"`
	ListID       string    `bson:"listId"`
	BoardId      string    `bson:"boardId"`
	SwimlaneID   string    `bson:"swimlaneId"`
	Members      []string  `bson:"members"`
	Rank         float64   `bson:"sort"`
	Description  string    `bson:"description"`
	StartAt      time.Time `bson:"startAt"`
	LabelIds     []string  `bson:"labelIds"`
	CustomFields []struct {
		ID    string `bson:"_id"`
		Value string `bson:"value"`
	}
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

func selectWekanCardFromSiret(username string, siret string) (*WekanCard, error) {
	wcu := wekanConfig.forUser(username)
	userID := wekanConfig.userID(username)
	if userID == "" {
		return nil, fmt.Errorf("selectWekanCardFromSiret() -> utilisateur wekan inconnu: %s", username)
	}

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
	fmt.Println(queryOr)
	query := bson.M{
		"$or": queryOr,
	}

	projection := bson.M{
		"title":        1,
		"listId":       1,
		"boardId":      1,
		"members":      1,
		"swimlaneId":   1,
		"customFields": 1,
		"sort":         1,
		"labelIds":     1,
		"startAt":      1,
		"description":  1,
	}

	var wekanCard WekanCard
	options := options.FindOne().SetProjection(projection).SetSort(bson.M{"sort": 1})
	err := mgoDB.Collection("cards").FindOne(context.Background(), query, options).Decode(&wekanCard)
	if err != nil {
		return nil, nil
	}

	return &wekanCard, nil
}

func selectWekanCards(username *string, boardIds []string, swimlaneIds []string, listIds []string, labelIds [][]labelID) ([]*WekanCard, error) {
	query := bson.M{
		"type":     "cardType-card",
		"archived": false,
	}

	if username != nil {
		userID := wekanConfig.userID(*username)
		if userID == "" {
			return nil, fmt.Errorf("selectWekanCards() -> utilisateur wekan inconnu: %s", *username)
		}
		query["$expr"] = bson.M{"$in": bson.A{userID, "$members"}}
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

	// sélection des champs à retourner et tri
	projection := bson.M{
		"title":        1,
		"listId":       1,
		"boardId":      1,
		"members":      1,
		"swimlaneId":   1,
		"customFields": 1,
		"sort":         1,
		"labelIds":     1,
		"startAt":      1,
		"description":  1,
	}
	options := options.Find().SetProjection(projection).SetSort(bson.M{"sort": 1})

	cardsCursor, err := mgoDB.Collection("cards").Find(context.Background(), query, options)
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

func lookupWekanConfig() (WekanConfig, error) {
	cur, err := mgoDB.Collection("boards").Aggregate(context.Background(), buildWekanConfigPipeline())
	if err != nil {
		return WekanConfig{}, err
	}
	var wekanConfigs []WekanConfig
	err = cur.All(context.Background(), &wekanConfigs)
	if err != nil {
		return WekanConfig{}, err
	}
	if len(wekanConfigs) != 1 {
		return WekanConfig{}, nil
	}
	config := wekanConfigs[0]
	config.Boards = make(map[string]*WekanConfigBoard)
	config.BoardIds = make(map[string]*WekanConfigBoard)
	for _, board := range config.Slugs {
		config.BoardIds[board.BoardID] = board
		if region, ok := viper.GetStringMapString("wekanRegions")[board.Slug]; ok {
			config.Boards[region] = board
		}
	}
	config.Regions = make(map[string]string)
	for r, s := range viper.GetStringMapString("wekanRegions") {
		config.Regions[s] = r
	}
	return config, nil
}

func buildWekanConfigPipeline() []bson.M {
	matchTableaux := bson.M{
		"$match": bson.M{
			"slug": bson.M{
				"$regex": primitive.Regex{
					Pattern: "^tableau-crp.*",
					Options: "i",
				}}}}

	lookupSwimlanes := bson.M{
		"$lookup": bson.M{
			"from": "swimlanes",
			"let":  bson.M{"board": "$_id"},
			"pipeline": []bson.M{
				{"$match": bson.M{
					"$expr": bson.M{
						"$eq": bson.A{"$boardId", "$$board"},
					}}},
				{"$project": bson.M{
					"_id": 0,
					"v":   "$_id",
					"k":   bson.M{"$arrayElemAt": bson.A{bson.M{"$split": bson.A{"$title", " "}}, 0}},
				}},
			},
			"as": "swimlanes",
		},
	}

	formatSwimlanes := bson.M{
		"$project": bson.M{
			"_id":     0,
			"boardId": "$_id",
			"title":   "$title",
			"slug":    "$slug",
			"labels":  "$labels",
			"members": "$members.userId",
			"swimlanes": bson.M{
				"$arrayToObject": "$swimlanes",
			}}}

	lookupLists := bson.M{
		"$lookup": bson.M{
			"from": "lists",
			"let":  bson.M{"board": "$boardId"},
			"pipeline": []bson.M{
				{"$match": bson.M{
					"$expr": bson.M{"$eq": bson.A{"$boardId", "$$board"}},
				}},
				{"$sort": bson.M{"order": 1}},
				{"$project": bson.M{
					"_id":   1,
					"title": 1,
				}}},
			"as": "lists",
		}}

	lookupEffectifCustomField := bson.M{
		"$lookup": bson.M{
			"from": "customFields",
			"let":  bson.M{"boardId": "$boardId"},
			"pipeline": []bson.M{
				{"$match": bson.M{"name": bson.M{"$eq": "Effectif"}}},
				{"$unwind": "$boardIds"},
				{"$match": bson.M{"$expr": bson.M{"$eq": bson.A{"$boardIds", "$$boardId"}}}},
				{"$project": bson.M{
					"_id": 0,
					"effectifField": bson.M{
						"effectifFieldId":    "$_id",
						"effectifFieldItems": "$settings.dropdownItems._id",
					}}},
			},
			"as": "effectifField",
		}}

	lookupSiretCustomField := bson.M{
		"$lookup": bson.M{
			"from": "customFields",
			"let":  bson.M{"boardId": "$boardId"},
			"pipeline": []bson.M{
				{"$match": bson.M{"name": bson.M{"$eq": "SIRET"}}},
				{"$unwind": "$boardIds"},
				{"$match": bson.M{"$expr": bson.M{"$eq": bson.A{"$boardIds", "$$boardId"}}}},
				{"$project": bson.M{"_id": 0, "siretField": "$_id"}},
			},
			"as": "siretField",
		}}

	lookupActiviteCustomField := bson.M{
		"$lookup": bson.M{
			"from": "customFields",
			"let":  bson.M{"boardId": "$boardId"},
			"pipeline": []bson.M{
				{"$match": bson.M{"name": bson.M{"$eq": "Activité"}}},
				{"$unwind": "$boardIds"},
				{"$match": bson.M{"$expr": bson.M{"$eq": bson.A{"$boardIds", "$$boardId"}}}},
				{"$project": bson.M{"_id": 0, "activiteField": "$_id"}},
			},
			"as": "activiteField",
		}}

	lookupFicheSFCustomField := bson.M{
		"$lookup": bson.M{
			"from": "customFields",
			"let":  bson.M{"boardId": "$boardId"},
			"pipeline": []bson.M{
				{"$match": bson.M{"name": bson.M{"$eq": "Fiche Signaux Faibles"}}},
				{"$unwind": "$boardIds"},
				{"$match": bson.M{"$expr": bson.M{"$eq": bson.A{"$boardIds", "$$boardId"}}}},
				{"$project": bson.M{"_id": 0, "ficheSFField": "$_id"}},
			},
			"as": "ficheSFField",
		}}

	lookupContactFieldCustomField := bson.M{
		"$lookup": bson.M{
			"from": "customFields",
			"let":  bson.M{"boardId": "$boardId"},
			"pipeline": []bson.M{
				{"$match": bson.M{"name": bson.M{"$eq": "Contact"}}},
				{"$unwind": "$boardIds"},
				{"$match": bson.M{"$expr": bson.M{"$eq": bson.A{"$boardIds", "$$boardId"}}}},
				{"$project": bson.M{"_id": 0, "contactField": "$_id"}},
			},
			"as": "contactField",
		}}

	buildBoardConfig := bson.M{
		"$project": bson.M{
			"k": "$slug",
			"v": bson.M{
				"boardId":   "$boardId",
				"slug":      "$slug",
				"title":     "$title",
				"swimlanes": "$swimlanes",
				"members":   "$members",
				"labels":    "$labels",
				"customFields": bson.M{
					"siretField":    bson.M{"$arrayElemAt": bson.A{"$siretField.siretField", 0}},
					"contactField":  bson.M{"$arrayElemAt": bson.A{"$contactField.contactField", 0}},
					"effectifField": bson.M{"$arrayElemAt": bson.A{"$effectifField.effectifField", 0}},
					"ficheSFField":  bson.M{"$arrayElemAt": bson.A{"$ficheSFField.ficheSFField", 0}},
					"activiteField": bson.M{"$arrayElemAt": bson.A{"$activiteField.activiteField", 0}},
				},
				"lists":       "$lists._id", // pour la retro compatibilité
				"listDetails": "$lists",
			}},
	}

	groupByBoard := bson.M{
		"$group": bson.M{
			"_id":    "",
			"boards": bson.M{"$push": "$$ROOT"},
		},
	}

	formatBoards := bson.M{
		"$project": bson.M{
			"_id":   0,
			"slugs": bson.M{"$arrayToObject": "$boards"},
		}}

	lookupUsers := bson.M{
		"$lookup": bson.M{
			"from": "users",
			"let":  bson.M{},
			"pipeline": []bson.M{
				{"$match": bson.M{"username": bson.M{"$exists": true}}},
				{"$project": bson.M{"_id": 0, "k": "$username", "v": "$_id"}},
			},
			"as": "users",
		}}

	formatUsers := bson.M{
		"$project": bson.M{
			"_id":   0,
			"slugs": 1,
			"users": bson.M{"$arrayToObject": "$users"},
		}}

	return []bson.M{
		matchTableaux,
		lookupSwimlanes,
		formatSwimlanes,
		lookupLists,
		lookupEffectifCustomField,
		lookupSiretCustomField,
		lookupActiviteCustomField,
		lookupFicheSFCustomField,
		lookupContactFieldCustomField,
		buildBoardConfig,
		groupByBoard,
		formatBoards,
		lookupUsers,
		formatUsers,
	}
}

func loadWekanConfig() WekanConfig {
	if wekanConfig.mu == nil {
		wekanConfig.mu = &sync.Mutex{}
	}
	wc, err := lookupWekanConfig()
	if err != nil {
		log.Printf("wekanConfigLoader() -> problem loading config: %s", err.Error())
	} else {
		wekanConfig.mu.Lock()
		wekanConfig.BoardIds = wc.BoardIds
		wekanConfig.Boards = wc.Boards
		wekanConfig.Regions = wc.Regions
		wekanConfig.Slugs = wc.Slugs
		wekanConfig.Users = wc.Users
		wekanConfig.mu.Unlock()
	}
	return wekanConfig
}

func watchWekanConfig(period time.Duration) {
	go func() {
		for {
			loadWekanConfig()
			time.Sleep(period)
		}
	}()
}
