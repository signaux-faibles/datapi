package main

import (
	"time"

	"github.com/globalsign/mgo/bson"
)

type score struct {
	ID        bson.ObjectId `json:"-"`
	Score     float64       `json:"score"`
	Diff      float64       `json:"diff"`
	Timestamp time.Time     `json:"-"`
	Alert     string        `json:"alert"`
	Periode   string        `json:"periode"`
	Batch     string        `json:"batch"`
	Algo      string        `json:"algo"`
}

// Procol donne le statut et la date de statut pour une entreprise en matière de procédures collectives
type procol struct {
	Etat string    `json:"etat"`
	Date time.Time `json:"date_procol"`
}

// Etablissement is an object
type etablissement struct {
	ID    string `bson:"_id"`
	Value struct {
		Key        string      `json:"key"`
		Sirene     sirene      `json:"sirene"`
		Cotisation []float64   `json:"cotisation"`
		Debit      []debit     `json:"debit"`
		APDemande  []apDemande `json:"apdemande"`
		APConso    []apConso   `json:"apconso"`
		Compte     struct {
			Siret   string    `json:"siret"`
			Numero  string    `json:"numero_compte"`
			Periode time.Time `json:"periode"`
		} `json:"compte"`
		Effectif        []effectif    `json:"effectif"`
		DernierEffectif effectif      `json:"dernier_effectif"`
		Delai           []interface{} `json:"delai"`
		Procol          []procol      `json:"procol"`
		LastProcol      procol        `json:"last_procol"`
	} `bson:"value"`
	Scores []score `json:"scores"`
}

// Entreprise object
type entreprise struct {
	ID    map[string]string `json:"_id"`
	Value struct {
		Sirets   []string      `json:"sirets"`
		Diane    []diane       `json:"diane"`
		BDF      []interface{} `json:"bdf"`
		SireneUL sireneUL      `json:"sirene_ul"`
	} `bson:"value"`
}

// Effectif detail
type effectif struct {
	Periode  time.Time `json:"periode"`
	Effectif int       `json:"effectif"`
}

// Debit detail
type debit struct {
	PartOuvriere  float64   `json:"part_ouvriere"`
	PartPatronale float64   `json:"part_patronale"`
	Periode       time.Time `json:"periode"`
}

// APConso detail
type apConso struct {
	IDConso       string    `json:"id_conso"`
	HeureConsomme float64   `json:"heure_consomme"`
	Montant       float64   `json:"montant"`
	Effectif      int       `json:"int"`
	Periode       time.Time `json:"periode"`
}

// SireneUL detail
type sireneUL struct {
	Siren           string `json:"siren"`
	RaisonSociale   string `json:"raison_sociale"`
	StatutJuridique string `json:"statut_juridique"`
}

// APDemande detail
type apDemande struct {
	DateStatut time.Time `json:"date_statut"`
	Periode    struct {
		Start time.Time `json:"start"`
		End   time.Time `json:"end"`
	} `json:"periode"`
	EffectifAutorise int     `json:"effectif_autorise"`
	EffectifConsomme int     `json:"effectif_consomme"`
	IDDemande        string  `json:"id_conso"`
	Effectif         int     `json:"int"`
	MTA              float64 `json:"mta"`
	HTA              float64 `json:"hta"`
	MotifRecoursSE   int     `json:"motif_recours_se"`
	HeureConsomme    float64 `json:"heure_consomme"`
	Montant          float64 `json:"montant"`
}

