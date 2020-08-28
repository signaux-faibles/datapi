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
	Siren  string `json:"siren"`
	Sirene struct {
		RaisonSociale   string `json:"raisonSociale"`
		StatutJuridique string `json:"statutJuridique"`
	}
	Diane                 []diane                `json:"diane"`
	Bdf                   []bdf                  `json:"bdf"`
	EtablissementsSummary []EtablissementSummary `json:"etablissementsSummary,omitempty"`
	Etablissements        []Etablissement        `json:"etablissements,omitEmpty"`
}

// EtablissementSummary …
type EtablissementSummary struct {
	Siret          string `json:"siret"`
	Departement    string `json:"departement"`
	Region         string `json:"region"`
	CodeActivite   string `json:"codeActivite"`
	CodeActiviteN1 string ``
}

type findEtablissementsParams struct {
	Sirets []string `json:"sirets"`
	Sirens []string `json:"sirens"`
}

// Etablissements Collection d'établissements
type Etablissements struct {
	Query struct {
		Sirets []string `json:"sirets"`
		Sirens []string `json:"sirens"`
	} `json:"-"`
	Etablissements map[string]Etablissement `json:"etablissements"`
	Entreprises    map[string]Entreprise    `json:"entreprises"`
}

// Etablissement type établissement pour l'API
type Etablissement struct {
	Siren      string      `json:"siren"`
	Siret      string      `json:"siret"`
	Entreprise *Entreprise `json:"entreprise,omitempty"`
	VisiteFCE  *bool       `json:"visiteFCE"`
	Sirene     struct {
		Adresse    string  `json:"adresse"`
		Dept       string  `json:"dept"`
		CodeDept   string  `json:"codeDept"`
		CodePostal string  `json:"codePostal"`
		Region     string  `json:"region"`
		Lattitude  float64 `json:"lattitude"`
		Longitude  float64 `json:"longitude"`
		NumeroVoie string  `json:"numeroVoie"`
		TypeVoie   string  `json:"typeVoie"`
		Commune    string  `json:"string"`
		NAF        struct {
			APE     string `json:"activite"`
			Secteur string `json:"secteur"`
			N1      string `json:"n1"`
			N2      string `json:"n2"`
			N3      string `json:"n3"`
			N4      string `json:"n4"`
			N5      string `json:"n5"`
		} `json:"naf"`
	} `json:"sirene"`
	PeriodeUrssaf EtablissementPeriodeUrssaf `json:"periodeUrssaf,omitempty"`
	Delai         []EtablissementDelai       `json:"delai,omitempty"`
	APDemande     []EtablissementAPDemande   `json:"apDemande,omitempty"`
	APConso       []EtablissementAPConso     `json:"apConso,omitempty"`
	Procol        []EtablissementProcol      `json:"procol,omitempty"`
	Scores        []EtablissementScore       `json:"scores,omitempty"`
}

// EtablissementPeriodeUrssaf …
type EtablissementPeriodeUrssaf struct {
	Periode            []*time.Time `json:"periodes,omitempty"`
	Cotisation         []*float64   `json:"cotisation,omitempty"`
	PartPatronale      []*float64   `json:"partPatronale,omitempty"`
	PartSalariale      []*float64   `json:"partSalariale,omitempty"`
	MontantMajorations []*float64   `json:"montantMajorations,omitempty"`
	Effectif           []*int       `json:"effectif,omitempty"`
}

// EtablissementDelai …
type EtablissementDelai struct {
	Action            string    `json:"action"`
	AnneeCreation     int       `json:"anneeCreation"`
	DateCreation      time.Time `json:"dateCreation"`
	DateEcheance      time.Time `json:"dateEcheance"`
	DureeDelai        int       `json:"dureeDelai"`
	Indic6m           string    `json:"indic6m"`
	MontantEcheancier float64   `json:"montantEcheancier"`
	NoContentieux     string    `json:"noContentieux"`
	Stade             string    `json:"stade"`
	Denomination      string    `json:"denomination"`
	NoCompte          string    `json:"noCompte"`
}

