package core

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
	pgx "github.com/jackc/pgx/v4"
	"github.com/spf13/viper"
)

type score struct {
	Siret         string        `json:"-"`
	Siren         string        `json:"siret"`
	Libelle       string        `json:"-"`
	ID            bson.ObjectId `json:"-"`
	Score         float64       `json:"score"`
	Diff          float64       `json:"diff"`
	Timestamp     time.Time     `json:"-"`
	Alert         string        `json:"alert"`
	Periode       string        `json:"periode"`
	Batch         string        `json:"batch"`
	Algo          string        `json:"algo"`
	ExplSelection struct {
		SelectConcerning [][]string `json:"select_concerning"`
		SelectReassuring [][]string `json:"select_reassuring"`
	} `json:"expl_selection"`
	MacroExpl             map[string]float64 `json:"macro_expl"`
	MicroExpl             map[string]float64 `json:"micro_expl"`
	MacroRadar            map[string]float64 `json:"macro_radar"`
	AlertPreRedressements string             `json:"alert_pre_redressements"`
}

// Procol donne le statut et la date de statut pour une entreprise en matière de procédures collectives
type procol struct {
	Siret     string    `json:"siret"`
	DateEffet time.Time `json:"date_effet"`
	Action    string    `json:"action_procol"`
	Stade     string    `json:"stade_procol"`
}

// Etablissement is an object
type etablissement struct {
	ID    string `json:"_id"`
	Value struct {
		Key       string      `json:"key"`
		Sirene    sirene      `json:"sirene"`
		Debit     []debit     `json:"debit"`
		APDemande []apDemande `json:"apdemande"`
		APConso   []apConso   `json:"apconso"`
		Compte    struct {
			Siret   string    `json:"siret"`
			Numero  string    `json:"numero_compte"`
			Periode time.Time `json:"periode"`
		} `json:"compte"`
		Periodes                []time.Time `json:"periodes"`
		Effectif                []*int      `json:"effectif"`
		DebitPartPatronale      []*float64  `json:"debit_part_patronale"`
		DebitPartOuvriere       []*float64  `json:"debit_part_ouvriere"`
		DebitMontantMajorations []*float64  `json:"debit_montant_majorations"`
		Cotisation              []*float64  `json:"cotisation"`
		Delai                   []delai     `json:"delai"`
		Procol                  []procol    `json:"procol"`
	} `bson:"value"`
	Scores []score `json:"scores"`
}

type paydex struct {
	Siren      string    `json:"-"`
	DateValeur time.Time `json:"date_valeur"`
	NBJours    int       `json:"nb_jours"`
}

type delai struct {
	Siret             string    `json:"siret"`
	Action            string    `json:"action"`
	AnneeCreation     int       `json:"anne_creation"`
	DateCreation      time.Time `json:"date_creation"`
	DateEcheance      time.Time `json:"date_echeance"`
	Denomination      string    `json:"denomination"`
	DureeDelai        int       `json:"duree_delai"`
	Indic6m           string    `json:"indic_6m"`
	MontantEcheancier float64
	NumeroCompte      string `json:"numero_compte"`
	NumeroContentieux string `json:"numero_contentieux"`
	Stade             string `json:"stade"`
}

// Entreprise object
type entreprise struct {
	ID    string `json:"_id"`
	Value struct {
		Sirets     []string    `json:"sirets" hash:"-"`
		Siren      string      `json:"siren"`
		Diane      []diane     `json:"diane"`
		BDF        []bdf       `json:"bdf"`
		Ellisphere *ellisphere `json:"ellisphere"`
		SireneUL   sireneUL    `json:"sirene_ul"`
		Paydex     *[]paydex   `json:"paydex"`
	} `bson:"value"`
}

// Debit detail
type debit struct {
	PartOuvriere       float64   `json:"part_ouvriere"`
	PartPatronale      float64   `json:"part_patronale"`
	MontantMajorations float64   `json:"montant_majorations"`
	Periode            time.Time `json:"periode"`
}

