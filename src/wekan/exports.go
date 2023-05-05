package wekan

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/signaux-faibles/datapi/src/core"
	"github.com/signaux-faibles/datapi/src/utils"
	"github.com/signaux-faibles/libwekan"
	"go.mongodb.org/mongo-driver/bson"
	"strings"
	"time"
)

func textFromComment(comment libwekan.Comment) string {
	return comment.Text
}

func GetKanbanDBExportSiret(k core.KanbanDBExport) string {
	return k.Siret
}

func joinCardsAndExports(cardAndComments []libwekan.CardWithComments, dbExports core.KanbanDBExports) core.KanbanExports {
	e := utils.ToMap(dbExports, func(k *core.KanbanDBExport) string {
		return k.Siret
	})

	var exports core.KanbanExports

	for _, cardAndComment := range cardAndComments {
		siret := cardToSiret(wekanConfig)(cardAndComment.Card)
		if e[siret] != nil {
			exports = append(exports, join(cardAndComment, *e[siret]))
		}
	}
	return exports
}

func join(cardAndComments libwekan.CardWithComments, kanbanDBExport core.KanbanDBExport) core.KanbanExport {
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
	}

	we.DateDebutSuivi = dateUrssaf(card.StartAt)
	commentTexts := utils.Convert(comments, textFromComment)

	we.DescriptionWekan = strings.TrimSuffix(card.Description+"\n\n"+strings.ReplaceAll(strings.Join(commentTexts, "\n\n"), "#export", ""), "\n")

	board := wekanConfig.Boards[card.BoardID].Board
	we.Labels = utils.Convert(card.LabelIDs, labelIDToLabelName(board))

	if card.EndAt != nil {
		we.DateFinSuivi = dateUrssaf(*card.EndAt)
	}
	we.Board = string(wekanConfig.Boards[card.BoardID].Board.Title)

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
	zone []string) (core.KanbanDBExports, error) {
	rows, err := db.Query(ctx, sqlDbExport, roles, user.Username, sirets, zone)
	defer rows.Close()
	if err != nil {
		return core.KanbanDBExports{}, err
	}

	var kanbanDBExports core.KanbanDBExports
	for rows.Next() {
		var s []interface{}
		kanbanDBExports, s = kanbanDBExports.NewDbExport()
		err := rows.Scan(s...)
		if err != nil {
			return core.KanbanDBExports{}, err
		}
	}

	return core.KanbanDBExports{}, nil
}

func selectKanbanExportsWithoutCard(
	ctx context.Context,
	sirets []string,
	db *pgxpool.Pool,
	user libwekan.User,
	roles []string,
	zone []string,
) (core.KanbanExports, error) {
	return core.KanbanExports{}, nil
}

func (s wekanService) ExportFollowsForUser(ctx context.Context, params core.KanbanSelectCardsForUserParams, db *pgxpool.Pool, roles []string) (core.KanbanExports, error) {
	wc := wekanConfig.Copy()
	pipeline := buildCardsForUserPipeline(wc, params)
	pipeline.AppendStage(bson.M{
		"$project": bson.M{
			"card": "$$ROOT",
		},
	})

	pipeline.AppendStage(bson.M{
		"$lookup": bson.M{
			"from":         "cards_comments",
			"as":           "comments",
			"localField":   "card._id",
			"foreignField": "cardId",
		},
	})

	fmt.Println(json.MarshalIndent(pipeline, " ", " "))
	cards, err := wekan.SelectCardsWithCommentsFromPipeline(ctx, "boards", append(bson.A{}, pipeline...))
	fmt.Println(len(cards))
	if err != nil {
		return core.KanbanExports{}, err
	}
	sirets := utils.Convert(cards, cardWithCommentsToSiret(wc))

	// my-cards et all-cards utilisent la même méthode
	if utils.Contains([]string{"my-cards", "all-cards"}, params.Type) {
		kanbanDBExports, err := selectKanbanDBExportsWithSirets(ctx, sirets, db, params.User, roles, params.Zone)
		return joinKanbanDBExportsWithCards(kanbanDBExports, cards), err
	}
	// no-cards retourne les suivis datapi sans carte kanban
	kanbanExports, err := selectKanbanExportsWithoutCard(ctx, sirets, db, params.User, roles, params.Zone)
	return kanbanExports, err
}
