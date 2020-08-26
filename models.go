package main

import (
	"context"
	"database/sql"
	"errors"
	"time"

	pgx "github.com/jackc/pgx/v4"
)

// Entreprise type entreprise pour l'API
type Entreprise struct {
	Siren string `json:"siren"`
}

// Etablissement type établissement pour l'API
type Etablissement struct {
	Siren string `json:"siren"`
	Siret string `json:"siret"`
}

// Follow type follow pour l'API
type Follow struct {
	Siren  string `json:"siren"`
	UserID string `json:"userId"`
}

// Comment commentaire sur une enterprise
type Comment struct {
	Siren   string `json:"siren"`
	UserID  string `json:"userId"`
	Message string `json:"message"`
}

type paramsListeScores struct {
	Departements []string `json:"zone,omitempty"`
	EtatsProcol  []string `json:"procol,omitempty"`
	Activites    []string `json:"activite,omitempty"`
}

// Liste de détection
type Liste struct {
	ID     string            `json:"id"`
	Batch  string            `json:"batch"`
	Algo   string            `json:"algo"`
	Query  paramsListeScores `json:"-"`
	Scores []Score           `json:"scores,omitempty"`
}

// Score d'une liste de détection
type Score struct {
	Siren             string   `json:"siren"`
	Siret             string   `json:"siret"`
	Score             float64  `json:"score"`
	Diff              *float64 `json:"diff"`
	RaisonSociale     *string  `json:"raison_sociale"`
	Commune           *string  `json:"commune"`
	LibelleActivite   *string  `json:"libelle_activite"`
	CodeActivite      *string  `json:"code_activite"`
	Departement       *string  `json:"departement"`
	DernierEffectif   *int     `json:"dernier_effectif"`
	HausseUrssaf      *bool    `json:"urssaf"`
	ActivitePartielle *bool    `json:"activite_partielle"`
	DernierCA         *int     `json:"ca"`
	DernierREXP       *int     `json:"resultat_expl"`
	EtatProcol        *string  `json:"etat_procol"`
}

func findAllEntreprises(db *sql.DB) ([]Entreprise, error) {
	return nil, errors.New("Non implémenté")
}

func (entreprise *Entreprise) findEntrepriseBySiren(db *pgx.Conn) error {
	return errors.New("Non implémenté")
}

func findAllEtablissements(db *sql.DB) ([]Etablissement, error) {
	return nil, errors.New("Non implémenté")
}

func (etablissement *Etablissement) findEtablissementBySiret(db *sql.DB) error {
	return errors.New("Non implémenté")
}

func (entreprise *Entreprise) findEtablissementsBySiren(db *sql.DB) ([]Etablissement, error) {
	return nil, errors.New("Non implémenté")
}

func (follow *Follow) findEntreprisesFollowedByUser(db *sql.DB) ([]Entreprise, error) {
	return nil, errors.New("Non implémenté")
}

func (follow *Follow) createEntrepriseFollow(db *sql.DB) error {
	return errors.New("Non implémenté")
}

func (entreprise *Entreprise) findCommentsBySiren(db *sql.DB) ([]Comment, error) {
	return nil, errors.New("Non implémenté")
}

func (comment *Comment) createEntrepriseComment(db *sql.DB) error {
	return errors.New("Non implémenté")
}

func findAllListes() ([]Liste, error) {
	var listes []Liste
	rows, err := db.Query(context.Background(), "select algo, batch, libelle from liste order by batch desc, algo asc")
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var l Liste
		err := rows.Scan(&l.Algo, &l.Batch, &l.ID)
		if err != nil {
			return nil, err
		}
		listes = append(listes, l)
	}

	return listes, nil
}

func findLastListeScores(db *sql.DB) ([]Score, error) {
	return nil, errors.New("Non implémenté")
}

