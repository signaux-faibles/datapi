package main

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"time"

	"github.com/cnf/structhash"
	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
	pgx "github.com/jackc/pgx/v4"
	"github.com/spf13/viper"
)

type score struct {
	Siret     string        `json:"-"`
	Libelle   string        `json:"-"`
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
		Sirets   []string `json:"sirets" hash:"-"`
		Siren    string   `json:"siren"`
		Diane    []diane  `json:"diane"`
		BDF      []bdf    `json:"bdf"`
		SireneUL sireneUL `json:"sirene_ul"`
	} `bson:"value"`
}

// Effectif detail
type effectif struct {
	Periode  time.Time `json:"periode"`
	Effectif int       `json:"effectif"`
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
	IDDemande          string  `json:"id_conso"`
	MTA                float64 `json:"mta"`
	HTA                float64 `json:"hta"`
	MotifRecoursSE     int     `json:"motif_recours_se"`
	HeureConsomme      float64 `json:"heure_consomme"`
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
	ChiffreAffaire                  *float64  `json:"ca,omitempty"`
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
	ResultatExploitation            *float64  `json:"resultat_expl"`
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

func (e entreprise) getBatch(batch *pgx.Batch, htrees map[string]*htree) map[string][]int {
	htreeEntreprise := htrees["entreprise"]
	htreeEntrepriseBdF := htrees["entreprise_bdf"]
	htreeEntrepriseDiane := htrees["entreprise_diane"]

	updates := make(map[string][]int)

	sqlEntrepriseInsert := `insert into entreprise 
	(siren, raison_sociale, prenom1, prenom2, prenom3, prenom4, 
	  nom, nom_usage, statut_juridique, creation, hash)
  values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);`

	hash := structhash.Md5(e.Value.SireneUL, 0)

	if id, ok := htreeEntreprise.contains(hash); !ok {
		batch.Queue(
			sqlEntrepriseInsert,
			e.Value.SireneUL.Siren,
			e.Value.SireneUL.RaisonSociale,
			e.Value.SireneUL.Prenom1UniteLegale,
			e.Value.SireneUL.Prenom2UniteLegale,
			e.Value.SireneUL.Prenom3UniteLegale,
			e.Value.SireneUL.Prenom4UniteLegale,
			e.Value.SireneUL.NomUniteLegale,
			e.Value.SireneUL.NomUsageUniteLegale,
			e.Value.SireneUL.StatutJuridique,
			e.Value.SireneUL.Creation,
			hash,
		)
	} else {
		updates["entreprise"] = append(updates["entreprise"], id)
	}

	for _, b := range e.Value.BDF {
		sqlEntrepriseBDF := `insert into entreprise_bdf
			(siren, arrete_bilan_bdf, annee_bdf, delai_fournisseur, financier_court_terme,
	 		poids_frng, dette_fiscale, frais_financier, taux_marge, hash)
			values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);`

		hash := structhash.Md5(b, 0)
		if id, ok := htreeEntrepriseBdF.contains(hash); !ok {
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
				hash,
			)
		} else {
			updates["entreprise_bdf"] = append(updates["entreprise_bdf"], id)
		}
	}

	for _, d := range e.Value.Diane {
		if d.ArreteBilan.IsZero() {
			continue
		}

		sqlEntrepriseDiane := `insert into entreprise_diane
				(siren, arrete_bilan_diane, chiffre_affaire, credit_client, resultat_expl, achat_marchandises,
				 achat_matieres_premieres, autonomie_financiere, autres_achats_charges_externes, autres_produits_charges_reprises,
				 ca_exportation, capacite_autofinancement, capacite_remboursement, charge_exceptionnelle, charge_personnel,
				 charges_financieres, conces_brev_et_droits_sim, consommation, couverture_ca_besoin_fdr, couverture_ca_fdr,
				 credit_fournisseur, degre_immo_corporelle, dette_fiscale_et_sociale, dotation_amortissement, endettement,
				 endettement_global, equilibre_financier, excedent_brut_d_exploitation, exercice_diane, exportation,
				 financement_actif_circulant, frais_de_RetD, impot_benefice, impots_taxes, independance_financiere, interets,
				 liquidite_generale, liquidite_reduite, marge_commerciale, nombre_etab_secondaire, nombre_filiale, nombre_mois,
				 operations_commun, part_autofinancement, part_etat, part_preteur, part_salaries, participation_salaries,
				 performance, poids_bfr_exploitation, procedure_collective, production, productivite_capital_financier,
				 productivite_capital_investi, productivite_potentiel_production, produit_exceptionnel, produits_financiers,
				 rendement_brut_fonds_propres, rendement_capitaux_propres, rendement_ressources_durables, rentabilite_economique,
				 rentabilite_nette, resultat_avant_impot, rotation_stocks, statut_juridique, subventions_d_exploitation,
				 taille_compo_groupe, taux_d_investissement_productif, taux_endettement, taux_interet_financier, taux_interet_sur_ca,
				 taux_valeur_ajoutee, valeur_ajoutee, hash)
				values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19,
				 $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34, $35, $36, $37, $38, $39, $40, $41, $42,
				 $43, $44, $45, $46, $47, $48, $49, $50, $51, $52, $53, $54, $55, $56, $57, $58, $59, $60, $61, $62, $63, $64, $65,
				 $66, $67, $68, $69, $70, $71, $72, $73, $74);`

		hash := structhash.Md5(d, 0)
		if id, ok := htreeEntrepriseDiane.contains(hash); !ok {
			batch.Queue(
				sqlEntrepriseDiane,
				d.NumeroSiren, d.ArreteBilan, d.ChiffreAffaire, d.CreditClient, d.ResultatExploitation, d.AchatMarchandises,
				d.AchatMatieresPremieres, d.AutonomieFinanciere, d.AutresAchatsChargesExternes, d.AutresProduitsChargesReprises,
				d.CAExportation, d.CapaciteAutofinancement, d.CapaciteRemboursement, d.ChargeExceptionnelle, d.ChargePersonnel,
				d.ChargesFinancieres, d.ConcesBrevEtDroitsSim, d.Consommation, d.CouvertureCaBesoinFdr, d.CouvertureCaFdr,
				d.CreditFournisseur, d.DegreImmoCorporelle, d.DetteFiscaleEtSociale, d.DotationAmortissement, d.Endettement,
				d.EndettementGlobal, d.EquilibreFinancier, d.ExcedentBrutDExploitation, d.Exercice, d.Exportation,
				d.FinancementActifCirculant, d.FraisDeRetD, d.ImpotBenefice, d.ImpotsTaxes, d.IndependanceFinanciere, d.Interets,
				d.LiquiditeGenerale, d.LiquiditeReduite, d.MargeCommerciale, d.NombreEtabSecondaire, d.NombreFiliale, d.NombreMois,
				d.OperationsCommun, d.PartAutofinancement, d.PartEtat, d.PartPreteur, d.PartSalaries, d.ParticipationSalaries,
				d.Performance, d.PoidsBFRExploitation, d.ProcedureCollective, d.Production, d.ProductiviteCapitalFinancier,
				d.ProductiviteCapitalInvesti, d.ProductivitePotentielProduction, d.ProduitExceptionnel, d.ProduitsFinanciers,
				d.RendementBrutFondsPropres, d.RendementCapitauxPropres, d.RendementRessourcesDurables, d.RentabiliteEconomique,
				d.RentabiliteNette, d.ResultatAvantImpot, d.RotationStocks, d.StatutJuridique, d.SubventionsDExploitation,
				d.TailleCompoGroupe, d.TauxDInvestissementProductif, d.TauxEndettement, d.TauxInteretFinancier, d.TauxInteretSurCA,
				d.TauxValeurAjoutee, d.ValeurAjoutee, hash,
			)
		} else {
			updates["entreprise_diane"] = append(updates["entreprise_diane"], id)
		}
	}

	return updates
}

