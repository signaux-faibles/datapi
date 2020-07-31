package main

import (
	"database/sql"
	"errors"
)

type Entreprise struct {
	Siren string `json:"siren"`
}

type Etablissement struct {
	Siren string `json:"siren"`
	Siret string `json:"siret"`
}

type Follow struct {
	Siren string `json:"siren"`
	UserID string `json:"userId"`
}

type Comment struct {
	Siren string `json:"siren"`
	UserID string `json:"userId"`
	Message string `json:"message"`
}

type Liste struct {
	ID string `json:"id"`
}

type Score struct {
	Siren string `json:"siren"`
	Siret string `json:"siret"`
	ListeID string `json:"listeId"`
	ScoreValue string `json:"scoreValue"`
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