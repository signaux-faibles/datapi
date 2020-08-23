package main

import (
	"strings"
	"time"

	"github.com/globalsign/mgo/bson"
	pgx "github.com/jackc/pgx/v4"
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
	ID    string `json:"_id"`
	Value struct {
		Key    string `json:"key"`
		Sirene sirene `json:"sirene"`

		Debit     []debit     `json:"debit"`
		APDemande []apDemande `json:"apdemande"`
		APConso   []apConso   `json:"apconso"`
		Compte    struct {
			Siret   string    `json:"siret"`
			Numero  string    `json:"numero_compte"`
			Periode time.Time `json:"periode"`
		} `json:"compte"`
		Periodes                []time.Time `json:"periodes"`
		Effectif                []int       `json:"effectif"`
		DebitPartPatronale      []float64   `json:"debit_part_patronale"`
		DebitPartOuvriere       []float64   `json:"debit_part_ouvriere"`
		DebitMontantMajorations []float64   `json:"debit_montant_majorations"`
		Cotisation              []float64   `json:"cotisation"`
		Delai                   []delai     `json:"delai"`
		Procol                  []procol    `json:"procol"`
		LastProcol              procol      `json:"last_procol"`
	} `bson:"value"`
	Scores []score `json:"scores"`
}

type delai struct {
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
	ID    map[string]string `json:"_id"`
	Value struct {
		Sirets   []string `json:"sirets"`
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
	Annee               int       `json:"annee_bdf"`
	ArreteBilan         time.Time `json:"arrete_bilan_bdf"`
	DelaiFournisseur    float64   `json:"delai_fournisseur"`
	DetteFiscale        float64   `json:"dette_fiscale"`
	FinancierCourtTerme float64   `json:"financier_court_terme"`
	FraisFinancier      float64   `json:"FraisFinancier"`
	PoidsFrng           float64   `json:"poids_frng"`
	TauxMarge           int       `json:"taux_marge"`
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

func (e entreprise) getBatch() *pgx.Batch {
	var batch pgx.Batch

	sqlEntreprise := `insert into entreprise 
	(siren, version, date_add, raison_sociale, statut_juridique)
	values ($1, -1, current_timestamp, $2, $3)`

	batch.Queue(
		sqlEntreprise,
		e.Value.SireneUL.Siren,
		e.Value.SireneUL.RaisonSociale,
		e.Value.SireneUL.StatutJuridique,
	)

	for _, b := range e.Value.BDF {
		sqlEntrepriseBDF := `insert into entreprise_bdf
			(siren, version, date_add, arrete_bilan_bdf, annee_bdf, delai_fournisseur, financier_court_terme,
	 		poids_frng, dette_fiscale, frais_financier, taux_marge)
			values ($1, -1, current_timestamp, $2, $3, $4, $5, $6, $7, $8, $9)`

		batch.Queue(
			sqlEntrepriseBDF,
			e.Value.SireneUL.Siren,
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

	for _, d := range e.Value.Diane {
		if d.ArreteBilan.IsZero() {
			continue
		}

		sqlEntrepriseDiane := `insert into entreprise_diane
				(siren, version, date_add, arrete_bilan_diane, chiffre_affaire, credit_client, resultat_expl, achat_marchandises,
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
				 taux_valeur_ajoutee, valeur_ajoutee)
				values ($1, -1, current_timestamp, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19,
				 $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34, $35, $36, $37, $38, $39, $40, $41, $42,
				 $43, $44, $45, $46, $47, $48, $49, $50, $51, $52, $53, $54, $55, $56, $57, $58, $59, $60, $61, $62, $63, $64, $65,
				 $66, $67, $68, $69, $70, $71, $72, $73);`

		batch.Queue(
			sqlEntrepriseDiane,
			e.Value.SireneUL.Siren, d.ArreteBilan, d.ChiffreAffaire, d.CreditClient, d.ResultatExploitation, d.AchatMarchandises,
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
			d.TauxValeurAjoutee, d.ValeurAjoutee,
		)
	}

	return &batch
}

func (e etablissement) getBatch() *pgx.Batch {
	var batch pgx.Batch

	sqlEtablissement := `insert into etablissement
		(siret, version, date_add, siren, adresse, ape, code_postal, commune, departement, lattitude, longitude, nature_juridique,
			numero_voie, region, type_voie)
		values ($1, -1, current_timestamp, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13);`

	batch.Queue(
		sqlEtablissement,
		e.Value.Key,
		e.Value.Key[0:9],
		strings.Join(e.Value.Sirene.Adresse, "\n"),
		e.Value.Sirene.Ape,
		e.Value.Sirene.CodePostal,
		e.Value.Sirene.Commune,
		e.Value.Sirene.Departement,
		e.Value.Sirene.Lattitude,
		e.Value.Sirene.Longitude,
		e.Value.Sirene.NatureJuridique,
		e.Value.Sirene.NumeroVoie,
		e.Value.Sirene.Region,
		e.Value.Sirene.TypeVoie,
	)

	for _, a := range e.Value.APConso {
		sqlAPConso := `insert into etablissement_apconso
		(siret, version, id_conso, heure_consomme, montant, effectif, periode)
		values ($1, -1, $2, $3, $4, $5, $6)`

		batch.Queue(
			sqlAPConso,
			e.Value.Key,
			a.IDConso,
			a.HeureConsomme,
			a.Montant,
			a.Effectif,
			a.Periode,
		)
	}

	for _, a := range e.Value.APDemande {
		sqlAPDemande := `insert into etablissement_apdemande
				(siret, version, id_demande, effectif_entreprise, effectif, date_statut, periode_start, periode_end,
				 hta, mta, effectif_autorise, motif_recours_se, heure_consomme, montant_consomme, effectif_consomme)
				values ($1, -1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`

		batch.Queue(
			sqlAPDemande,
			e.Value.Key,
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
		sqlUrssaf := `insert into etablissement_periode_urssaf
				(siret, version, periode, cotisation, part_patronale, part_salariale, montant_majorations, effectif, last_periode)
				values ($1, -1, $2, $3, $4, $5, $6, $7, $8)`

		batch.Queue(
			sqlUrssaf,
			e.Value.Key,
			a,
			e.Value.Cotisation[i],
			e.Value.DebitPartPatronale[i],
			e.Value.DebitPartOuvriere[i],
			e.Value.DebitMontantMajorations[i],
			e.Value.Effectif[i],
			i == len(e.Value.Periodes)-1,
		)
	}

	for _, a := range e.Value.Delai {
		sqlDelai := `insert into etablissement_delai (siret, version, action, annee_creation, date_creation, date_echeance,
				denomination, duree_delai, indic_6m, montant_echeancier, numero_compte, numero_contentieux, stade)
				values ($1, -1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

		batch.Queue(
			sqlDelai,
			e.Value.Key,
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

	return &batch
}
