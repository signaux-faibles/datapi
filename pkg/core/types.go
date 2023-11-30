// Package core contient le code de base de Datapi. Pour le moment c'est encore le fourre-tout mais un jour ça sera beau !!!
package core

import (
	"context"
	"datapi/pkg/db"
	"regexp"
	"time"
)

// Siret : représente l'indifiant `Siret` d'un établissement
type Siret string
type Siren string

// Siren : interpole la valeur Siren depuis le Siret d'un établissement
func (s Siret) Siren() Siren {
	return Siren(s[0:9])
}

func (s Siret) IsValidSiretString() bool {
	siretRegexp, _ := regexp.Compile("[0-9]{14}")
	return siretRegexp.MatchString(string(s))
}

func (s Siret) Exists(ctx context.Context) (bool, error) {
	conn := db.Get()
	row := conn.QueryRow(ctx, "select distinct true from v_summaries union all select false limit 1")
	var result bool
	err := row.Scan(&result)
	return result, err
}

var validSirenRegexp = regexp.MustCompile("^[0-9]{9}")

func (s Siren) IsValid() bool {
	return validSirenRegexp.MatchString(string(s))
}

// Diane représente les infos Diane sur une entreprise
type Diane struct {
	NumeroSiren          string    `json:"numero_siren,omitempty"`
	ArreteBilan          time.Time `json:"arrete_bilan_diane,omitempty"`
	Exercice             float64   `json:"exercice_diane,omitempty"`
	ChiffreAffaire       *float64  `json:"ca,omitempty"`
	ResultatExploitation *float64  `json:"resultat_expl"`
	BeneficeOuPerte      *float64  `json:"benefice_ou_perte,omitempty"`

	ExcedentBrutDExploitation *float64 `json:"excedent_brut_d_exploitation,omitempty"`
}

//// Bdf représente les infos Bdf sur une entreprise
//type Bdf struct {
//	Siren               string    `json:"siren,omitempty"`
//	Annee               int       `json:"annee_bdf"`
//	ArreteBilan         time.Time `json:"arrete_bilan_bdf"`
//	DelaiFournisseur    float64   `json:"delai_fournisseur"`
//	DetteFiscale        float64   `json:"dette_fiscale"`
//	FinancierCourtTerme float64   `json:"financier_court_terme"`
//	FraisFinancier      float64   `json:"FraisFinancier"`
//	PoidsFrng           float64   `json:"poids_frng"`
//	TauxMarge           float64   `json:"taux_marge"`
//}

// Ellisphere représente les infos Ellisphere sur une entreprise
type Ellisphere struct {
	Siren               string  `json:"-"`
	CodeGroupe          string  `json:"code_groupe,omitempty"`
	SirenGroupe         string  `json:"siren_groupe,omitempty"`
	RefIDGroupe         string  `json:"refid_groupe,omitempty"`
	RaisocGroupe        string  `json:"raison_sociale_groupe,omitempty"`
	AdresseGroupe       string  `json:"adresse_groupe,omitempty"`
	PersonnePouMGroupe  string  `json:"personne_pou_m_groupe,omitempty"`
	NiveauDetention     int     `json:"niveau_detention,omitempty"`
	PartFinanciere      float64 `json:"part_financiere,omitempty"`
	CodeFiliere         string  `json:"code_filiere,omitempty"`
	RefIDFiliere        string  `json:"refid_filiere,omitempty"`
	PersonnePouMFiliere string  `json:"personne_pou_m_filiere,omitempty"`
}

// CodeDepartement correspond à un numéro de département français
type CodeDepartement string

// Region correspond à une région française
type Region string