func (e etablissement) getBatch(batch *pgx.Batch, htrees map[string]*htree) map[string][]int {
	htreeEtablissement := htrees["etablissement"]
	htreeEtablissementAPConso := htrees["etablissement_apconso"]
	htreeEtablissementAPDemande := htrees["etablissement_apdemande"]
	htreeEtablissementPeriodeUrssaf := htrees["etablissement_periode_urssaf"]
	htreeEtablissementDelai := htrees["etablissement_delai"]
	htreeEtablissementProcol := htrees["etablissement_procol"]
	htreeScore := htrees["score"]

	updates := make(map[string][]int)

	sqlEtablissement := `insert into etablissement
		(siret, siren, siege, complement_adresse, numero_voie, indice_repetition, 
		type_voie, voie, commune, commune_etranger, distribution_speciale, code_commune,
		cedex, code_pays_etranger, pays_etranger, code_postal, departement,
		code_activite, nomen_activite, creation, latitude, longitude, hash)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, 
		$13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23);`

	hash := structhash.Md5(e.Value.Sirene, 0)

	defaultNomen := "NAFRev2"

	if id, ok := htreeEtablissement.contains(hash); !ok {
		e.Value.Sirene.Siret = e.Value.Key
		batch.Queue(
			sqlEtablissement,
			e.Value.Sirene.Siret,
			e.Value.Sirene.Siren,
			e.Value.Sirene.Siege,
			e.Value.Sirene.ComplementAdresse,
			e.Value.Sirene.NumVoie,
			e.Value.Sirene.IndRep,
			e.Value.Sirene.TypeVoie,
			e.Value.Sirene.Voie,
			e.Value.Sirene.Commune,
			e.Value.Sirene.CommuneEtranger,
			e.Value.Sirene.DistributionSpeciale,
			e.Value.Sirene.CodeCommune,
			e.Value.Sirene.Cedex,
			e.Value.Sirene.CodePaysEtranger,
			e.Value.Sirene.PaysEtranger,
			e.Value.Sirene.CodePostal,
			e.Value.Sirene.Departement,
			coalescepString(e.Value.Sirene.APE, e.Value.Sirene.CodeActivite),
			coalescepString(e.Value.Sirene.NomenActivite, &defaultNomen),
			e.Value.Sirene.Creation,
			e.Value.Sirene.Latitude,
			e.Value.Sirene.Longitude,
			hash,
		)
	} else {
		updates["etablissement"] = append(updates["etablissement"], id)
	}

	for _, a := range e.Value.APConso {
		sqlAPConso := `insert into etablissement_apconso
		(siret, siren, id_conso, heure_consomme, montant, effectif, periode, hash)
		values ($1, $2, $3, $4, $5, $6, $7, $8);`

		a.Siret = e.Value.Key

		hash := structhash.Md5(a, 0)

		if id, ok := htreeEtablissementAPConso.contains(hash); !ok {
			batch.Queue(
				sqlAPConso,
				a.Siret,
				a.Siret[0:9],
				a.IDConso,
				a.HeureConsomme,
				a.Montant,
				a.Effectif,
				a.Periode,
				hash,
			)
		} else {
			updates["etablissement_apconso"] = append(updates["etablissement_apconso"], id)
		}
	}

	periodesOrigine, _ := time.Parse("20060102", "20180101")
	for _, a := range e.Value.APDemande {
		sqlAPDemande := `insert into etablissement_apdemande
				(siret, siren, id_demande, effectif_entreprise, effectif, date_statut, periode_start, periode_end,
				 hta, mta, effectif_autorise, motif_recours_se, heure_consomme, montant_consomme, effectif_consomme, hash)
				values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`

		a.Siret = e.Value.Key

		hash := structhash.Md5(a, 0)

		if id, ok := htreeEtablissementAPDemande.contains(hash); !ok {
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
				hash,
			)
		} else {
			updates["etablissement_apdemande"] = append(updates["etablissement_apdemande"], id)
		}
	}

	for i, a := range e.Value.Periodes {
		if *e.Value.Cotisation[i]+*e.Value.DebitPartPatronale[i]+*e.Value.DebitPartOuvriere[i]+*e.Value.DebitMontantMajorations[i] != 0 ||
			e.Value.Effectif[i] != nil && e.Value.Periodes[i].After(periodesOrigine) {
			sqlUrssaf := `insert into etablissement_periode_urssaf
				(siret, siren, periode, cotisation, part_patronale, part_salariale, montant_majorations, effectif, hash)
				values ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

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

			hash := structhash.Md5(periode, 0)

			if id, ok := htreeEtablissementPeriodeUrssaf.contains(hash); !ok {
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
					hash,
				)
			} else {
				updates["etablissement_periode_urssaf"] = append(updates["etablissement_periode_urssaf"], id)
			}
		}
	}

	for _, a := range e.Value.Delai {
		sqlDelai := `insert into etablissement_delai (siret, siren, action, annee_creation, date_creation, date_echeance,
				denomination, duree_delai, indic_6m, montant_echeancier, numero_compte, numero_contentieux, stade, hash)
				values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14);`

		a.Siret = e.Value.Key
		hash := structhash.Md5(a, 0)

		if id, ok := htreeEtablissementDelai.contains(hash); !ok {

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
				hash,
			)
		} else {
			updates["etablissement_delai"] = append(updates["etablissement_delai"], id)
		}
	}

	for _, p := range e.Value.Procol {
		sqlProcol := `insert into etablissement_procol (siret, siren, date_effet,
			action_procol, stade_procol, hash) values ($1, $2, $3, $4, $5, $6);`

		p.Siret = e.Value.Key

		hash := structhash.Md5(p, 0)

		if id, ok := htreeEtablissementProcol.contains(hash); !ok {
			batch.Queue(
				sqlProcol,
				p.Siret,
				p.Siret[0:9],
				p.DateEffet,
				p.Action,
				p.Stade,
				hash,
			)
		} else {
			updates["etablissement_procol"] = append(updates["etablissement_procol"], id)
		}
	}

	for _, s := range groupScores(e.Scores) {
		sqlScore := `insert into score (siret, siren, libelle_liste, batch, algo, periode, score, diff, alert, hash)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

		s.Siret = e.Value.Key
		s.Libelle = s.toLibelle()

		if s.Algo == viper.GetString("algoImport") {
			hash := structhash.Md5(s, 0)
			if id, ok := htreeScore.contains(hash); !ok {
				batch.Queue(sqlScore,
					s.Siret,
					s.Siret[:9],
					s.toLibelle(),
					s.Batch,
					s.Algo,
					s.Periode,
					s.Score,
					s.Diff,
					s.Alert,
					hash,
				)
			} else {
				updates["score"] = append(updates["score"], id)
			}
		}
	}

	return updates
}