// SireneUL detail
type sireneUL struct {
	Siren               string     `json:"siren,omitempty"`
	RaisonSociale       string     `json:"raison_sociale"`
	Prenom1UniteLegale  string     `json:"prenom1_unite_legale,omitempty"`
	Prenom2UniteLegale  string     `json:"prenom2_unite_legale,omitempty"`
	Prenom3UniteLegale  string     `json:"prenom3_unite_legale,omitempty"`
	Prenom4UniteLegale  string     `json:"prenom4_unite_legale,omitempty"`
	NomUniteLegale      string     `json:"nom_unite_legale,omitempty"`
	NomUsageUniteLegale string     `json:"nom_usage_unite_legale,omitempty"`
	StatutJuridique     string     `json:"statut_juridique"`
	Creation            *time.Time `json:"date_creation,omitempty"`
}

type ellisphere struct {
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

// APConso detail
type apConso struct {
	Siret         string    `json:"siret"`
	IDConso       string    `json:"id_conso"`
	HeureConsomme float64   `json:"heure_consomme"`
	Montant       float64   `json:"montant"`
	Effectif      int       `json:"effectif"`
	Periode       time.Time `json:"periode"`
}

// APDemande detail
type apDemande struct {
	Siret      string    `json:"siret"`
	DateStatut time.Time `json:"date_statut"`
	Periode    struct {
		Start time.Time `json:"start"`
		End   time.Time `json:"end"`
	} `json:"periode"`
	EffectifEntreprise int     `json:"effectif_entreprise"`
	Effectif           int     `json:"effectif"`
	EffectifAutorise   int     `json:"effectif_autorise"`
	EffectifConsomme   int     `json:"effectif_consomme"`
	IDDemande          string  `json:"id_demande"`
	MTA                float64 `json:"mta"`
	HTA                float64 `json:"hta"`
	MotifRecoursSE     int     `json:"motif_recours_se"`
	HeureConsomme      float64 `json:"heure_consommee"`
	MontantConsomme    float64 `json:"montant_consommee"`
}

type bdf struct {
	Siren               string    `json:"siren,omitempty"`
	Annee               int       `json:"annee_bdf"`
	ArreteBilan         time.Time `json:"arrete_bilan_bdf"`
	DelaiFournisseur    float64   `json:"delai_fournisseur"`
	DetteFiscale        float64   `json:"dette_fiscale"`
	FinancierCourtTerme float64   `json:"financier_court_terme"`
	FraisFinancier      float64   `json:"FraisFinancier"`
	PoidsFrng           float64   `json:"poids_frng"`
	TauxMarge           float64   `json:"taux_marge"`
}

type diane struct {
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

// Sirene detail
type sirene struct {
	Siret                string     `json:"-"`
	Siren                string     `json:"siren,omitempty"`
	Nic                  string     `json:"nic,omitempty"`
	Siege                bool       `json:"siege,omitempty"`
	ComplementAdresse    *string    `json:"complement_adresse,omitempty"`
	NumVoie              *string    `json:"numero_voie,omitempty"`
	IndRep               *string    `json:"indrep,omitempty"`
	TypeVoie             *string    `json:"type_voie,omitempty"`
	Voie                 *string    `json:"voie,omitempty"`
	Commune              *string    `json:"commune,omitempty"`
	CommuneEtranger      *string    `json:"commune_etranger,omitempty"`
	DistributionSpeciale *string    `json:"distribution_speciale,omitempty"`
	CodeCommune          *string    `json:"code_commune,omitempty"`
	CodeCedex            *string    `json:"code_cedex,omitempty"`
	Cedex                *string    `json:"cedex,omitempty"`
	CodePaysEtranger     *string    `json:"code_pays_etranger,omitempty"`
	PaysEtranger         *string    `json:"pays_etranger,omitempty"`
	CodePostal           *string    `json:"code_postal,omitempty"`
	Departement          *string    `json:"departement,omitempty"`
	APE                  *string    `json:"ape,omitempty"`
	CodeActivite         *string    `json:"code_activite,omitempty"`
	NomenActivite        *string    `json:"nomen_activite,omitempty"`
	Creation             *time.Time `json:"date_creation,omitempty"`
	Longitude            float64    `json:"longitude,omitempty"`
	Latitude             float64    `json:"latitude,omitempty"`
}

func (e entreprise) intoBatch(batch *pgx.Batch) {
	sqlEntrepriseBDF := `insert into entreprise_bdf
	(siren, arrete_bilan_bdf, annee_bdf, delai_fournisseur, financier_court_terme,
	 poids_frng, dette_fiscale, frais_financier, taux_marge)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9);`

	for _, b := range e.Value.BDF {
		batch.Queue(
			sqlEntrepriseBDF,
			b.Siren,
			b.ArreteBilan,
			b.Annee,
			b.DelaiFournisseur,
			b.FinancierCourtTerme,
			b.PoidsFrng,
			b.DetteFiscale,
			b.FraisFinancier,
			b.TauxMarge,
		)
	}

	sqlEntrepriseDiane := `insert into entreprise_diane
		(siren, arrete_bilan_diane, achat_marchandises, achat_matieres_premieres, autonomie_financiere, 
		autres_achats_charges_externes, autres_produits_charges_reprises, benefice_ou_perte, ca_exportation,
		capacite_autofinancement, capacite_remboursement, ca_par_effectif, charge_exceptionnelle, charge_personnel,
	 	charges_financieres, chiffre_affaire, conces_brev_et_droits_sim, concours_bancaire_courant,	consommation, 
		couverture_ca_besoin_fdr, couverture_ca_fdr, credit_client, credit_fournisseur, degre_immo_corporelle,
		dette_fiscale_et_sociale, dotation_amortissement, effectif_consolide, efficacite_economique, endettement,
	 	endettement_global, equilibre_financier, excedent_brut_d_exploitation, exercice_diane, exportation,
	 	financement_actif_circulant, frais_de_RetD, impot_benefice, impots_taxes, independance_financiere, interets,
	 	liquidite_generale, liquidite_reduite, marge_commerciale, nombre_etab_secondaire, nombre_filiale, nombre_mois,
	 	operations_commun, part_autofinancement, part_etat, part_preteur, part_salaries, participation_salaries,
	 	performance, poids_bfr_exploitation, procedure_collective, production, productivite_capital_financier,
	 	productivite_capital_investi, productivite_potentiel_production, produit_exceptionnel, produits_financiers,
	 	rendement_brut_fonds_propres, rendement_capitaux_propres, rendement_ressources_durables, rentabilite_economique,
	 	rentabilite_nette, resultat_avant_impot, resultat_expl, rotation_stocks, statut_juridique, subventions_d_exploitation,
	 	taille_compo_groupe, taux_d_investissement_productif, taux_endettement, taux_interet_financier, taux_interet_sur_ca,
		taux_marge_commerciale,	taux_valeur_ajoutee, valeur_ajoutee)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19,
	 	$20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34, $35, $36, $37, $38, $39, $40, $41, $42,
	 	$43, $44, $45, $46, $47, $48, $49, $50, $51, $52, $53, $54, $55, $56, $57, $58, $59, $60, $61, $62, $63, $64, $65,
 	 	$66, $67, $68, $69, $70, $71, $72, $73, $74, $75, $76, $77, $78, $79);`

	for _, d := range e.Value.Diane {
		if d.ArreteBilan.IsZero() {
			continue
		}

		batch.Queue(
			sqlEntrepriseDiane,
			d.NumeroSiren, d.ArreteBilan, d.AchatMarchandises, d.AchatMatieresPremieres, d.AutonomieFinanciere,
			d.AutresAchatsChargesExternes, d.AutresProduitsChargesReprises, d.BeneficeOuPerte, d.CAExportation,
			d.CapaciteAutofinancement, d.CapaciteRemboursement, d.CAparEffectif, d.ChargeExceptionnelle, d.ChargePersonnel,
			d.ChargesFinancieres, d.ChiffreAffaire, d.ConcesBrevEtDroitsSim, d.ConcoursBancaireCourant, d.Consommation,
			d.CouvertureCaBesoinFdr, d.CouvertureCaFdr, d.CreditClient, d.CreditFournisseur, d.DegreImmoCorporelle,
			d.DetteFiscaleEtSociale, d.DotationAmortissement, d.EffectifConsolide, d.EfficaciteEconomique, d.Endettement,
			d.EndettementGlobal, d.EquilibreFinancier, d.ExcedentBrutDExploitation, d.Exercice, d.Exportation,
			d.FinancementActifCirculant, d.FraisDeRetD, d.ImpotBenefice, d.ImpotsTaxes, d.IndependanceFinanciere, d.Interets,
			d.LiquiditeGenerale, d.LiquiditeReduite, d.MargeCommerciale, d.NombreEtabSecondaire, d.NombreFiliale, d.NombreMois,
			d.OperationsCommun, d.PartAutofinancement, d.PartEtat, d.PartPreteur, d.PartSalaries, d.ParticipationSalaries,
			d.Performance, d.PoidsBFRExploitation, d.ProcedureCollective, d.Production, d.ProductiviteCapitalFinancier,
			d.ProductiviteCapitalInvesti, d.ProductivitePotentielProduction, d.ProduitExceptionnel, d.ProduitsFinanciers,
			d.RendementBrutFondsPropres, d.RendementCapitauxPropres, d.RendementRessourcesDurables, d.RentabiliteEconomique,
			d.RentabiliteNette, d.ResultatAvantImpot, d.ResultatExploitation, d.RotationStocks, d.StatutJuridique, d.SubventionsDExploitation,
			d.TailleCompoGroupe, d.TauxDInvestissementProductif, d.TauxEndettement, d.TauxInteretFinancier, d.TauxInteretSurCA,
			d.TauxMargeCommerciale, d.TauxValeurAjoutee, d.ValeurAjoutee,
		)

	}

	if e.Value.Ellisphere != nil {
		sqlEllisphere := `insert into entreprise_ellisphere
		(siren, code, refid, raison_sociale,	adresse, personne_pou_m,
		niveau_detention, part_financiere, code_filiere, refid_filiere,
		personne_pou_m_filiere) 
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
		el := *e.Value.Ellisphere
		el.Siren = e.Value.SireneUL.Siren
		batch.Queue(sqlEllisphere, el.Siren, el.CodeGroupe, el.RefIDGroupe, el.RaisocGroupe,
			el.AdresseGroupe, el.PersonnePouMGroupe, el.NiveauDetention, el.PartFinanciere,
			el.CodeFiliere, el.RefIDFiliere, el.PersonnePouMFiliere)
	}

	if e.Value.Paydex != nil {
		for _, p := range *e.Value.Paydex {
			sqlPaydex := `insert into entreprise_paydex
				(siren, date_valeur, nb_jours) 
				values ($1, $2, $3)`
			el := p
			el.Siren = e.Value.SireneUL.Siren
			batch.Queue(sqlPaydex, el.Siren, el.DateValeur, el.NBJours)
		}
	}
}

func (e etablissement) intoBatch(batch *pgx.Batch) {
	for _, a := range e.Value.APConso {
		sqlAPConso := `insert into etablissement_apconso
		(siret, siren, id_conso, heure_consomme, montant, effectif, periode)
		values ($1, $2, $3, $4, $5, $6, $7);`

		a.Siret = e.Value.Key

		batch.Queue(
			sqlAPConso,
			a.Siret,
			a.Siret[0:9],
			a.IDConso,
			a.HeureConsomme,
			a.Montant,
			a.Effectif,
			a.Periode,
		)

	}

	periodesOrigine, _ := time.Parse("20060102", "20180101")
	for _, a := range e.Value.APDemande {
		sqlAPDemande := `insert into etablissement_apdemande
				(siret, siren, id_demande, effectif_entreprise, effectif, date_statut, periode_start, periode_end,
				 hta, mta, effectif_autorise, motif_recours_se, heure_consomme, montant_consomme, effectif_consomme)
				values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`

		a.Siret = e.Value.Key

		batch.Queue(
			sqlAPDemande,
			a.Siret,
			a.Siret[0:9],
			a.IDDemande,
			a.EffectifEntreprise,
			a.Effectif,
			a.DateStatut,
			a.Periode.Start,
			a.Periode.End,
			a.HTA,
			a.MTA,
			a.EffectifAutorise,
			a.MotifRecoursSE,
			a.HeureConsomme,
			a.MontantConsomme,
			a.EffectifConsomme,
		)
	}

	for i, a := range e.Value.Periodes {
		if *e.Value.Cotisation[i]+*e.Value.DebitPartPatronale[i]+*e.Value.DebitPartOuvriere[i]+*e.Value.DebitMontantMajorations[i] != 0 ||
			e.Value.Effectif[i] != nil && e.Value.Periodes[i].After(periodesOrigine) {
			sqlUrssaf := `insert into etablissement_periode_urssaf
				(siret, siren, periode, cotisation, part_patronale, part_salariale, montant_majorations, effectif)
				values ($1, $2, $3, $4, $5, $6, $7, $8)`

			var periode = struct {
				Siret                   string
				Effectif                *int
				DebitPartPatronale      *float64
				DebitPartOuvriere       *float64
				DebitMontantMajorations *float64
				Cotisation              *float64
				Periode                 time.Time
			}{
				Siret:                   e.Value.Key,
				Periode:                 a,
				Effectif:                e.Value.Effectif[i],
				DebitPartPatronale:      e.Value.DebitPartPatronale[i],
				DebitPartOuvriere:       e.Value.DebitPartOuvriere[i],
				DebitMontantMajorations: e.Value.DebitMontantMajorations[i],
				Cotisation:              e.Value.Cotisation[i],
			}

			batch.Queue(
				sqlUrssaf,
				periode.Siret,
				periode.Siret[0:9],
				periode.Periode,
				periode.Cotisation,
				periode.DebitPartPatronale,
				periode.DebitPartOuvriere,
				periode.DebitMontantMajorations,
				periode.Effectif,
			)
		}
	}

	for _, a := range e.Value.Delai {
		sqlDelai := `insert into etablissement_delai (siret, siren, action, annee_creation, date_creation, date_echeance,
				denomination, duree_delai, indic_6m, montant_echeancier, numero_compte, numero_contentieux, stade)
				values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13);`

		a.Siret = e.Value.Key

		batch.Queue(
			sqlDelai,
			a.Siret,
			a.Siret[0:9],
			a.Action,
			a.AnneeCreation,
			a.DateCreation,
			a.DateEcheance,
			a.Denomination,
			a.DureeDelai,
			a.Indic6m,
			a.MontantEcheancier,
			a.NumeroCompte,
			a.NumeroContentieux,
			a.Stade,
		)
	}

	for _, p := range e.Value.Procol {
		sqlProcol := `insert into etablissement_procol (siret, siren, date_effet,
			action_procol, stade_procol) values ($1, $2, $3, $4, $5);`

		p.Siret = e.Value.Key

		batch.Queue(
			sqlProcol,
			p.Siret,
			p.Siret[0:9],
			p.DateEffet,
			p.Action,
			p.Stade,
		)
	}
}

func contains(array []string, test string) bool {
	for _, s := range array {
		if s == test {
			return true
		}
	}
	return false
}

func (s scoreFile) toLibelle() string {
	months := map[string]string{
		"01": "Janvier",
		"02": "Février",
		"03": "Mars",
		"04": "Avril",
		"05": "Mai",
		"06": "Juin",
		"07": "Juillet",
		"08": "Août",
		"09": "Septembre",
		"10": "Octobre",
		"11": "Novembre",
		"12": "Décembre",
	}
	year := "20" + s.Batch[0:2]
	month := s.Batch[2:4]
	return months[month] + " " + year
}

// func scoreToListe() pgx.Batch {
// 	var batch pgx.Batch
// 	batch.Queue(`insert into liste (libelle, batch, algo)
// 	select distinct libelle_liste as libelle, batch, algo from score;`)
// 	return batch
// }

func importHandler(c *gin.Context) {
	tx, err := Db().Begin(context.Background())
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	log.Print("preparing import (truncate tables)")
	err = prepareImport(&tx)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	sourceEntreprise := viper.GetString("sourceEntreprise")
	log.Printf("processing entreprise file %s", sourceEntreprise)
	err = processEntreprise(sourceEntreprise, &tx)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	sourceEtablissement := viper.GetString("sourceEtablissement")
	log.Printf("processing etablissement file %s", sourceEtablissement)
	err = processEtablissement(sourceEtablissement, &tx)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	// log.Print("refreshing materialized views")
	// err = refreshMaterializedViews(&tx)
	// if err != nil {
	// 	c.AbortWithError(500, err)
	// 	return
	// }
	log.Print("commiting changes to database")
	tx.Commit(context.Background())
	log.Print("drop dead data")
	_, err = Db().Exec(context.Background(), "vacuum;")
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
}

func processEntreprise(fileName string, tx *pgx.Tx) error {
	file, err := os.Open(fileName)
	if err != nil {
		log.Printf("error opening file: %s", err.Error())

		return err
	}
	defer file.Close()
	batches, wg := newBatchRunner(tx)
	unzip, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(unzip)
	i := 0
	var batch pgx.Batch

	for {
		var e entreprise
		err := decoder.Decode(&e)
		if err != nil {
			close(batches)
			wg.Wait()
			fmt.Printf("\033[2K\r%s terminated: %d objects inserted\n", fileName, i)
			if err == io.EOF {
				return nil
			}
			return err
		}

		if e.Value.SireneUL.Siren != "" {
			e.intoBatch(&batch)

			i++
			if math.Mod(float64(i), 1000) == 0 {
				batches <- batch
				batch = pgx.Batch{}
				fmt.Printf("\033[2K\r%s: %d objects inserted", fileName, i)
			}
		}
	}
}

func processEtablissement(fileName string, tx *pgx.Tx) error {
	file, err := os.Open(fileName)
	if err != nil {
		log.Printf("error opening file: %s", err.Error())
		return err
	}
	defer file.Close()
	if err != nil {
		return err
	}
	batches, wg := newBatchRunner(tx)
	unzip, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(unzip)

	i := 0
	var batch pgx.Batch

	for {
		var e etablissement
		err := decoder.Decode(&e)
		if err != nil {
			close(batches)
			wg.Wait()
			fmt.Printf("\033[2K\r%s terminated: %d objects inserted\n", fileName, i)
			if err == io.EOF {
				return nil
			}
			return err
		}

		if len(e.ID) > 14 && e.Value.Sirene.Departement != nil {
			e.Value.Key = e.ID[len(e.ID)-14:]
			e.intoBatch(&batch)

			i++
			if math.Mod(float64(i), 1000) == 0 {
				batches <- batch
				batch = pgx.Batch{}
				fmt.Printf("\033[2K\r%s: %d objects sent to postgres", fileName, i)
			}
		}
	}
}

// func refreshMaterializedViews(tx *pgx.Tx) error {
// 	//!\ v_summaries doit être rafraichie en dernier
// 	views := []string{
// 		"v_alert_entreprise",
// 		"v_alert_etablissement",
// 		"v_apdemande",
// 		"v_diane_variation_ca",
// 		"v_etablissement_raison_sociale",
// 		"v_hausse_urssaf",
// 		"v_last_effectif",
// 		"v_last_procol",
// 		"v_naf",
// 		"v_roles",
// 		"v_summaries",
// 	}
// 	for _, v := range views {
// 		fmt.Printf("\033[2K\rrefreshing %s", v)
// 		_, err := (*tx).Exec(context.Background(), fmt.Sprintf("refresh materialized view %s", v))
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	fmt.Println("")
// 	log.Printf("success: %d views refreshed\n", len(views))
// 	return nil
// }

func prepareImport(tx *pgx.Tx) error {
	// var batch pgx.Batch
	var tables = []string{
		"entreprise_ellisphere",
		"entreprise_bdf",
		"entreprise_diane",
		"entreprise_paydex",
		"etablissement_apconso",
		"etablissement_apdemande",
		"etablissement_periode_urssaf",
		"etablissement_delai",
		"etablissement_procol",
	}

	for _, table := range tables {
		_, err := (*tx).Exec(context.Background(), fmt.Sprintf("truncate table %s;", table))
		if err != nil {
			return err
		}
	}

	return nil
}

type scoreFile struct {
	Siren         string  `json:"siren"`
	Score         float64 `json:"score"`
	Diff          float64 `json:"diff"`
	Alert         string  `json:"alert"`
	Periode       string  `json:"periode"`
	Batch         string  `json:"batch"`
	Algo          string  `json:"algo"`
	ExplSelection struct {
		SelectConcerning [][]string `json:"selectConcerning"`
		SelectReassuring [][]string `json:"selectReassuring"`
	} `json:"explSelection"`
	MacroRadar            map[string]float64 `json:"macroRadar"`
	Redressements         []string           `json:"redressements"`
	AlertPreRedressements string             `json:"alertPreRedressements"`
}

func listImportHandler(c *gin.Context) {
	algo := c.Params.ByName("algo")
	if algo == "" {
		c.AbortWithStatusJSON(400, "please provide algo name !")
		return
	}

	filename := viper.GetString("listPath")
	file, err := os.Open(filename)
	if err != nil {
		c.AbortWithStatusJSON(500, "open file: "+err.Error())
		return
	}
	raw, err := ioutil.ReadAll(file)
	file.Close()
	if err != nil {
		c.AbortWithStatusJSON(500, "read file: "+err.Error())
		return
	}
	var scores []scoreFile
	err = json.Unmarshal(raw, &scores)
	if err != nil {
		c.AbortWithStatusJSON(500, "unmarshal JSON: "+err.Error())
		return
	}

	tx, err := Db().Begin(context.Background())
	if err != nil {
		c.AbortWithStatusJSON(500, "begin TX: "+err.Error())
	}

	_, err = tx.Exec(context.Background(), `create table tmp_score (
		siren text,
		score real,
		libelle_liste text,
		diff real,
		alert text,
		batch text,
		algo text,
		expl_selection_concerning jsonb default '{}', 
		expl_selection_reassuring jsonb default '{}',
		macro_radar jsonb default '{}',
		alert_pre_redressements text,
		redressements text[] default '{}'
	);`)
	if err != nil {
		c.AbortWithStatusJSON(500, "create tmp_score: "+err.Error())
		return
	}

	batch := &pgx.Batch{}
	for _, s := range scores {
		queueScoreToBatch(s, batch)
	}
	batch.Queue(`insert into score
			(siret, siren, libelle_liste, batch, algo, periode,
			score, diff, alert, expl_selection_concerning,
			expl_selection_reassuring, macro_radar,
			redressements, alert_pre_redressements)
		select
			e.siret, t.siren, t.libelle_liste, batch, $1,
			current_date, score, diff, alert, expl_selection_concerning,
			expl_selection_reassuring, macro_radar,
			redressements, alert_pre_redressements
		from tmp_score t
		inner join etablissement e on e.siren = t.siren and e.siege`, algo)

	results := tx.SendBatch(context.Background(), batch)
	err = results.Close()

	if err != nil {
		c.AbortWithStatusJSON(500, "execute batch: "+err.Error())
		return
	}
	err = tx.Commit(context.Background())
	if err != nil {
		c.AbortWithStatusJSON(500, "commit: "+err.Error())
	}
}

func queueScoreToBatch(s scoreFile, batch *pgx.Batch) {
	sqlScore := `insert into tmp_score (siren, libelle_liste, batch, algo, score, diff, alert,
 		expl_selection_concerning, expl_selection_reassuring, macro_radar, alert_pre_redressements, redressements)
   	values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	if s.ExplSelection.SelectConcerning == nil {
		s.ExplSelection.SelectConcerning = make([][]string, 0)
	}
	if s.ExplSelection.SelectReassuring == nil {
		s.ExplSelection.SelectReassuring = make([][]string, 0)
	}

	if s.MacroRadar == nil {
		s.MacroRadar = make(map[string]float64)
	}
	if s.Redressements == nil {
		s.Redressements = make([]string, 0)
	}

	batch.Queue(sqlScore,
		s.Siren,
		s.toLibelle(),
		s.Batch,
		s.Algo,
		s.Score,
		s.Diff,
		s.Alert,
		s.ExplSelection.SelectConcerning,
		s.ExplSelection.SelectReassuring,
		s.MacroRadar,
		s.AlertPreRedressements,
		s.Redressements,
	)
}