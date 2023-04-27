package wekan

import (
	"context"
	"github.com/signaux-faibles/datapi/src/core"
	"github.com/signaux-faibles/libwekan"
	"github.com/spf13/viper"
)

func (s wekanService) CreateCard(ctx context.Context, params core.KanbanNewCardParams, username libwekan.Username) error {
	user, ok := GetUser(username)
	if !ok {
		return core.ForbiddenError{"l'utilisateur n'est pas enregistré dans wekan"}
	}
	board, swimlane, err := getBoardWithSwimlaneID(params.SwimlaneID)
	if err != nil {
		return err
	}
	list, err := getListWithBoardID(board.Board.ID, "A définir")
	if err != nil {
		return err
	}
	etablissement, err := core.GetEtablissementDataFromDb(params.Siret)
	if err != nil {
		return err
	}
	card, err := buildCard(board, list.ID, swimlane.ID, params.Description, params.Siret, user, etablissement)
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
) (libwekan.Card, error) {
	activiteField := buildActiviteField(configBoard, etablissement.CodeActivite, etablissement.LibelleActivite)
	effectifField := buildEffectifField(configBoard, etablissement.Effectif)
	contactField := buildContactField(configBoard)
	siretField := buildSiretField(configBoard, siret)
	ficheField := buildFicheField(configBoard, siret)

	card := libwekan.BuildCard(configBoard.Board.ID, listID, swimlaneID, etablissement.RaisonSociale, description, user.ID)
	card.CustomFields = []libwekan.CardCustomField{activiteField, effectifField, contactField, siretField, ficheField}
	card.Members = []libwekan.UserID{user.ID}

	return card, nil
}

func getListWithBoardID(boardID libwekan.BoardID, title string) (libwekan.List, error) {
	wekanConfigMutex.Lock()
	wc := wekanConfig.Copy()
	wekanConfigMutex.Unlock()

	configBoard, ok := wc.Boards[boardID]
	if !ok {
		return libwekan.List{}, core.UnknownBoardError{"id=" + string(boardID)}
	}
	for _, list := range configBoard.Lists {
		if list.Title == title {
			return list, nil
		}
	}
	return libwekan.List{}, core.UnknownListError{"titre=" + title}
}

func getBoardWithSwimlaneID(swimlaneID libwekan.SwimlaneID) (libwekan.ConfigBoard, libwekan.Swimlane, error) {
	wekanConfigMutex.Lock()
	wc := wekanConfig.Copy()
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
		return libwekan.ConfigBoard{}, libwekan.Swimlane{}, core.ForbiddenError{"l'utilisateur n'est pas habilité sur ce tableau"}
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

	id := configBoard.CustomFields.CustomFieldID("Effectif")
	if effectifIndex >= 0 {
		return libwekan.CardCustomField{id, effectifItems[effectifIndex]}
	}
	return libwekan.CardCustomField{id, ""}
}

func buildContactField(configBoard libwekan.ConfigBoard) libwekan.CardCustomField {
	id := configBoard.CustomFields.CustomFieldID("Contact")
	return libwekan.CardCustomField{id, ""}
}

func buildSiretField(configBoard libwekan.ConfigBoard, siret core.Siret) libwekan.CardCustomField {
	id := configBoard.CustomFields.CustomFieldID("SIRET")
	return libwekan.CardCustomField{id, string(siret)}
}

func buildFicheField(configBoard libwekan.ConfigBoard, siret core.Siret) libwekan.CardCustomField {
	id := configBoard.CustomFields.CustomFieldID("Fiche Signaux Faibles")
	return libwekan.CardCustomField{id, viper.GetString("webBaseURL") + "ets/" + string(siret)}
}