func groupScores(scores []score) []score {
	var a = make(map[string]score)
	for _, s := range scores {
		key := s.Batch + s.Algo
		if i, ok := a[key]; !ok || i.Timestamp.After(s.Timestamp) {
			a[key] = s
		}
	}
	var res []score
	for _, i := range a {
		res = append(res, i)
	}
	return res
}

func (s score) toLibelle() string {
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
	return s.Batch[0:4] + " - " + months[month] + " " + year
}

func scoreToListe() pgx.Batch {
	var batch pgx.Batch
	batch.Queue(`insert into liste (libelle, batch, algo) 
	select distinct libelle_liste as libelle, batch, algo from score;`)
	return batch
}

type importObject struct {
	ID    string      `json:"_id"`
	Value importValue `json:"value"`
}

type importValue struct {
	Sirets   *[]string      `json:"sirets"`
	Bdf      *[]interface{} `json:"bdf"`
	Diane    *[]diane       `json:"diane"`
	SireneUL *sireneUL      `json:"sirene_ul"`
}

func importHandler(c *gin.Context) {
	tx, err := db.Begin(context.Background())
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	log.Print("preparing import (caching hashes and upgrading version)")
	htrees, err := prepareImport(&tx)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	sourceEntreprise := viper.GetString("sourceEntreprise")
	log.Printf("processing entreprise file %s", sourceEntreprise)
	err = processEntreprise(sourceEntreprise, htrees, &tx)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	sourceEtablissement := viper.GetString("sourceEtablissement")
	log.Printf("processing etablissement file %s", sourceEtablissement)
	err = processEtablissement(sourceEtablissement, htrees, &tx)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	log.Print("refreshing materialized views")
	err = refreshMaterializedViews(&tx)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	log.Print("commiting changes to database")
	tx.Commit(context.Background())
	log.Print("drop dead data")
	_, err = db.Exec(context.Background(), "vacuum;")
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
}

