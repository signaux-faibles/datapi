package core

import (
	"context"
	"fmt"
	"github.com/signaux-faibles/datapi/src/db"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"

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

func GetEtablissementDataFromDb(siret Siret) (EtablissementData, error) {
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

// WekanCards est une liste de WekanCard
// TODO: remove type
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

// TODO libwekan
func cardFromSiretPipeline(wcu OldWekanConfig, siret string) bson.A {
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
