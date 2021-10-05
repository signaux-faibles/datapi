package main

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func debugConfig(c *gin.Context) {
	cards, err := selectWekanCards("marie.alloy@direccte.gouv.fr", nil, nil, nil, nil)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(cards.Sirets(), len(cards.Sirets()))
	// selectWekanCards([]string{"christophe.ninucci@direccte.gouv.fr"}, nil, nil, nil, nil)
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
	Rank         float64   `bson:"sort"`
	Description  string    `bson:"description"`
	StartAt      time.Time `bson:"startAt"`
	LabelsIDs    []string  `bson:"labelsIds"`
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
	board, ok := wekanConfig.BoardIds[c.BoardId]
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

func selectWekanCards(username string, boardIds []string, swimlaneIds []string, listIds []string, labelIds []string) (WekanCards, error) {
	userID, ok := wekanConfig.Users[username]
	if !ok {
		return WekanCards{}, fmt.Errorf("unknown wekan user: %s", username)
	}

	filter := bson.M{
		"$expr": bson.M{"$in": bson.A{userID, "$members"}},
		"type":  "cardType-card",
	}

	if len(boardIds) > 0 {
		var ids []string
		for _, id := range boardIds {
			if _, ok := wekanConfig.BoardIds[id]; ok {
				ids = append(ids, id)
			}
		}
		filter["boardId"] = bson.M{"$in": ids}
	} else {
		var ids []string
		for boardId, _ := range wekanConfig.BoardIds {
			ids = append(ids, boardId)
		}
		filter["boardId"] = bson.M{"$in": ids}
	}

	if len(swimlaneIds) > 0 {
		filter["swimlaneId"] = bson.M{"$in": swimlaneIds}
	}
	if len(listIds) > 0 {
		filter["listId"] = bson.M{"$in": listIds}
	}

	// sélection des champs à retourner et tri
	projection := bson.M{
		"title":        1,
		"listId":       1,
		"boardId":      1,
		"swimlaneId":   1,
		"customFields": 1,
		"sort":         1,
		"labelIds":     1,
		"startAt":      1,
		"description":  1,
	}
	options := options.Find().SetProjection(projection).SetSort(bson.M{"sort": 1})

	cardsCursor, err := mgoDB.Collection("cards").Find(context.Background(), filter, options)
	if err != nil {
		return WekanCards{}, err
	}
	var wekanCards []WekanCard
	err = cardsCursor.All(context.Background(), &wekanCards)
	if err != nil {
		return WekanCards{}, err
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
		return WekanConfig{}, fmt.Errorf("unexpected config result, len() = %d", len(wekanConfigs))
	}
	config := wekanConfigs[0]
	config.BoardIds = make(map[string]*WekanConfigBoard)
	for _, board := range config.Boards {
		config.BoardIds[board.BoardID] = board
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
					"k":   bson.M{"$substrCP": bson.A{"$title", 0, 2}},
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

	buildBoardConfig := bson.M{
		"$project": bson.M{
			"k": "$slug",
			"v": bson.M{
				"boardId":   "$boardId",
				"slug":      "$slug",
				"title":     "$title",
				"swimlanes": "$swimlanes",
				"customFields": bson.M{
					"siretField":    bson.M{"$arrayElemAt": bson.A{"$siretField.siretField", 0}},
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
			"_id":    0,
			"boards": bson.M{"$arrayToObject": "$boards"},
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
			"_id":    0,
			"boards": 1,
			"users":  bson.M{"$arrayToObject": "$users"},
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
		buildBoardConfig,
		groupByBoard,
		formatBoards,
		lookupUsers,
		formatUsers,
	}
}