func processEntreprise(fileName string, htrees map[string]*htree, tx *pgx.Tx) error {
	file, err := os.Open(fileName)
	if err != nil {
		log.Printf("error opening file: %s", err.Error())

		return err
	}
	defer file.Close()
	batches, wg := newBatchRunner(tx)
	unzip, err := gzip.NewReader(file)
	decoder := json.NewDecoder(unzip)
	i := 0
	var batch pgx.Batch

	var updates = make(map[string][]int)

	for {
		var e entreprise
		err := decoder.Decode(&e)
		if err != nil {
			updatesToBatch(updates, &batch)
			batches <- batch
			close(batches)
			wg.Wait()
			fmt.Printf("\033[2K\r%s terminated: %d objects inserted\n", fileName, i)
			if err == io.EOF {
				return nil
			}
			return err
		}

		if e.Value.SireneUL.Siren != "" {
			newUpdates := e.getBatch(&batch, htrees)
			for k, v := range newUpdates {
				updates[k] = append(updates[k], v...)
			}
			i++
			if math.Mod(float64(i), 1000) == 0 {
				updatesToBatch(updates, &batch)
				batches <- batch
				updates = make(map[string][]int)
				batch = pgx.Batch{}
				fmt.Printf("\033[2K\r%s: %d objects inserted", fileName, i)
			}
		}
	}
}