// EtablissementAPDemande …
type EtablissementAPDemande struct {
	IDDemande          string    `json:"idDemande"`
	EffectifEntreprise int       `json:"effectifEntreprise"`
	Effectif           int       `json:"effectif"`
	EffectifAutorise   int       `json:"effectifAutorise"`
	DateStatut         time.Time `json:"dateStatut"`
	PeriodeStart       time.Time `json:"debut"`
	PeriodeEnd         time.Time `json:"fin"`
	HTA                float64   `json:"hta"`
	MTA                float64   `json:"mta"`
	MotifRecoursSE     int       `json:"motifRecoursSE"`
	HeureConsomme      float64   `json:"heureConsomme"`
	MontantConsomme    float64   `json:"montantConsomme"`
	EffectifConsomme   int       `json:"effectifConsomme"`
}

// EtablissementAPConso …
type EtablissementAPConso struct {
	IDConso       string    `json:"idConso"`
	HeureConsomme float64   `json:"heureConsomme"`
	Montant       float64   `json:"montant"`
	Effectif      int       `json:"effectif"`
	Periode       time.Time `json:"date"`
}

// EtablissementProcol …
type EtablissementProcol struct {
	DateEffet time.Time `json:"dateEffet"`
	Action    string    `json:"action"`
	Stade     string    `json:"stade"`
}

// EtablissementScore …
type EtablissementScore struct {
	IDListe string    `json:"idListe"`
	Batch   string    `json:"batch"`
	Algo    string    `json:"algo"`
	Periode time.Time `json:"periode"`
	Score   float64   `json:"score"`
	Diff    float64   `json:"diff"`
	Alert   string    `json:"alert"`
}

// Follow type follow pour l'API
type Follow struct {
	Siret  string `json:"siren"`
	UserID string `json:"userId"`
	Query  struct {
	}
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
	EffectifMin  *int     `json:"effectifMin"`
	EffectifMax  *int     `json:"effectifMax"`
	VueFrance    *bool    `json:"vueFrance"`
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
	Siren              string     `json:"siren"`
	Siret              string     `json:"siret"`
	Score              float64    `json:"-"`
	Diff               *float64   `json:"-"`
	RaisonSociale      *string    `json:"raison_sociale"`
	Commune            *string    `json:"commune"`
	LibelleActivite    *string    `json:"libelle_activite"`
	LibelleActiviteN1  *string    `json:"libelle_activite_n1"`
	CodeActivite       *string    `json:"code_activite"`
	Departement        *string    `json:"departement"`
	LibelleDepartement *string    `json:"libelleDepartement"`
	DernierEffectif    *int       `json:"dernier_effectif"`
	HausseUrssaf       *bool      `json:"urssaf"`
	ActivitePartielle  *bool      `json:"activite_partielle"`
	DernierCA          *int       `json:"ca"`
	VariationCA        *float64   `json:"variation_ca"`
	ArreteBilan        *time.Time `json:"arrete_bilan"`
	DernierREXP        *int       `json:"resultat_expl"`
	EtatProcol         *string    `json:"etat_procol"`
	Alert              *string    `json:"alert"`
}

func (e Etablissements) sirensFromQuery() []string {
	var sirens = make(map[string]struct{})
	for _, siret := range e.Query.Sirets {
		sirens[siret[0:9]] = struct{}{}
	}
	for _, siren := range e.Query.Sirens {
		sirens[siren] = struct{}{}
	}
	var list []string
	for k := range sirens {
		list = append(list, k)
	}
	return list
}