func (liste *Liste) load() error {
	sqlListe := `select batch, algo from liste where libelle=$1 and version=0`
	row := db.QueryRow(context.Background(), sqlListe, liste.ID)
	batch, algo := "", ""
	err := row.Scan(&batch, &algo)
	if err != nil {
		return err
	}
	liste.Batch = batch
	liste.Algo = algo
	return nil
}

func (liste *Liste) getScores(roles scope) error {
	if liste.Batch == "" {
		err := liste.load()
		if err != nil {
			return err
		}
	}
	sqlScores := `with roles as (select siren, array_agg(distinct departement) as roles
			from etablissement
			where version = 0
			group by siren),
		effectif as (select siret, last(effectif order by periode) as effectif 
			from etablissement_periode_urssaf where effectif is not null and version = 0 and last_periode = true
			group by siret),
		diane as (select siren, last(chiffre_affaire order by arrete_bilan_diane) as chiffre_affaire, 
			last(resultat_expl order by arrete_bilan_diane) as resultat_expl,
			last(arrete_bilan_diane) as arrete_bilan_diane
			from entreprise_diane group by siren),
		procol as (select siret, last(action_procol order by date_effet) as last_procol 
			from etablissement_procol 
			where version = 0 group by siret),
		apdemande as (select distinct siret, true as ap 
			from etablissement_apdemande
			where periode_start + '1 year'::interval > $6 and version = 0),
		urssaf as (select siret, (array_agg(part_patronale + part_salariale order by periode desc))[0:3] as dette
			from etablissement_periode_urssaf 
			group by siret)
		select 
			et.siret,
			et.siren,
			en.raison_sociale, 
			et.commune, 
			d.libelle, 
			s.score,
			s.diff,
			di.chiffre_affaire,
			di.resultat_expl,
			ef.effectif,
			n.libelle,
			et.ape,
			coalesce(ep.last_procol, 'in_bonis') as last_procol,
			ap.ap as activite_partielle,
			case when u.dette[0] > u.dette[1] or u.dette[1] > u.dette[2] then true else false end as hausseUrssaf
		from score s
		inner join roles r on r.roles && $1 and r.siren = s.siren
		inner join etablissement et on et.siret = s.siret and et.version = 0
		inner join entreprise en on en.siren = s.siren and en.version = 0
		inner join departements d on d.code = et.departement
		inner join naf n on n.code = et.ape
		inner join naf n1 on n.id_n1 = n1.id
		left join effectif ef on ef.siret = s.siret
		left join urssaf u on u.siret = s.siret
		left join apdemande ap on ap.siret = s.siret
		left join procol ep on ep.siret = s.siret
		left join diane di on di.siren = s.siren
		where 
			s.libelle_liste = $2
			and (ep.last_procol=any($3) or $3 is null)
			and (et.departement=any($4) or $4 is null)
			and (n1.code=any($5) or $5 is null)`

	rows, err := db.Query(context.Background(), sqlScores, roles.zoneGeo(), liste.ID, liste.Query.EtatsProcol, liste.Query.Departements, liste.Query.Activites, time.Now())
	if err != nil {
		return err
	}

	var scores []Score
	for rows.Next() {
		var score Score
		err := rows.Scan(
			&score.Siret,
			&score.Siren,
			&score.RaisonSociale,
			&score.Commune,
			&score.Departement,
			&score.Score,
			&score.Diff,
			&score.DernierCA,
			&score.DernierREXP,
			&score.DernierEffectif,
			&score.LibelleActivite,
			&score.CodeActivite,
			&score.EtatProcol,
			&score.ActivitePartielle,
			&score.HausseUrssaf,
		)
		if err != nil {
			return err
		}
		scores = append(scores, score)
	}

	liste.Scores = scores
	return nil
}

func batchToDate(batch string) (time.Time, error) {
	if len(batch) < 4 {
		return time.Time{}, errors.New("incorrect batch number")
	}

	return time.Parse("0601", batch[0:4])
}
