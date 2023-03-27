// Package core contient le code de base de Datapi. Pour le moment c'est encore le fourre-tout mais un jour ça sera beau !!!
package core

import "time"

// Diane représente les infos Diane sur une entreprise
type Diane struct {
	NumeroSiren                     string    `json:"numero_siren,omitempty"`
	ArreteBilan                     time.Time `json:"arrete_bilan_diane,omitempty"`
	AchatMarchandises               *float64  `json:"achat_marchandises,omitempty"`
	AchatMatieresPremieres          *float64  `json:"achat_matieres_premieres,omitempty"`
	AutonomieFinanciere             *float64  `json:"autonomie_financiere,omitempty"`
	AutresAchatsChargesExternes     *float64  `json:"autres_achats_charges_externes,omitempty"`
	AutresProduitsChargesReprises   *float64  `json:"autres_produits_charges_reprises,omitempty"`
	BeneficeOuPerte                 *float64  `json:"benefice_ou_perte,omitempty"`
	CAExportation                   *float64  `json:"ca_exportation,omitempty"`
	CapaciteAutofinancement         *float64  `json:"capacite_autofinancement,omitempty"`
	CapaciteRemboursement           *float64  `json:"capacite_remboursement,omitempty"`
	CAparEffectif                   *float64  `json:"ca_par_effectif,omitempty"`
	ChargeExceptionnelle            *float64  `json:"charge_exceptionnelle,omitempty"`
	ChargePersonnel                 *float64  `json:"charge_personnel,omitempty"`
	ChargesFinancieres              *float64  `json:"charges_financieres,omitempty"`
	ChiffreAffaire                  *float64  `json:"ca,omitempty"`
	ConcesBrevEtDroitsSim           *float64  `json:"conces_brev_et_droits_sim,omitempty"`
	ConcoursBancaireCourant         *float64  `json:"concours_bancaire_courant,omitempty"`
	Consommation                    *float64  `json:"consommation,omitempty"`
	CouvertureCaBesoinFdr           *float64  `json:"couverture_ca_besoin_fdr,omitempty"`
	CouvertureCaFdr                 *float64  `json:"couverture_ca_fdr,omitempty"`
	CreditClient                    *float64  `json:"credit_client,omitempty"`
	CreditFournisseur               *float64  `json:"credit_fournisseur,omitempty"`
	DegreImmoCorporelle             *float64  `json:"degre_immo_corporelle,omitempty"`
	DetteFiscaleEtSociale           *float64  `json:"dette_fiscale_et_sociale,omitempty"`
	DotationAmortissement           *float64  `json:"dotation_amortissement,omitempty"`
	EffectifConsolide               *int      `json:"effectif_consolide,omitempty"`
	EfficaciteEconomique            *float64  `json:"efficacite_economique,omitempty"`
	Endettement                     *float64  `json:"endettement,omitempty"`
	EndettementGlobal               *float64  `json:"endettement_global,omitempty"`
	EquilibreFinancier              *float64  `json:"equilibre_financier,omitempty"`
	ExcedentBrutDExploitation       *float64  `json:"excedent_brut_d_exploitation,omitempty"`
	Exercice                        float64   `json:"exercice_diane,omitempty"`
	Exportation                     *float64  `json:"exportation,omitempty"`
	FinancementActifCirculant       *float64  `json:"financement_actif_circulant,omitempty"`
	FraisDeRetD                     *float64  `json:"frais_de_RetD,omitempty"`
	ImpotBenefice                   *float64  `json:"impot_benefice,omitempty"`
	ImpotsTaxes                     *float64  `json:"impots_taxes,omitempty"`
	IndependanceFinanciere          *float64  `json:"independance_financiere,omitempty"`
	Interets                        *float64  `json:"interets,omitempty"`
	LiquiditeGenerale               *float64  `json:"liquidite_generale,omitempty"`
	LiquiditeReduite                *float64  `json:"liquidite_reduite,omitempty"`
	MargeCommerciale                *float64  `json:"marge_commerciale,omitempty"`
	NombreEtabSecondaire            *int      `json:"nombre_etab_secondaire,omitempty"`
	NombreFiliale                   *int      `json:"nombre_filiale,omitempty"`
	NombreMois                      *int      `json:"nombre_mois,omitempty"`
	OperationsCommun                *float64  `json:"operations_commun,omitempty"`
	PartAutofinancement             *float64  `json:"part_autofinancement,omitempty"`
	PartEtat                        *float64  `json:"part_etat,omitempty"`
	PartPreteur                     *float64  `json:"part_preteur,omitempty"`
	PartSalaries                    *float64  `json:"part_salaries,omitempty"`
	ParticipationSalaries           *float64  `json:"participation_salaries,omitempty"`
	Performance                     *float64  `json:"performance,omitempty"`
	PoidsBFRExploitation            *float64  `json:"poids_bfr_exploitation,omitempty"`
	ProcedureCollective             bool      `json:"procedure_collective,omitempty"`
	Production                      *float64  `json:"production,omitempty"`
	ProductiviteCapitalFinancier    *float64  `json:"productivite_capital_financier,omitempty"`
	ProductiviteCapitalInvesti      *float64  `json:"productivite_capital_investi,omitempty"`
	ProductivitePotentielProduction *float64  `json:"productivite_potentiel_production,omitempty"`
	ProduitExceptionnel             *float64  `json:"produit_exceptionnel,omitempty"`
	ProduitsFinanciers              *float64  `json:"produits_financiers,omitempty"`
	RendementBrutFondsPropres       *float64  `json:"rendement_brut_fonds_propres,omitempty"`
	RendementCapitauxPropres        *float64  `json:"rendement_capitaux_propres,omitempty"`
	RendementRessourcesDurables     *float64  `json:"rendement_ressources_durables,omitempty"`
	RentabiliteEconomique           *float64  `json:"rentabilite_economique,omitempty"`
	RentabiliteNette                *float64  `json:"rentabilite_nette,omitempty"`
	ResultatAvantImpot              *float64  `json:"resultat_avant_impot,omitempty"`
	ResultatExploitation            *float64  `json:"resultat_expl"`
	RotationStocks                  *float64  `json:"rotation_stocks,omitempty"`
	StatutJuridique                 string    `json:"statut_juridique,omitempty"`
	SubventionsDExploitation        *float64  `json:"subventions_d_exploitation,omitempty"`
	TailleCompoGroupe               *int      `json:"taille_compo_groupe,omitempty"`
	TauxDInvestissementProductif    *float64  `json:"taux_d_investissement_productif,omitempty"`
	TauxEndettement                 *float64  `json:"taux_endettement,omitempty"`
	TauxInteretFinancier            *float64  `json:"taux_interet_financier,omitempty"`
	TauxInteretSurCA                *float64  `json:"taux_interet_sur_ca,omitempty"`
	TauxMargeCommerciale            *float64  `json:"taux_marge_commerciale,omitempty"`
	TauxValeurAjoutee               *float64  `json:"taux_valeur_ajoutee,omitempty"`
	ValeurAjoutee                   *float64  `json:"valeur_ajoutee,omitempty"`
	NomEntreprise                   string    `json:"nom_entreprise,omitempty"`
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