func updatesToBatch(updates map[string][]int, batch *pgx.Batch) {
	for table, ids := range updates {
		sqlUpdate := fmt.Sprintf(`update %s set version = 0 where id = any($1)`, table)
		batch.Queue(sqlUpdate, ids)
	}
}

func processEtablissement(fileName string, htrees map[string]*htree, tx *pgx.Tx) error {
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
	decoder := json.NewDecoder(unzip)

	i := 0
	var batch pgx.Batch
	var updates = make(map[string][]int)

	for {
		var e etablissement
		err := decoder.Decode(&e)
		if err != nil {
			updatesToBatch(updates, &batch)
			batches <- batch
			batches <- scoreToListe()
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
			newUpdates := e.getBatch(&batch, htrees)
			for k, v := range newUpdates {
				updates[k] = append(updates[k], v...)
			}

			i++
			if math.Mod(float64(i), 1000) == 0 {
				batches <- batch
				batch = pgx.Batch{}
				fmt.Printf("\033[2K\r%s: %d objects sent to postgres", fileName, i)
			}
		}
	}
}

func refreshMaterializedViews(tx *pgx.Tx) error {
	views := []string{
		"v_roles",
		"v_diane_variation_ca",
		"v_last_effectif",
		"v_last_procol",
		"v_hausse_urssaf",
		"v_apdemande",
		"v_alert_etablissement",
		"v_alert_entreprise"}
	for _, v := range views {
		fmt.Printf("\033[2K\rrefreshing %s", v)
		_, err := (*tx).Exec(context.Background(), fmt.Sprintf("refresh materialized view %s", v))
		if err != nil {
			return err
		}
	}
	fmt.Println("")
	log.Printf("success: %d views refreshed\n", len(views))
	return nil
}

func prepareImport(tx *pgx.Tx) (map[string]*htree, error) {
	var batch pgx.Batch
	var tables = []string{
		"entreprise",
		"entreprise_bdf",
		"entreprise_diane",
		"etablissement",
		"etablissement_apconso",
		"etablissement_apdemande",
		"etablissement_periode_urssaf",
		"etablissement_delai",
		"etablissement_procol",
		"score",
		"liste",
	}
	for _, table := range tables {
		batch.Queue(fmt.Sprintf("update %s set version = version + 1 returning id, case when version = 1 then hash else null end", table))
	}
	var htrees = make(map[string]*htree)
	r := (*tx).SendBatch(context.Background(), &batch)
	for _, table := range tables {
		tree := &htree{}
		fmt.Printf("update table %s\n", table)
		rows, err := r.Query()
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var hash []byte
			var id int
			err := rows.Scan(&id, &hash)
			if err != nil {
				return nil, err
			}
			if hash != nil {
				tree.insert(hash, id)
			}
		}
		htrees[table] = tree
	}
	return htrees, r.Close()
}
