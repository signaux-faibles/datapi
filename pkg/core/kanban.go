package core

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/signaux-faibles/libwekan"
	"net/http"
	"time"
)

// KanbanService service définissant les méthodes de Kanban nécessaires dans Datapi
// TODO il  faudrait changer les types `libwekan` en des types `datapi`
type KanbanService interface {
	LoadConfigForUser(username libwekan.Username) KanbanConfig
	SelectCardsFromSiret(ctx context.Context, siret string, username libwekan.Username) ([]KanbanCard, error)
	ExportCardsFromSiret(ctx context.Context, siret string, username libwekan.Username) ([]KanbanCard, error)
	SelectFollowsForUser(ctx context.Context, params KanbanSelectCardsForUserParams, db *pgxpool.Pool, roles []string) (Summaries, error)
	ExportFollowsForUser(ctx context.Context, params KanbanSelectCardsForUserParams, db *pgxpool.Pool, roles []string) (KanbanExports, error)
	SelectKanbanExportsWithSiret(ctx context.Context, siret string, username string, db *pgxpool.Pool, roles []string) (KanbanExports, error)
	GetUser(username libwekan.Username) (libwekan.User, bool)
	CreateCard(ctx context.Context, params KanbanNewCardParams, username libwekan.Username, db *pgxpool.Pool) error
	PartCard(ctx context.Context, cardID libwekan.CardID, userID libwekan.UserID) error
	UnarchiveCard(ctx context.Context, cardID libwekan.CardID, username libwekan.Username) error
}

type KanbanUsers map[libwekan.UserID]KanbanUser

type KanbanUser struct {
	Username libwekan.Username `json:"username"`
	Active   bool              `json:"active"`
}

type KanbanLists map[libwekan.ListID]KanbanList
type KanbanList struct {
	Title string  `json:"title"`
	Sort  float64 `json:"sort"`
}

type KanbanSwimlanes map[libwekan.SwimlaneID]KanbanSwimlane
type KanbanSwimlane struct {
	Title string  `json:"title"`
	Sort  float64 `json:"sort"`
}

type KanbanBoardMembers map[libwekan.UserID]KanbanBoardMember
type KanbanBoardMember struct {
	Username libwekan.Username `json:"username"`
	Active   bool              `json:"active"`
}

type KanbanBoardLabels map[libwekan.BoardLabelID]KanbanBoardLabel
type KanbanBoardLabel struct {
	Color string                  `json:"color"`
	Name  libwekan.BoardLabelName `json:"name"`
}

type KanbanBoards map[libwekan.BoardID]KanbanBoard
type KanbanBoard struct {
	Title     libwekan.BoardTitle `json:"title"`
	Slug      libwekan.BoardSlug  `json:"slug"`
	Lists     KanbanLists         `json:"lists"`
	Swimlanes KanbanSwimlanes     `json:"swimlanes"`
	Labels    KanbanBoardLabels   `json:"labels"`
	Members   KanbanBoardMembers  `json:"members"`
}

type KanbanBoardSwimlane struct {
	BoardID    libwekan.BoardID    `json:"boardID"`
	SwimlaneID libwekan.SwimlaneID `json:"swimlaneID"`
}

type KanbanConfig struct {
	Departements map[CodeDepartement][]KanbanBoardSwimlane `json:"departements"`
	Boards       KanbanBoards                              `json:"boards"`
	Users        KanbanUsers                               `json:"users"`
	UserID       libwekan.UserID                           `json:"userID"`
}

type KanbanCard struct {
	ID                libwekan.CardID         `json:"id,omitempty"`
	ListID            libwekan.ListID         `json:"listID,omitempty"`
	ListTitle         string                  `json:"listTitle,omitempty"`
	Archived          bool                    `json:"archived"`
	BoardID           libwekan.BoardID        `json:"boardID,omitempty"`
	BoardTitle        libwekan.BoardTitle     `json:"boardTitle,omitempty"`
	SwimlaneID        libwekan.SwimlaneID     `json:"swimlaneID,omitempty"`
	URL               string                  `json:"url,omitempty"`
	Description       string                  `json:"description,omitempty"`
	AssigneeIDs       []libwekan.UserID       `json:"assigneesID,omitempty"`
	MemberIDs         []libwekan.UserID       `json:"memberIDs,omitempty"`
	CreatorID         libwekan.UserID         `json:"creatorID,omitempty"`
	Creator           libwekan.Username       `json:"creator,omitempty"`
	LastActivity      time.Time               `json:"lastActivity,omitempty"`
	StartAt           time.Time               `json:"startAt,omitempty"`
	EndAt             *time.Time              `json:"endAt,omitempty"`
	LabelIDs          []libwekan.BoardLabelID `json:"labelIds,omitempty"`
	UserIsBoardMember bool                    `json:"userIsBoardMember"`
}

