package wekan

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/signaux-faibles/libwekan"
	"go.mongodb.org/mongo-driver/bson"

	"datapi/pkg/core"
	"datapi/pkg/utils"
)

func textFromComment(comment libwekan.Comment) string {
	return comment.Text
}

func joinCardAndKanbanDBExport(cardAndComments libwekan.CardWithComments, kanbanDBExport core.KanbanDBExport) core.KanbanExport {
	card := cardAndComments.Card
	comments := cardAndComments.Comments

	apartSwitch := map[bool]string{
		true:  "Demande sur les 12 derniers mois",
		false: "Pas de demande récente",
	}
	urssafSwitch := map[bool]string{
		true:  "Hausse sur les 3 derniers mois (%s)",
		false: "Pas de hausse sur les 3 derniers mois (%s)",
	}
	salarialSwitch := map[bool]string{
		true:  "Dette salariale existante (%s)",
		false: "Aucune dette salariale existante (%s)",
	}
	siegeSwitch := map[bool]string{
		true:  "Siège social",
		false: "Établissement secondaire",
	}
	procolSwitch := map[string]string{
		"in_bonis":          "In bonis",
		"liquidation":       "Liquidation judiciaire",
		"redressement":      "Redressement judiciaire",
		"plan_continuation": "Plan de continuation",
		"sauvegarde":        "Sauvegarde",
		"plan_sauvegarde":   "Plan de Sauvegarde",
	}

	we := core.KanbanExport{
		RaisonSociale:              kanbanDBExport.RaisonSociale,
		Siret:                      kanbanDBExport.Siret,
		TypeEtablissement:          siegeSwitch[kanbanDBExport.Siege],
		TeteDeGroupe:               kanbanDBExport.TeteDeGroupe,
		Departement:                fmt.Sprintf("%s (%s)", kanbanDBExport.LibelleDepartement, kanbanDBExport.CodeDepartement),
		Commune:                    kanbanDBExport.Commune,
		TerritoireIndustrie:        kanbanDBExport.LibelleTerritoireIndustrie,
		SecteurActivite:            kanbanDBExport.SecteurActivite,
		Activite:                   fmt.Sprintf("%s (%s)", kanbanDBExport.LibelleActivite, kanbanDBExport.CodeActivite),
		SecteursCovid:              core.SecteurCovid.Get(kanbanDBExport.CodeActivite),
		StatutJuridique:            kanbanDBExport.StatutJuridiqueN2,
		DateOuvertureEtablissement: dateCreation(kanbanDBExport.DateOuvertureEtablissement),
		DateCreationEntreprise:     dateCreation(kanbanDBExport.DateCreationEntreprise),
		Effectif:                   fmt.Sprintf("%d (%s)", kanbanDBExport.DernierEffectif, kanbanDBExport.DateDernierEffectif.Format("01/2006")),
		ActivitePartielle:          apartSwitch[kanbanDBExport.ActivitePartielle],
		DetteSociale:               fmt.Sprintf(urssafSwitch[kanbanDBExport.DetteSociale], dateUrssaf(kanbanDBExport.DateUrssaf)),
		PartSalariale:              fmt.Sprintf(salarialSwitch[kanbanDBExport.PartSalariale], dateUrssaf(kanbanDBExport.DateUrssaf)),
		AnneeExercice:              anneeExercice(kanbanDBExport.ExerciceDiane),
		ChiffreAffaire:             libelleCA(kanbanDBExport.ChiffreAffaire, kanbanDBExport.ChiffreAffairePrecedent, kanbanDBExport.VariationCA),
		ExcedentBrutExploitation:   libelleFin(kanbanDBExport.ExcedentBrutExploitation),
		ResultatExploitation:       libelleFin(kanbanDBExport.ResultatExploitation),
		ProcedureCollective:        procolSwitch[kanbanDBExport.ProcedureCollective],
		DetectionSF:                libelleAlerte(kanbanDBExport.DerniereListe, kanbanDBExport.DerniereAlerte),
		Archived:                   card.Archived,
	}

	we.DateDebutSuivi = dateUrssaf(card.StartAt)
	commentTexts := utils.Convert(comments, textFromComment)

	we.DescriptionWekan = strings.TrimSuffix(card.Description+"\n\n"+strings.ReplaceAll(strings.Join(commentTexts, "\n\n"), "#export", ""), "\n")

	board := WekanConfig.Boards[card.BoardID].Board
	we.Labels = utils.Convert(card.LabelIDs, labelIDToLabelName(board))

	if card.EndAt != nil {
		we.DateFinSuivi = dateUrssaf(*card.EndAt)
	}
	we.Board = string(WekanConfig.Boards[card.BoardID].Board.Title)

	we.LastActivity = card.DateLastActivity

	return we
}

