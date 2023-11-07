package kanban

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/signaux-faibles/libwekan"
	"github.com/spf13/viper"
	"slices"

	"datapi/pkg/core"
	"datapi/pkg/utils"
)

// CreateCard permet la création d'une carte dans la base de données wekan
func (service wekanService) CreateCard(ctx context.Context, params core.KanbanNewCardParams, username libwekan.Username, db *pgxpool.Pool) error {
	user, ok := GetUser(username)
	if !ok {
		return core.ForbiddenError{Reason: "l'utilisateur n'est pas enregistré dans wekan"}
	}
	board, swimlane, err := getBoardWithSwimlaneID(params.SwimlaneID)
	if err != nil {
		return err
	}
	list, err := getListWithBoardID(board.Board.ID, 0)
	if err != nil {
		return err
	}
	etablissement, err := getEtablissementDataFromDb(ctx, db, params.Siret)
	if err != nil {
		return err
	}

	card, err := buildCard(board, list.ID, swimlane.ID, params.Description, params.Siret, user, etablissement, params.Labels)
	if err != nil {
		return err
	}
	return wekan.InsertCard(ctx, card)
}

func buildCard(
	configBoard libwekan.ConfigBoard,
	listID libwekan.ListID,
	swimlaneID libwekan.SwimlaneID,
	description string,
	siret core.Siret,
	user libwekan.User,
	etablissement core.EtablissementData,
	labels []libwekan.BoardLabelName,
) (libwekan.Card, error) {
	card := libwekan.BuildCard(configBoard.Board.ID, listID, swimlaneID, etablissement.RaisonSociale, description, user.ID)

	// le créateur de la carte est assigné automatiquement
	card.Assignees = []libwekan.UserID{user.ID}

	activiteField := buildActiviteField(configBoard, etablissement.CodeActivite, etablissement.LibelleActivite)
	effectifField := buildEffectifField(configBoard, etablissement.Effectif)
	contactField := buildContactField(configBoard)
	siretField := buildSiretField(configBoard, siret)
	ficheField := buildFicheField(configBoard, siret)
	card.CustomFields = []libwekan.CardCustomField{activiteField, effectifField, contactField, siretField, ficheField}

	labelIDs := utils.Convert(labels, labelNameToIDConvertor(configBoard))
	if labelIDs != nil {
		card.LabelIDs = labelIDs
	}
	return card, nil
}

func labelNameToIDConvertor(board libwekan.ConfigBoard) func(libwekan.BoardLabelName) libwekan.BoardLabelID {
	return func(name libwekan.BoardLabelName) libwekan.BoardLabelID {
		for _, label := range board.Board.Labels {
			if label.Name == name {
				return label.ID
			}
		}
		return ""
	}
}

func getListWithBoardID(boardID libwekan.BoardID, rank int) (libwekan.List, error) {
	wekanConfigMutex.Lock()
	wc := WekanConfig.Copy()
	wekanConfigMutex.Unlock()

	configBoard, ok := wc.Boards[boardID]
	if !ok {
		return libwekan.List{}, core.UnknownBoardError{BoardIdentifier: "id=" + string(boardID)}
	}

	var lists []libwekan.List
	for _, list := range configBoard.Lists {
		lists = append(lists, list)
	}
	slices.SortFunc(lists, func(listA libwekan.List, listB libwekan.List) int {
		if listA.Sort-listB.Sort > 0 {
			return 1
		} else {
			return -1
		}
	})

	if rank >= len(lists) {
		return libwekan.List{}, core.UnknownListError{ListIdentifier: fmt.Sprint("rank=%d", rank)}
	}
	return lists[rank], nil
}

func getBoardWithSwimlaneID(swimlaneID libwekan.SwimlaneID) (libwekan.ConfigBoard, libwekan.Swimlane, error) {
	wekanConfigMutex.Lock()
	wc := WekanConfig.Copy()
	wekanConfigMutex.Unlock()

	var swimlane libwekan.Swimlane
	var configBoard libwekan.ConfigBoard
	var ok bool

	for _, configBoard = range wc.Boards {
		if swimlane, ok = configBoard.Swimlanes[swimlaneID]; ok {
			break
		}
	}
	if !ok {
		return libwekan.ConfigBoard{}, libwekan.Swimlane{}, core.ForbiddenError{Reason: "l'utilisateur n'est pas habilité sur ce tableau"}
	}

	return configBoard, swimlane, nil
}

func buildActiviteField(configBoard libwekan.ConfigBoard, codeActivite string, libelleActivite string) libwekan.CardCustomField {
	var activite string
	if len(libelleActivite) > 0 && len(codeActivite) > 0 {
		activite = libelleActivite + " (" + codeActivite + ")"
	}
	id := configBoard.CustomFields.CustomFieldID("Activité")
	return libwekan.CardCustomField{
		ID:    id,
		Value: activite,
	}
}

func buildEffectifField(configBoard libwekan.ConfigBoard, effectif int) libwekan.CardCustomField {
	var effectifIndex = -1
	effectifClass := [4]int{10, 20, 50, 100}
	effectifItems := [4]string{"10-20", "20-50", "50-100", "100+"}
	for i := len(effectifClass) - 1; i >= 0; i-- {
		if effectif >= effectifClass[i] {
			effectifIndex = i
			break
		}
	}

	customFieldID := configBoard.CustomFields.CustomFieldID("Effectif")
	if effectifIndex >= 0 {
		item := effectifItems[effectifIndex]
		customFieldValueID := customFieldDropdownValueToID(configBoard, item)
		return libwekan.CardCustomField{ID: customFieldID, Value: customFieldValueID}
	}
	return libwekan.CardCustomField{ID: customFieldID}
}

func customFieldDropdownValueToID(configBoard libwekan.ConfigBoard, value string) string {
	id := configBoard.CustomFields.CustomFieldID("Effectif")
	for _, field := range configBoard.CustomFields[id].Settings.DropdownItems {
		if field.Name == value {
			return field.ID
		}
	}
	return ""
}

func buildContactField(configBoard libwekan.ConfigBoard) libwekan.CardCustomField {
	id := configBoard.CustomFields.CustomFieldID("Contact")
	return libwekan.CardCustomField{ID: id}
}

func buildSiretField(configBoard libwekan.ConfigBoard, siret core.Siret) libwekan.CardCustomField {
	id := configBoard.CustomFields.CustomFieldID("SIRET")
	return libwekan.CardCustomField{ID: id, Value: string(siret)}
}

func buildFicheField(configBoard libwekan.ConfigBoard, siret core.Siret) libwekan.CardCustomField {
	id := configBoard.CustomFields.CustomFieldID("Fiche Signaux Faibles")
	return libwekan.CardCustomField{ID: id, Value: viper.GetString("webBaseURL") + "ets/" + string(siret)}
}