func (e *Etablissements) getBatch(roles scope) *pgx.Batch {
	var batch pgx.Batch

	batch.Queue(
		`select 
		et.siret, et.siren, et.siren,	en.raison_sociale, en.statut_juridique,
		et.numero_voie, et.type_voie, et.adresse, et.code_postal, et.commune, et.departement,
		d.libelle, r.libelle,
		et.lattitude, et.longitude,
		et.visite_fce, n.code_n1, n.code_n2, n.code_n3, n.code_n4, n.code_n5
		from etablissement0 et
		inner join v_roles ro on ro.siren = et.siren and ($3 && ro.roles)
		inner join v_naf n on n.code_n5 = et.ape
		inner join departements d on d.code = et.departement
		inner join regions r on d.id_region = r.id
		left join entreprise0 en on en.siren = et.siren
	where 
		(et.siret=any($1) or et.siren=any($2))
		and coalesce($1, $2) is not null
	`, e.Query.Sirets, e.Query.Sirens, roles.zoneGeo())

	batch.Queue(`select en.siren, arrete_bilan_diane, chiffre_affaire, credit_client, resultat_expl, achat_marchandises,
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
		taux_valeur_ajoutee, valeur_ajoutee
		from entreprise_diane0 en
		inner join v_roles ro on ro.siren = en.siren and (ro.roles && $2)
		where en.siren=any($1)`,
		e.sirensFromQuery(),
		roles.zoneGeo())

	batch.Queue(`select en.siren, annee_bdf, delai_fournisseur, financier_court_terme, poids_frng,
		dette_fiscale, frais_financier, taux_marge
		from entreprise_bdf0 en
		inner join v_roles ro on ro.siren = en.siren and (ro.roles && $2) and $2 @> array['bdf']
		where en.siren=any($1)`,
		e.sirensFromQuery(),
		roles.zoneGeo())

	batch.Queue(`select siret, id_conso, heure_consomme, montant, effectif, periode from etablissement_apconso0 e
		inner join v_roles ro on ro.siren = e.siren and (ro.roles) && $1 and $1 @> array['dgefp']
		where e.siret=any($2) or e.siren=any($3) and coalesce($2, $3) is not null;`,
		roles.zoneGeo(), e.Query.Sirets, e.Query.Sirens)

	batch.Queue(`select siret, id_demande, effectif_entreprise, effectif, date_statut, periode_start, 
		periode_end, hta, mta, effectif_autorise, motif_recours_se, heure_consomme, montant_consomme, effectif_consomme
		from etablissement_apdemande0 e
		inner join v_roles ro on ro.siren = e.siren and (ro.roles) && $1 and $1 @> array['dgefp']
		where e.siret=any($2) or e.siren=any($3) and coalesce($2, $3) is not null;`,
		roles.zoneGeo(), e.Query.Sirets, e.Query.Sirens)

	batch.Queue(`select siret, periode, cotisation, part_patronale, part_salariale, montant_majorations, effectif
		from etablissement_periode_urssaf0 e
		inner join v_roles ro on ro.siren = e.siren and  ro.roles && $1 and $1 @> array['urssaf']
		where e.siret=any($2) or e.siren=any($3) and coalesce($2, $3) is not null
		order by e.siret, e.periode;`,
		roles.zoneGeo(), e.Query.Sirets, e.Query.Sirens)

	batch.Queue(`select e.siret, action, annee_creation, date_creation, date_echeance, denomination,
		duree_delai, indic_6m, montant_echeancier, numero_compte, numero_contentieux, stade
		from etablissement_delai0 e
		inner join v_roles ro on ro.siren = e.siren and ro.roles && $1 and $1 @> array['urssaf']
		where e.siret=any($2) or e.siren=any($3) and coalesce($2, $3) is not null;`,
		roles.zoneGeo(), e.Query.Sirets, e.Query.Sirens)

	batch.Queue(`select siret, date_effet, action_procol, stade_procol
		from etablissement_procol0 e
		inner join v_roles ro on ro.siren = e.siren and ro.roles && $1
		where e.siret=any($2) or e.siren=any($3) and coalesce($2, $3) is not null;`,
		roles.zoneGeo(), e.Query.Sirets, e.Query.Sirens)

	batch.Queue(`select siret, libelle_liste, batch, algo, periode, score, diff, alerte
		from score0 e
		inner join v_roles ro on ro.siren = e.siren and ro.roles && $1
		where e.siret=any($2) or e.siren=any($3) and coalesce($2, $3) is not null;`,
		roles.zoneGeo(), e.Query.Sirets, e.Query.Sirens)

	return &batch
}