type diane struct {
	ChiffreAffaire                  float64   `json:"ca,omitempty"`
	Exercice                        float64   `json:"exercice_diane,omitempty"`
	NomEntreprise                   string    `json:"nom_entreprise,omitempty"`
	NumeroSiren                     string    `json:"numero_siren,omitempty"`
	StatutJuridique                 string    `json:"statut_juridique,omitempty"`
	ProcedureCollective             bool      `json:"procedure_collective,omitempty"`
	EffectifConsolide               *int      `json:"effectif_consolide,omitempty"`
	DetteFiscaleEtSociale           *float64  `json:"dette_fiscale_et_sociale,omitempty"`
	FraisDeRetD                     *float64  `json:"frais_de_RetD,omitempty"`
	ConcesBrevEtDroitsSim           *float64  `json:"conces_brev_et_droits_sim,omitempty"`
	NombreEtabSecondaire            *int      `json:"nombre_etab_secondaire,omitempty"`
	NombreFiliale                   *int      `json:"nombre_filiale,omitempty"`
	TailleCompoGroupe               *int      `json:"taille_compo_groupe,omitempty"`
	ArreteBilan                     time.Time `json:"arrete_bilan_diane,omitempty"`
	NombreMois                      *int      `json:"nombre_mois,omitempty"`
	ConcoursBancaireCourant         *float64  `json:"concours_bancaire_courant,omitempty"`
	EquilibreFinancier              *float64  `json:"equilibre_financier,omitempty"`
	IndependanceFinanciere          *float64  `json:"independance_financiere,omitempty"`
	Endettement                     *float64  `json:"endettement,omitempty"`
	AutonomieFinanciere             *float64  `json:"autonomie_financiere,omitempty"`
	DegreImmoCorporelle             *float64  `json:"degre_immo_corporelle,omitempty"`
	FinancementActifCirculant       *float64  `json:"financement_actif_circulant,omitempty"`
	LiquiditeGenerale               *float64  `json:"liquidite_generale,omitempty"`
	LiquiditeReduite                *float64  `json:"liquidite_reduite,omitempty"`
	RotationStocks                  *float64  `json:"rotation_stocks,omitempty"`
	CreditClient                    *float64  `json:"credit_client,omitempty"`
	CreditFournisseur               *float64  `json:"credit_fournisseur,omitempty"`
	CAparEffectif                   *float64  `json:"ca_par_effectif,omitempty"`
	TauxInteretFinancier            *float64  `json:"taux_interet_financier,omitempty"`
	TauxInteretSurCA                *float64  `json:"taux_interet_sur_ca,omitempty"`
	EndettementGlobal               *float64  `json:"endettement_global,omitempty"`
	TauxEndettement                 *float64  `json:"taux_endettement,omitempty"`
	CapaciteRemboursement           *float64  `json:"capacite_remboursement,omitempty"`
	CapaciteAutofinancement         *float64  `json:"capacite_autofinancement,omitempty"`
	CouvertureCaFdr                 *float64  `json:"couverture_ca_fdr,omitempty"`
	CouvertureCaBesoinFdr           *float64  `json:"couverture_ca_besoin_fdr,omitempty"`
	PoidsBFRExploitation            *float64  `json:"poids_bfr_exploitation,omitempty"`
	Exportation                     *float64  `json:"exportation,omitempty"`
	EfficaciteEconomique            *float64  `json:"efficacite_economique,omitempty"`
	ProductivitePotentielProduction *float64  `json:"productivite_potentiel_production,omitempty"`
	ProductiviteCapitalFinancier    *float64  `json:"productivite_capital_financier,omitempty"`
	ProductiviteCapitalInvesti      *float64  `json:"productivite_capital_investi,omitempty"`
	TauxDInvestissementProductif    *float64  `json:"taux_d_investissement_productif,omitempty"`
	RentabiliteEconomique           *float64  `json:"rentabilite_economique,omitempty"`
	Performance                     *float64  `json:"performance,omitempty"`
	RendementBrutFondsPropres       *float64  `json:"rendement_brut_fonds_propres,omitempty"`
	RentabiliteNette                *float64  `json:"rentabilite_nette,omitempty"`
	RendementCapitauxPropres        *float64  `json:"rendement_capitaux_propres,omitempty"`
	RendementRessourcesDurables     *float64  `json:"rendement_ressources_durables,omitempty"`
	TauxMargeCommerciale            *float64  `json:"taux_marge_commerciale,omitempty"`
	TauxValeurAjoutee               *float64  `json:"taux_valeur_ajoutee,omitempty"`
	PartSalaries                    *float64  `json:"part_salaries,omitempty"`
	PartEtat                        *float64  `json:"part_etat,omitempty"`
	PartPreteur                     *float64  `json:"part_preteur,omitempty"`
	PartAutofinancement             *float64  `json:"part_autofinancement,omitempty"`
	CAExportation                   *float64  `json:"ca_exportation,omitempty"`
	AchatMarchandises               *float64  `json:"achat_marchandises,omitempty"`
	AchatMatieresPremieres          *float64  `json:"achat_matieres_premieres,omitempty"`
	Production                      *float64  `json:"production,omitempty"`
	MargeCommerciale                *float64  `json:"marge_commerciale,omitempty"`
	Consommation                    *float64  `json:"consommation,omitempty"`
	AutresAchatsChargesExternes     *float64  `json:"autres_achats_charges_externes,omitempty"`
	ValeurAjoutee                   *float64  `json:"valeur_ajoutee,omitempty"`
	ChargePersonnel                 *float64  `json:"charge_personnel,omitempty"`
	ImpotsTaxes                     *float64  `json:"impots_taxes,omitempty"`
	SubventionsDExploitation        *float64  `json:"subventions_d_exploitation,omitempty"`
	ExcedentBrutDExploitation       *float64  `json:"excedent_brut_d_exploitation,omitempty"`
	AutresProduitsChargesReprises   *float64  `json:"autres_produits_charges_reprises,omitempty"`
	DotationAmortissement           *float64  `json:"dotation_amortissement,omitempty"`
	ResultatExploitation            float64   `json:"resultat_expl"`
	OperationsCommun                *float64  `json:"operations_commun,omitempty"`
	ProduitsFinanciers              *float64  `json:"produits_financiers,omitempty"`
	ChargesFinancieres              *float64  `json:"charges_financieres,omitempty"`
	Interets                        *float64  `json:"interets,omitempty"`
	ResultatAvantImpot              *float64  `json:"resultat_avant_impot,omitempty"`
	ProduitExceptionnel             *float64  `json:"produit_exceptionnel,omitempty"`
	ChargeExceptionnelle            *float64  `json:"charge_exceptionnelle,omitempty"`
	ParticipationSalaries           *float64  `json:"participation_salaries,omitempty"`
	ImpotBenefice                   *float64  `json:"impot_benefice,omitempty"`
	BeneficeOuPerte                 *float64  `json:"benefice_ou_perte,omitempty"`
}

// Sirene detail
type sirene struct {
	Region          string   `json:"region"`
	Commune         string   `json:"commune"`
	RaisonSociale   string   `json:"raison_sociale"`
	TypeVoie        string   `json:"type_voie"`
	Siren           string   `json:"siren"`
	CodePostal      string   `json:"code_postal"`
	Lattitude       float64  `json:"lattitude"`
	Adresse         []string `json:"adresse"`
	Departement     string   `json:"departement"`
	NatureJuridique string   `json:"nature_juridique"`
	NumeroVoie      string   `json:"numero_voie"`
	Ape             string   `json:"ape"`
	Longitude       float64  `json:"longitude"`
	Nic             string   `json:"nic"`
	NicSiege        string   `json:"nic_siege"`
}
