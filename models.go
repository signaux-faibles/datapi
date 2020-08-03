package main

import (
	"database/sql"
	"errors"
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

// Liste de détection
type Liste struct {
	ID     string  `json:"id"`
	Scores []Score `json:"scores"`
}

// Score d'une liste de détection
type Score struct {
	Siren      string  `json:"siren"`
	Siret      string  `json:"siret"`
	ListeID    string  `json:"listeId"`
	ScoreValue float64 `json:"scoreValue"`
}

func findAllEntreprises(db *sql.DB) ([]Entreprise, error) {
	return nil, errors.New("Non implémenté")
}

func (entreprise *Entreprise) findEntrepriseBySiren(db *sql.DB) error {
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

func findAllListes(db *sql.DB) ([]Liste, error) {
	return nil, errors.New("Non implémenté")
}

func findLastListeScores(db *sql.DB) ([]Score, error) {
	return nil, errors.New("Non implémenté")
}

func (liste *Liste) findListeScores(db *sql.DB) ([]Score, error) {
	return nil, errors.New("Non implémenté")
}