func labelIDToLabelName(board libwekan.Board) func(libwekan.BoardLabelID) string {
	return func(labelID libwekan.BoardLabelID) string {
		for _, label := range board.Labels {
			if label.ID == labelID {
				return string(label.Name)
			}
		}
		return ""
	}
}

func anneeExercice(exercice int) string {
	if exercice == 0 {
		return ""
	}
	return fmt.Sprintf("%d", exercice)
}

func dateUrssaf(dt time.Time) string {
	if dt.IsZero() || dt.Format("02/01/2006") == "01/01/1900" || dt.Format("02/01/2006") == "01/01/0001" {
		return "n/c"
	}
	return dt.Format("01/2006")
}

func dateCreation(dt time.Time) string {
	if dt.IsZero() || dt.Format("02/01/2006") == "01/01/1900" || dt.Format("02/01/2006") == "01/01/0001" {
		return "n/c"
	}
	return dt.Format("02/01/2006")
}

func libelleAlerte(liste string, alerte string) string {
	if alerte == "Alerte seuil F1" {
		return fmt.Sprintf("Risque élevé (%s)", liste)
	}
	if alerte == "Alerte seuil F2" {
		return fmt.Sprintf("Risque modéré (%s)", liste)
	}
	if alerte == "Pas d'alerte" {
		return fmt.Sprintf("Pas de risque (%s)", liste)
	}
	return "Hors périmètre"
}

func libelleFin(val float64) string {
	if val == 0 {
		return "n/c"
	}
	return fmt.Sprintf("%.0f k€", val)
}

func libelleCA(val float64, valPrec float64, variationCA float64) string {
	if val == 0 {
		return "n/c"
	}
	if valPrec == 0 {
		return fmt.Sprintf("%.0f k€", val)
	}
	if variationCA > 1.05 {
		return fmt.Sprintf("%.0f k€ (en hausse)", val)
	} else if variationCA < 0.95 {
		return fmt.Sprintf("%.0f k€ (en baisse)", val)
	}
	return fmt.Sprintf("%.0f k€", val)
}

func selectKanbanDBExportsWithSirets(ctx context.Context,
	sirets []string,
	db *pgxpool.Pool,
	user libwekan.User,
	roles []string,
	zone []string,
	raisonSociale *string) (core.KanbanDBExports, error) {
	rows, err := db.Query(ctx, sqlDbExport, roles, user.Username, sirets, zone, raisonSocialeLike(raisonSociale))
	if err != nil {
		return core.KanbanDBExports{}, err
	}
	defer rows.Close()

	var kanbanDBExports core.KanbanDBExports
	for rows.Next() {
		var s []interface{}
		kanbanDBExports, s = kanbanDBExports.NewDbExport()
		err := rows.Scan(s...)
		if err != nil {
			return core.KanbanDBExports{}, err
		}
	}
	return kanbanDBExports, nil
}