func (e *Etablissements) loadScore(rows *pgx.Rows) error {
	var scores = make(map[string][]EtablissementScore)
	for (*rows).Next() {
		var sc EtablissementScore
		var siret string
		err := (*rows).Scan(&siret, &sc.IDListe, &sc.Batch, &sc.Algo, &sc.Periode, &sc.Score, &sc.Diff, &sc.Alert)
		if err != nil {
			return err
		}
		scores[siret] = append(scores[siret], sc)
	}
	for k, v := range scores {
		etablissement := e.Etablissements[k]
		etablissement.Scores = v
		e.Etablissements[k] = etablissement
	}
	return nil
}

func (e *Etablissements) loadPeriodeUrssaf(rows *pgx.Rows) error {
	var cotisations = make(map[string][]*float64)
	var partPatronales = make(map[string][]*float64)
	var partSalariales = make(map[string][]*float64)
	var montantMajorations = make(map[string][]*float64)
	var effectifs = make(map[string][]*int)
	var periodes = make(map[string][]*time.Time)
	for (*rows).Next() {
		var pu struct {
			cotisation        *float64
			partPatronale     *float64
			partSalariale     *float64
			montantMajoration *float64
			effectif          *int
			periode           *time.Time
			siret             string
		}

		err := (*rows).Scan(&pu.siret, &pu.periode, &pu.cotisation, &pu.partPatronale, &pu.partSalariale, &pu.montantMajoration,
			&pu.effectif)
		if err != nil {
			return err
		}

		if *pu.partPatronale+*pu.partSalariale+*pu.montantMajoration != 0 || pu.effectif != nil {
			cotisations[pu.siret] = append(cotisations[pu.siret], pu.cotisation)
			partPatronales[pu.siret] = append(partPatronales[pu.siret], pu.partPatronale)
			partSalariales[pu.siret] = append(partSalariales[pu.siret], pu.partSalariale)
			montantMajorations[pu.siret] = append(montantMajorations[pu.siret], pu.montantMajoration)
			effectifs[pu.siret] = append(effectifs[pu.siret], pu.effectif)
			periodes[pu.siret] = append(periodes[pu.siret], pu.periode)
		}
	}

	for k, cotisation := range cotisations {
		etablissement := e.Etablissements[k]
		etablissement.PeriodeUrssaf.Cotisation = cotisation
		etablissement.PeriodeUrssaf.PartPatronale = partPatronales[k]
		etablissement.PeriodeUrssaf.PartSalariale = partSalariales[k]
		etablissement.PeriodeUrssaf.MontantMajorations = montantMajorations[k]
		etablissement.PeriodeUrssaf.Effectif = effectifs[k]
		etablissement.PeriodeUrssaf.Periode = periodes[k]
		e.Etablissements[k] = etablissement
	}
	return nil
}

func (e *Etablissements) loadDelai(rows *pgx.Rows) error {
	var delai = make(map[string][]EtablissementDelai)
	for (*rows).Next() {

		var dl EtablissementDelai
		var siret string
		err := (*rows).Scan(&siret, &dl.Action, &dl.AnneeCreation, &dl.DateCreation, &dl.DateCreation, &dl.Denomination,
			&dl.DureeDelai, &dl.Indic6m, &dl.MontantEcheancier, &dl.NoCompte, &dl.NoContentieux, &dl.Stade)
		if err != nil {
			return err
		}
		delai[siret] = append(delai[siret], dl)

	}
	for k, v := range delai {
		etablissement := e.Etablissements[k]
		etablissement.Delai = v
		e.Etablissements[k] = etablissement
	}
	return nil
}

