package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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

func buildWekanConfigPipeline() []bson.M {
	matchTableaux := bson.M{
		"$match": bson.M{
			"slug": bson.M{
				"$regex": primitive.Regex{
					Pattern: "^tableau-(c|da)rp.*",
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
			"members": bson.M{
				"$map": bson.M{
					"input": bson.M{
						"$filter": bson.M{
							"input": "$members",
							"as":    "memberf",
							"cond":  "$$memberf.isActive",
						},
					},
					"as": "memberm",
					"in": "$$memberm.userId",
				},
			},
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

// func wekanAllConfigHandler(c *gin.Context) {
// 	roles := scopeFromContext(c)
// 	if contains(roles, "wekan") {
// 		c.JSON(200, wekanConfig)
// 		return
// 	}
// 	c.AbortWithStatus(403)
// }

// func wekanReloadConfigHandler(c *gin.Context) {
// 	var err error
// 	wekanConfig, err = lookupWekanConfig()
// 	if err != nil {
// 		c.JSON(500, fmt.Sprintf("wekanReloadConfig: %s", err))
// 		return
// 	}
// 	c.JSON(200, true)
// }

func wekanConfigHandler(c *gin.Context) {
	roles := scopeFromContext(c)
	if contains(roles, "wekan") {
		username := c.GetString("username")
		c.JSON(200, wekanConfig.forUser(username))
		return
	}
	c.AbortWithStatus(403)
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
	newc.Slugs = make(map[string]*WekanConfigBoard)
	newc.Boards = make(map[string]*WekanConfigBoard)
	newc.mu = &sync.Mutex{}
	wc.mu.Lock()
	newc.Users = wc.Users
	for boardID, wekanConfigBoard := range wc.BoardIds {
		newc.BoardIds[boardID] = wekanConfigBoard
		newc.Boards[wekanConfigBoard.Title] = wekanConfigBoard
		newc.Slugs[wekanConfigBoard.Slug] = wekanConfigBoard
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

func (wc WekanConfig) listForListID(listID string, boardID string) string {
	wc.mu.Lock()
	defer wc.mu.Unlock()
	for _, listDetails := range wc.BoardIds[boardID].ListDetails {
		if listDetails.ID == listID {
			return listDetails.Title
		}
	}
	return ""
}

func (wc WekanConfig) userForUserID(userID string) string {
	wc.mu.Lock()
	defer wc.mu.Unlock()
	for name, id := range wc.Users {
		if userID == id {
			return name
		}
	}
	return ""
}

func (wc WekanConfig) forUser(username string) WekanConfig {
	wc.mu.Lock()
	defer wc.mu.Unlock()

	userID, ok := wc.Users[username]
	if !ok {
		wcu := WekanConfig{}
		wcu.mu = &sync.Mutex{}
		return wcu
	}

	var userWc WekanConfig
	userWc.Boards = make(map[string]*WekanConfigBoard)
	userWc.BoardIds = make(map[string]*WekanConfigBoard)
	userWc.Slugs = make(map[string]*WekanConfigBoard)

	for board, configBoard := range wc.Boards {
		if contains(configBoard.Members, userID) {
			userWc.Boards[board] = configBoard
			userWc.BoardIds[configBoard.BoardID] = configBoard
			userWc.Slugs[configBoard.Slug] = configBoard
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
type wekanListDetails struct {
	ID    string `bson:"_id"`
	Title string `bson:"title"`
}

type wekanLabel struct {
	Color string `bson:"color" json:"color"`
	ID    string `bson:"_id" json:"_id"`
	Name  string `bson:"name" json:"name"`
}

// WekanConfigBoard contient les détails de configuration d'une board Wekan
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