func raisonSocialeLike(raisonSociale *string) *string {
	if raisonSociale == nil {
		return nil
	}
	like := "%" + *raisonSociale + "%"
	return &like
}
func selectKanbanDBExportsWithoutCard(
	ctx context.Context,
	sirets []string,
	db *pgxpool.Pool,
	user libwekan.User,
	roles []string,
	zone []string,
	raisonSociale *string,
) (core.KanbanDBExports, error) {
	rows, err := db.Query(ctx, sqlDbExportWithoutCards, roles, user.Username, sirets, zone, raisonSocialeLike(raisonSociale))
	if err != nil {
		return core.KanbanDBExports{}, err
	}
	defer rows.Close()

	var kanbanDBExports core.KanbanDBExports
	for rows.Next() {
		var s []interface{}
		kanbanDBExports, s = kanbanDBExports.NewDbExport()
		err := rows.Scan(s...)
		if err != nil {
			return core.KanbanDBExports{}, err
		}
	}
	return kanbanDBExports, nil
}

func buildCardToCardAndCommentsPipeline() libwekan.Pipeline {
	return libwekan.Pipeline{bson.M{
		"$project": bson.M{
			"card": "$$ROOT",
		},
	}, bson.M{
		"$lookup": bson.M{
			"from":         "cards_comments",
			"as":           "comments",
			"localField":   "card._id",
			"foreignField": "cardId",
		},
	}}
}

func (service wekanService) ExportFollowsForUser(ctx context.Context, params core.KanbanSelectCardsForUserParams, db *pgxpool.Pool, roles []string) (core.KanbanExports, error) {
	wc := WekanConfig.Copy()
	pipeline := buildCardsForUserPipeline(wc, params)
	pipeline.AppendPipeline(buildCardToCardAndCommentsPipeline())

	if params.Type == "no-card" {
		params.Lists = nil
		params.Labels = nil
		params.BoardIDs = nil
		params.Since = nil
	}

	var err error
	var cards []libwekan.CardWithComments
	if utils.Contains(roles, "wekan") {
		cards, err = wekan.SelectCardsWithCommentsFromPipeline(ctx, "boards", pipeline)
	}
	if err != nil {
		return core.KanbanExports{}, err
	}
	sirets := utils.Convert(cards, cardWithCommentsToSiret(wc))

	// my-cards et all-cards utilisent la même méthode
	if utils.Contains([]string{"my-cards", "all-cards"}, params.Type) {
		kanbanDBExports, err := selectKanbanDBExportsWithSirets(ctx, sirets, db, params.User, roles, params.Zone, params.RaisonSociale)
		return joinCardsWithKanbanDBExports(kanbanDBExports, cards), err
	}

	// alors que no-card retourne les suivis datapi sans carte kanban
	kanbanDBExports, err := selectKanbanDBExportsWithoutCard(ctx, sirets, db, params.User, roles, params.Zone, params.RaisonSociale)
	return kanbanDBExportToKanbanExports(kanbanDBExports), err
}

func (service wekanService) SelectKanbanExportsWithSiret(ctx context.Context, siret string, username string, db *pgxpool.Pool, roles []string) (core.KanbanExports, error) {
	user, ok := WekanConfig.GetUserByUsername(libwekan.Username(username))
	if !ok {
		return nil, errors.New("utilisateur non trouvé")
	}
	pipeline := buildSelectCardsFromSiretPipeline(siret)
	pipeline.AppendPipeline(buildCardToCardAndCommentsPipeline())
	cardsWithComments, err := wekan.SelectCardsWithCommentsFromPipeline(ctx, "customFields", pipeline)
	if err != nil {
		return nil, err
	}
	kanbanDBExports, err := selectKanbanDBExportsWithSirets(ctx, []string{siret}, db, user, roles, nil, nil)
	kanbanExports := joinCardsWithKanbanDBExports(kanbanDBExports, cardsWithComments)

	if len(kanbanExports) > 0 {
		return kanbanExports, err
	}
	return kanbanDBExportToKanbanExports(kanbanDBExports), nil
}