func (e *Etablissements) loadProcol(rows *pgx.Rows) error {
	var procol = make(map[string][]EtablissementProcol)
	for (*rows).Next() {
		var pc EtablissementProcol
		var siret string
		err := (*rows).Scan(&siret, &pc.DateEffet, &pc.Action, &pc.Stade)
		if err != nil {
			return err
		}
		procol[siret] = append(procol[siret], pc)
	}
	for k, v := range procol {
		etablissement := e.Etablissements[k]
		etablissement.Procol = v
		e.Etablissements[k] = etablissement
	}
	return nil
}

func (e *Etablissements) loadBDF(rows *pgx.Rows) error {
	var bdfs = make(map[string][]bdf)
	for (*rows).Next() {

		var bd bdf
		var siren string
		err := (*rows).Scan(&siren, &bd.Annee, &bd.DelaiFournisseur, &bd.FinancierCourtTerme,
			&bd.PoidsFrng, &bd.DetteFiscale, &bd.FraisFinancier, &bd.TauxMarge,
		)
		if err != nil {
			return err
		}
		bdfs[siren] = append(bdfs[siren], bd)
	}
	for k, v := range bdfs {
		entreprise := e.Entreprises[k]
		entreprise.Bdf = v
		e.Entreprises[k] = entreprise
	}
	return nil
}

func (e *Etablissements) loadAPDemande(rows *pgx.Rows) error {
	var apdemandes = make(map[string][]EtablissementAPDemande)
	for (*rows).Next() {
		var ap EtablissementAPDemande
		var siret string
		err := (*rows).Scan(&siret, &ap.IDDemande, &ap.EffectifEntreprise, &ap.Effectif, &ap.DateStatut, &ap.PeriodeStart,
			&ap.PeriodeEnd, &ap.HTA, &ap.MTA, &ap.EffectifAutorise, &ap.MotifRecoursSE, &ap.HeureConsomme, &ap.MontantConsomme,
			&ap.EffectifConsomme)
		if err != nil {
			return err
		}
		apdemandes[siret] = append(apdemandes[siret], ap)
	}
	for k, v := range apdemandes {
		etablissement := e.Etablissements[k]
		etablissement.APDemande = v
		e.Etablissements[k] = etablissement
	}
	return nil
}

func (e *Etablissements) loadAPConso(rows *pgx.Rows) error {
	var apconsos = make(map[string][]EtablissementAPConso)
	for (*rows).Next() {
		var ap EtablissementAPConso
		var siret string
		err := (*rows).Scan(&siret, &ap.IDConso, &ap.HeureConsomme, &ap.Montant, &ap.Effectif, &ap.Periode)
		if err != nil {
			return err
		}
		apconsos[siret] = append(apconsos[siret], ap)
	}
	for k, v := range apconsos {
		etablissement := e.Etablissements[k]
		etablissement.APConso = v
		e.Etablissements[k] = etablissement
	}
	return nil
}