type KanbanSelectCardsForUserParams struct {
	User          libwekan.User             `json:"-"`
	Type          string                    `json:"type"`
	Zone          []string                  `json:"zone"`
	BoardIDs      []libwekan.BoardID        `json:"boardIDs"`
	Labels        []libwekan.BoardLabelName `json:"labels"`
	Since         *time.Time                `json:"since"`
	Lists         []string                  `json:"lists"`
	LabelMode     string                    `json:"labelMode"`
	RaisonSociale *string                   `json:"raisonSociale"`
}

type KanbanFollows struct {
	Count   int       `json:"count"`
	Follows []Summary `json:"summaries"`
}

type KanbanNewCardParams struct {
	SwimlaneID  libwekan.SwimlaneID       `json:"swimlaneID"`
	Description string                    `json:"description"`
	Labels      []libwekan.BoardLabelName `json:"labels"`
	Siret       Siret                     `json:"siret"`
}

// ExportHeader contient l'entête d'un document d'export
type ExportHeader struct {
	Auteur string
	Date   time.Time
}

// KanbanExports array of WekanExport
type KanbanExports []KanbanExport

// KanbanExport fournit les champs nécessaires pour l'export Wekan
type KanbanExport struct {
	RaisonSociale              string    `json:"raison_sociale"`
	Siret                      string    `json:"siret"`
	TypeEtablissement          string    `json:"type_etablissement"`
	TeteDeGroupe               string    `json:"tete_de_groupe"`
	Departement                string    `json:"departement"`
	Commune                    string    `json:"commune"`
	TerritoireIndustrie        string    `json:"territoire_industrie"`
	SecteurActivite            string    `json:"secteur_activite"`
	Activite                   string    `json:"activite"`
	SecteursCovid              string    `json:"secteurs_covid"`
	StatutJuridique            string    `json:"statut_juridique"`
	DateOuvertureEtablissement string    `json:"date_ouverture_etablissement"`
	DateCreationEntreprise     string    `json:"date_creation_entreprise"`
	Effectif                   string    `json:"effectif"`
	ActivitePartielle          string    `json:"activite_partielle"`
	DetteSociale               string    `json:"dette_sociale"`
	PartSalariale              string    `json:"part_salariale"`
	AnneeExercice              string    `json:"annee_exercice"`
	ChiffreAffaire             string    `json:"ca"`
	ExcedentBrutExploitation   string    `json:"ebe"`
	ResultatExploitation       string    `json:"rex"`
	ProcedureCollective        string    `json:"procol"`
	DetectionSF                string    `json:"detection_sf"`
	DateDebutSuivi             string    `json:"date_debut_suivi"`
	DateFinSuivi               string    `json:"date_fin_suivi"`
	DescriptionWekan           string    `json:"description_wekan"`
	Labels                     []string  `json:"labels"`
	Board                      string    `json:"-"`
	LastActivity               time.Time `json:"lastActivity"`
	Archived                   bool      `json:"-"`
}

func (k KanbanDBExport) GetSiret() string {
	return k.Siret
}