func (e *Etablissements) loadDiane(rows *pgx.Rows) error {
	var dianes = make(map[string][]diane)
	for (*rows).Next() {

		var di diane
		var siren string
		err := (*rows).Scan(&siren, &di.ArreteBilan, &di.ChiffreAffaire, &di.CreditClient, &di.ResultatExploitation, &di.AchatMarchandises,
			&di.AchatMatieresPremieres, &di.AutonomieFinanciere, &di.AutresAchatsChargesExternes, &di.AutresProduitsChargesReprises,
			&di.CAExportation, &di.CapaciteAutofinancement, &di.CapaciteRemboursement, &di.ChargeExceptionnelle, &di.ChargePersonnel,
			&di.ChargesFinancieres, &di.ConcesBrevEtDroitsSim, &di.Consommation, &di.CouvertureCaBesoinFdr, &di.CouvertureCaFdr,
			&di.CreditFournisseur, &di.DegreImmoCorporelle, &di.DetteFiscaleEtSociale, &di.DotationAmortissement, &di.Endettement,
			&di.EndettementGlobal, &di.EquilibreFinancier, &di.ExcedentBrutDExploitation, &di.Exercice, &di.Exportation,
			&di.FinancementActifCirculant, &di.FraisDeRetD, &di.ImpotBenefice, &di.ImpotsTaxes, &di.IndependanceFinanciere, &di.Interets,
			&di.LiquiditeGenerale, &di.LiquiditeReduite, &di.MargeCommerciale, &di.NombreEtabSecondaire, &di.NombreFiliale, &di.NombreMois,
			&di.OperationsCommun, &di.PartAutofinancement, &di.PartEtat, &di.PartPreteur, &di.PartSalaries, &di.ParticipationSalaries,
			&di.Performance, &di.PoidsBFRExploitation, &di.ProcedureCollective, &di.Production, &di.ProductiviteCapitalFinancier,
			&di.ProductiviteCapitalInvesti, &di.ProductivitePotentielProduction, &di.ProduitExceptionnel, &di.ProduitsFinanciers,
			&di.RendementBrutFondsPropres, &di.RendementCapitauxPropres, &di.RendementRessourcesDurables, &di.RentabiliteEconomique,
			&di.RentabiliteNette, &di.ResultatAvantImpot, &di.RotationStocks, &di.StatutJuridique, &di.SubventionsDExploitation,
			&di.TailleCompoGroupe, &di.TauxDInvestissementProductif, &di.TauxEndettement, &di.TauxInteretFinancier, &di.TauxInteretSurCA,
			&di.TauxValeurAjoutee, &di.ValeurAjoutee,
		)
		if err != nil {
			return err
		}
		dianes[siren] = append(dianes[siren], di)
	}
	for k, v := range dianes {
		entreprise := e.Entreprises[k]
		entreprise.Diane = v
		e.Entreprises[k] = entreprise
	}
	return nil
}

func (e *Etablissements) loadSirene(rows *pgx.Rows) error {
	for (*rows).Next() {
		var et Etablissement
		var en Entreprise
		err := (*rows).Scan(&et.Siret, &et.Siren, &en.Siren,
			&en.Sirene.RaisonSociale, &en.Sirene.StatutJuridique,
			&et.Sirene.NumeroVoie, &et.Sirene.TypeVoie, &et.Sirene.Adresse, &et.Sirene.CodePostal, &et.Sirene.Commune, &et.Sirene.CodeDept,
			&et.Sirene.Dept, &et.Sirene.Region,
			&et.Sirene.Lattitude, &et.Sirene.Longitude, &et.VisiteFCE,
			&et.Sirene.NAF.N1, &et.Sirene.NAF.N2, &et.Sirene.NAF.N3, &et.Sirene.NAF.N4, &et.Sirene.NAF.N5,
		)
		if err != nil {
			return err
		}
		e.Etablissements[et.Siret] = et
		e.Entreprises[en.Siren] = en
	}
	return nil
}