type KanbanDBExport struct {
	Siret                             string    `json:"siret"`
	RaisonSociale                     string    `json:"raisonSociale"`
	CodeDepartement                   string    `json:"codeDepartement"`
	LibelleDepartement                string    `json:"libelleDepartement"`
	Commune                           string    `json:"commune"`
	CodeTerritoireIndustrie           string    `json:"codeTerritoireIndustrie"`
	LibelleTerritoireIndustrie        string    `json:"libelleTerritoireIndustrie"`
	Siege                             bool      `json:"siege"`
	TeteDeGroupe                      string    `json:"teteDeGroupe"`
	CodeActivite                      string    `json:"codeActivite"`
	LibelleActivite                   string    `json:"libelleActivite"`
	SecteurActivite                   string    `json:"secteurActivite"`
	StatutJuridiqueN1                 string    `json:"statutJuridiqueN1"`
	StatutJuridiqueN2                 string    `json:"statutJuridiqueN2"`
	StatutJuridiqueN3                 string    `json:"statutJuridiqueN3"`
	DateOuvertureEtablissement        time.Time `json:"dateOuvertureEtablissement"`
	DateCreationEntreprise            time.Time `json:"dateCreationEntreprise"`
	DernierEffectif                   int       `json:"dernierEffecti"`
	DateDernierEffectif               time.Time `json:"dateDernierEffectif"`
	DateArreteBilan                   time.Time `json:"dateArreteBilan"`
	ExerciceDiane                     int       `json:"exerciceDiane"`
	ChiffreAffaire                    float64   `json:"chiffreAffaire"`
	ChiffreAffairePrecedent           float64   `json:"chiffreAffairePrecedent"`
	VariationCA                       float64   `json:"variationCA"`
	ResultatExploitation              float64   `json:"resultatExploitation"`
	ResultatExploitationPrecedent     float64   `json:"resultatExploitationPrecedent"`
	ExcedentBrutExploitation          float64   `json:"excedentBrutExploitation"`
	ExcedentBrutExploitationPrecedent float64   `json:"excedentBrutExploitationPrecedent"`
	DerniereListe                     string    `json:"derniereListe"`
	DerniereAlerte                    string    `json:"derniereAlerte"`
	ActivitePartielle                 bool      `json:"activitePartielle"`
	DetteSociale                      bool      `json:"detteSociale"`
	PartSalariale                     bool      `json:"partSalariale"`
	DateUrssaf                        time.Time `json:"dateUrssaf"`
	ProcedureCollective               string    `json:"procedureCollective"`
	DateProcedureCollective           time.Time `json:"dateProcedureCollective"`
	DateDebutSuivi                    time.Time `json:"dateDebutSuivi"`
	CommentSuivi                      string    `json:"commentSuivi"`
	InZone                            bool      `json:"inZone"`
}

type KanbanDBExports []*KanbanDBExport

func (exports KanbanDBExports) NewDbExport() (KanbanDBExports, []interface{}) {
	var e KanbanDBExport

	exports = append(exports, &e)
	t := []interface{}{
		&e.Siret,
		&e.RaisonSociale,
		&e.CodeDepartement,
		&e.LibelleDepartement,
		&e.Commune,
		&e.CodeTerritoireIndustrie,
		&e.LibelleTerritoireIndustrie,
		&e.Siege,
		&e.TeteDeGroupe,
		&e.CodeActivite,
		&e.LibelleActivite,
		&e.SecteurActivite,
		&e.StatutJuridiqueN1,
		&e.StatutJuridiqueN2,
		&e.StatutJuridiqueN3,
		&e.DateOuvertureEtablissement,
		&e.DateCreationEntreprise,
		&e.DernierEffectif,
		&e.DateDernierEffectif,
		&e.DateArreteBilan,
		&e.ExerciceDiane,
		&e.ChiffreAffaire,
		&e.ChiffreAffairePrecedent,
		&e.VariationCA,
		&e.ResultatExploitation,
		&e.ResultatExploitationPrecedent,
		&e.ExcedentBrutExploitation,
		&e.ExcedentBrutExploitationPrecedent,
		&e.DerniereListe,
		&e.DerniereAlerte,
		&e.ActivitePartielle,
		&e.DetteSociale,
		&e.PartSalariale,
		&e.DateUrssaf,
		&e.ProcedureCollective,
		&e.DateProcedureCollective,
		&e.DateDebutSuivi,
		&e.CommentSuivi,
		&e.InZone,
	}
	return exports, t
}

type KanbanCardAndComments struct {
	Card     libwekan.Card      `json:"card"`
	Comments []libwekan.Comment `json:"comments"`
}

func kanbanUnarchiveCardHandler(c *gin.Context) {
	var s session
	s.Bind(c)

	cardID := libwekan.CardID(c.Param("cardID"))

	err := kanban.UnarchiveCard(c, cardID, libwekan.Username(s.Username))
	if errors.Is(err, UnknownCardError{}) {
		c.JSON(http.StatusNotFound, err.Error())
	}
	if errors.Is(err, ForbiddenError{}) {
		c.JSON(http.StatusForbidden, err.Error())
	}
	if errors.Is(err, UnknownBoardError{}) {
		c.JSON(http.StatusNotFound, err.Error())
	}
	if errors.Is(err, libwekan.NothingDoneError{}) {
		c.JSON(http.StatusNoContent, err.Error())
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
	}
	c.JSON(200, "traitement effectué")
}