func (e *Etablissements) load(roles scope) error {
	tx, err := db.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Commit(context.Background())
	batch := e.getBatch(roles)
	b := tx.SendBatch(context.Background(), batch)
	if err != nil {
		return err
	}
	defer b.Close()

	e.Etablissements = make(map[string]Etablissement)
	e.Entreprises = make(map[string]Entreprise)

	// etablissement
	rows, err := b.Query()
	if err != nil {
		return err
	}
	err = e.loadSirene(&rows)
	if err != nil {
		return err
	}

	// diane
	rows, err = b.Query()
	if err != nil {
		return err
	}
	err = e.loadDiane(&rows)
	if err != nil {
		return err
	}

	// bdf
	rows, err = b.Query()
	if err != nil {
		return err
	}
	err = e.loadBDF(&rows)
	if err != nil {
		return err
	}

	// apconso
	rows, err = b.Query()
	if err != nil {
		return err
	}
	err = e.loadAPConso(&rows)
	if err != nil {
		return err
	}

	// apdemande
	rows, err = b.Query()
	if err != nil {
		return err
	}
	err = e.loadAPDemande(&rows)
	if err != nil {
		return err
	}

	// periode_urssaf
	rows, err = b.Query()
	if err != nil {
		return err
	}
	err = e.loadPeriodeUrssaf(&rows)
	if err != nil {
		return err
	}

	// delai
	rows, err = b.Query()
	if err != nil {
		return err
	}
	err = e.loadDelai(&rows)
	if err != nil {
		return err
	}

	// procol
	rows, err = b.Query()
	if err != nil {
		return err
	}
	err = e.loadProcol(&rows)
	if err != nil {
		return err
	}

	// scores
	rows, err = b.Query()
	if err != nil {
		return err
	}
	err = e.loadScore(&rows)
	if err != nil {
		return err
	}

	return nil
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

func (etablissement *Etablissement) findEtablissementBySiret() error {

	return errors.New("Non implémenté")
}

func (follow *Follow) findEntreprisesFollowedByUser(db *sql.DB) ([]Entreprise, error) {
	return nil, errors.New("Non implémenté")
}

func (follow *Follow) createEtablissementFollow(db *sql.DB) error {
	// sql
	return nil
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
	sqlScores := `with apdemande as (select siret, true as ap 
		from etablissement_apdemande0
		where periode_start + '1 year'::interval > $6
		group by siret)
	select 
		et.siret,
		et.siren,
		en.raison_sociale, 
		et.commune, 
		d.libelle, 
		d.code,
		s.score,
		s.diff,
		di.chiffre_affaire,
		di.arrete_bilan,
		di.variation_ca,
		di.resultat_expl,
		ef.effectif,
		n.libelle,
		n1.libelle,
		et.ape,
		coalesce(ep.last_procol, 'in_bonis') as last_procol,
		coalesce(ap.ap, false) as activite_partielle,
		case when u.dette[0] > u.dette[1] or u.dette[1] > u.dette[2] then true else false end as hausseUrssaf,
		s.alert
	from score0 s
	inner join v_roles r on r.roles && $1 and r.siren = s.siren
	inner join etablissement0 et on et.siret = s.siret
	inner join entreprise0 en on en.siren = s.siren
	inner join departements d on d.code = et.departement
	inner join naf n on n.code = et.ape
	inner join naf n1 on n.id_n1 = n1.id
	left join v_last_effectif ef on ef.siret = s.siret
	left join v_hausse_urssaf u on u.siret = s.siret
	left join apdemande ap on ap.siret = s.siret
	left join v_last_procol ep on ep.siret = s.siret
	left join v_diane_variation_ca di on di.siren = s.siren
	where s.libelle_liste = $2
	and (coalesce(ep.last_procol, 'in_bonis')=any($3) or $3 is null)
	and (et.departement=any($4) or $4 is null)
	and (n1.code=any($5) or $5 is null)
	and (ef.effectif >= $7 or $7 is null)
	and (ef.effectif <= $8 or $8 is null)
	and (et.departement=any($1) or $9 = true)
	and s.alert != 'Pas d''alerte'
	order by s.score desc;`
	rows, err := db.Query(context.Background(), sqlScores, roles.zoneGeo(), liste.ID, liste.Query.EtatsProcol,
		liste.Query.Departements, liste.Query.Activites, time.Now(), liste.Query.EffectifMin, liste.Query.EffectifMax, liste.Query.VueFrance)
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
			&score.LibelleDepartement,
			&score.Departement,
			&score.Score,
			&score.Diff,
			&score.DernierCA,
			&score.ArreteBilan,
			&score.VariationCA,
			&score.DernierREXP,
			&score.DernierEffectif,
			&score.LibelleActivite,
			&score.LibelleActiviteN1,
			&score.CodeActivite,
			&score.EtatProcol,
			&score.ActivitePartielle,
			&score.HausseUrssaf,
			&score.Alert,
		)
		if err != nil {
			return err
		}
		scores = append(scores, score)
	}

	liste.Scores = scores
	return nil
}

func getSiegeFromSiren(siren string) (string, error) {
	sqlSiege := `select siret from etablissement0
	where siren = $1 and siege`
	var siret string
	err := db.QueryRow(context.Background(), sqlSiege, siren).Scan(&siret)
	return siret, err
}
