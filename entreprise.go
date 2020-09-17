package main

import (
	"context"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	pgx "github.com/jackc/pgx/v4"
	"github.com/spf13/viper"
)

// Entreprise type entreprise pour l'API
type Entreprise struct {
	Siren  string `json:"siren"`
	Sirene struct {
		RaisonSociale   string    `json:"raisonSociale"`
		StatutJuridique string    `json:"statutJuridique"`
		Prenom1         string    `json:"prenom1,omitempty"`
		Prenom2         string    `json:"prenom2,omitempty"`
		Prenom3         string    `json:"prenom3,omitempty"`
		Prenom4         string    `json:"prenom4,omitempty"`
		Nom             string    `json:"nom,omitempty"`
		NomUsage        string    `json:"nomUsage,omitempty"`
		Creation        time.Time `json:"creation,omitempty"`
	}
	Diane                 []diane                `json:"diane"`
	Bdf                   []bdf                  `json:"-"`
	EtablissementsSummary []EtablissementSummary `json:"etablissementsSummary,omitempty"`
	Etablissements        []Etablissement        `json:"etablissements,omitEmpty"`
}

// EtablissementSummary …
type EtablissementSummary struct {
	Siret              *string `json:"siret,omitempty"`
	Departement        *string `json:"departement"`
	LibelleDepartement *string `json:"libelleDepartement"`
	RaisonSociale      *string `json:"raisonSociale"`
	Commune            *string `json:"commune"`
	Region             *string `json:"region"`
	CodeActivite       *string `json:"codeActivite"`
	CodeSecteur        *string `json:"codeSecteur"`
	Activite           *string `json:"activite"`
	Secteur            *string `json:"secteur"`
	DernierEffectif    *int    `json:"dernierEffectif,omitempty"`
	Visible            *bool   `json:"visible,omitempty"`
	Inzone             *bool   `json:"zone,omitempty"`
	Alert              *bool   `json:"alert,omitempty"`
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
		Latitude             float64 `json:"latitude"`
		Longitude            float64 `json:"longitude"`
		NumeroVoie           string  `json:"numeroVoie"`
		TypeVoie             string  `json:"typeVoie"`
		Voie                 string  `json:"voie"`
		IndiceRepetition     string  `json:"indiceRepetition"`
		DistributionSpeciale string  `json:"distributionSpeciale"`
		Commune              string  `json:"commune"`
		Dept                 string  `json:"dept"`
		CodeDept             string  `json:"codeDept"`
		CodePostal           string  `json:"codePostal"`
		Region               string  `json:"region"`
		NAF                  struct {
			LibelleActivite string `json:"libelleActivite"`
			LibelleSecteur  string `json:"libelleSecteur"`
			CodeSecteur     string `json:"codeSecteur"`
			CodeActivite    string `json:"codeActivite"`
			LibelleN2       string `json:"libelleN2"`
			LibelleN3       string `json:"libelleN3"`
			LibelleN4       string `json:"libelleN4"`
			NomenActivite   string `json:"nomenActivite"`
		} `json:"naf"`
	} `json:"sirene"`
	PeriodeUrssaf EtablissementPeriodeUrssaf `json:"periodeUrssaf,omitempty"`
	Delai         []EtablissementDelai       `json:"delai,omitempty"`
	APDemande     []EtablissementAPDemande   `json:"apDemande,omitempty"`
	APConso       []EtablissementAPConso     `json:"apConso,omitempty"`
	Procol        []EtablissementProcol      `json:"procol,omitempty"`
	Scores        []EtablissementScore       `json:"scores,omitempty"`
	Followed      bool                       `json:"followed"`
	Visible       bool                       `json:"visible"`
	InZone        bool                       `json:"inZone"`
	Alert         bool                       `json:"alert,omitempty"`
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

func getEntreprise(c *gin.Context) {
	roles := scopeFromContext(c)
	siren := c.Param("siren")
	username := c.GetString("username")
	siret, err := getSiegeFromSiren(siren)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}
	var etablissements Etablissements
	etablissements.Query.Sirets = []string{siret}
	err = etablissements.load(roles, username)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}
	entreprise, ok := etablissements.Entreprises[siren]
	if !ok {
		c.JSON(404, "ressource non disponible")
		return
	}
	for _, v := range etablissements.Etablissements {
		entreprise.Etablissements = append(entreprise.Etablissements, v)
	}
	c.JSON(200, entreprise)
}

func getEtablissement(c *gin.Context) {
	roles := scopeFromContext(c)
	siret := c.Param("siret")
	username := c.GetString("username")
	var etablissements Etablissements
	etablissements.Query.Sirets = []string{siret}
	err := etablissements.load(roles, username)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}
	result, ok := etablissements.Etablissements[siret]
	if !ok {
		c.JSON(404, "etablissement non disponible")
		return
	}
	entreprise := etablissements.Entreprises[siret[0:9]]
	result.Entreprise = &entreprise

	if err != nil {
		c.JSON(500, err.Error())
		return
	}
	c.JSON(200, result)
}

func getEntrepriseEtablissements(c *gin.Context) {
	roles := scopeFromContext(c)
	siren := c.Param("siren")
	username := c.GetString("username")
	var etablissements Etablissements
	etablissements.Query.Sirens = []string{siren}
	err := etablissements.load(roles, username)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}
	entreprise, ok := etablissements.Entreprises[siren]
	if !ok {
		c.JSON(404, "ressource ênon disponible")
		return
	}

	for _, v := range etablissements.Etablissements {
		entreprise.Etablissements = append(entreprise.Etablissements, v)
	}

	c.JSON(200, entreprise)
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

func (e *Etablissements) getBatch(roles scope, username string) *pgx.Batch {
	var batch pgx.Batch

	batch.Queue(
		`select 
		et.siret, et.siren, et.siren,	en.raison_sociale, en.statut_juridique,
		et.numero_voie, et.type_voie, et.voie, et.indice_repetition, et.distribution_speciale, 
		et.code_postal, et.commune, et.departement,
		d.libelle, r.libelle, et.latitude, et.longitude, et.visite_fce, 
		n.libelle_n1, n.code_n1, n.libelle_n5, n.code_n5,
		n.libelle_n2, n.libelle_n3, n.libelle_n4,
		coalesce(et.nomen_activite, 'NAFRev2'),
		f.id is not null as followed,
		ro.roles && $4 as visible,
		s.siret is not null as alert,
		en.prenom1, en.prenom2, en.prenom3, en.prenom4, en.nom, en.nom_usage, en.creation
		from etablissement0 et
		inner join v_naf n on n.code_n5 = et.code_activite
		inner join departements d on d.code = et.departement
		inner join regions r on d.id_region = r.id
		inner join v_roles ro on ro.siren = et.siren
		left join etablissement_follow f on f.siret = et.siret and f.active = true and f.username = $3
		left join entreprise0 en on en.siren = et.siren
		left join v_score s on s.siret = et.siret
		where 
		(et.siret=any($1) or et.siren=any($2))
		and coalesce($1, $2) is not null;
	`, e.Query.Sirets, e.Query.Sirens, username, roles.zoneGeo())

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
		where en.siren=any($1)
		order by en.arrete_bilan_diane;`,
		e.sirensFromQuery(),
	)

	batch.Queue(`select en.siren, annee_bdf, arrete_bilan_bdf, delai_fournisseur, financier_court_terme, poids_frng,
		dette_fiscale, frais_financier, taux_marge
		from entreprise_bdf0 en
		left join (select distinct siren from v_score) s on s.siren = en.siren
		left join (select distinct siren from etablissement_follow where username=$3 and active) f on f.siren = en.siren
		inner join v_roles ro on ro.siren = en.siren and $2 @> array['bdf'] and (s.siren is not null and (ro.roles && $2) or f.siren is not null) 
		where en.siren=any($1)
		order by arrete_bilan_bdf;`,
		e.sirensFromQuery(),
		roles.zoneGeo(),
		username,
	)

	batch.Queue(`select e.siret, libelle_liste, batch, algo, periode, score, diff, alert
		from score0 e
		inner join v_roles ro on ro.siren = e.siren and ro.roles && $1
		left join v_score s on s.siret = e.siret
		left join etablissement_follow f on f.siret = e.siret and f.active and f.username = $4
		where (e.siret=any($2) or e.siren=any($3)) and coalesce($2, $3) is not null and coalesce(s.siret, f.siret) is not null
		order by siret, batch desc, score desc;`,
		roles.zoneGeo(), e.Query.Sirets, e.Query.Sirens, username)

	batch.Queue(`select e.siret, id_conso, heure_consomme, montant, effectif, periode
		from etablissement_apconso0 e
		left join v_score s on s.siret = e.siret
		left join etablissement_follow f on f.siret = e.siret and f.username = $4 and f.active
		inner join v_roles ro on ro.siren = e.siren and (ro.roles && $1 and s.siren is not null or f.id is not null) and $1 @> array['dgefp']
		where e.siret=any($2) or e.siren=any($3) and coalesce($2, $3) is not null
		order by siret, periode;`,
		roles.zoneGeo(), e.Query.Sirets, e.Query.Sirens, username)

	batch.Queue(`select e.siret, id_demande, effectif_entreprise, effectif, date_statut, periode_start, 
		periode_end, hta, mta, effectif_autorise, motif_recours_se, heure_consomme, montant_consomme, effectif_consomme
		from etablissement_apdemande0 e
		left join v_score s on s.siret = e.siret
		left join etablissement_follow f on f.siret = e.siret and f.username = $4 and f.active
		inner join v_roles ro on ro.siren = e.siren and (ro.roles && $1 and s.siren is not null or f.id is not null) and $1 @> array['dgefp']
		where e.siret=any($2) or e.siren=any($3) and coalesce($2, $3) is not null
		order by siret, periode_start;`,
		roles.zoneGeo(), e.Query.Sirets, e.Query.Sirens, username)

	batch.Queue(`select e.siret, e.periode, 
		case when 'urssaf' = any($1) and (ro.roles && $1 and s.siret is not null or f.id is not null) and cotisation != 0 then 
			cotisation else null
		end, 
		case when 'urssaf' = any($1) and (ro.roles && $1 and s.siret is not null or f.id is not null) then part_patronale else null end, 
		case when 'urssaf' = any($1) and (ro.roles && $1 and s.siret is not null or f.id is not null) then part_salariale else null end, 
		case when 'urssaf' = any($1) and (ro.roles && $1 and s.siret is not null or f.id is not null) then montant_majorations else null end, 
		effectif
		from etablissement_periode_urssaf0 e
		left join v_score s on s.siret = e.siret
		left join etablissement_follow f on f.siret = e.siret and f.username = $4 and f.active
		inner join v_roles ro on ro.siren = e.siren
		where e.siret=any($2) or e.siren=any($3) and coalesce($2, $3) is not null
		order by siret, periode;`,
		roles.zoneGeo(), e.Query.Sirets, e.Query.Sirens, username)

	batch.Queue(`select e.siret, action, annee_creation, date_creation, date_echeance, denomination,
		duree_delai, indic_6m, montant_echeancier, numero_compte, numero_contentieux, stade
		from etablissement_delai0 e
		left join v_score s on s.siret = e.siret
		left join etablissement_follow f on f.siret = e.siret and f.username = $4 and f.active
		inner join v_roles ro on ro.siren = e.siren and (ro.roles && $1 and s.siret is not null or f.id is not null) and $1 @> array['urssaf']
		where e.siret=any($2) or e.siren=any($3) and coalesce($2, $3) is not null
		order by e.siret, date_creation;`,
		roles.zoneGeo(), e.Query.Sirets, e.Query.Sirens, username)

	batch.Queue(`select e.siret, date_effet, action_procol, stade_procol
		from etablissement_procol0 e
		where e.siret=any($1) or e.siren=any($2) and coalesce($1, $2) is not null
		order by e.siret, date_effet;`,
		e.Query.Sirets, e.Query.Sirens)

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

func (e *Etablissements) loadPeriodeUrssaf(rows *pgx.Rows, roles scope) error {
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

		if sumPFloats(pu.cotisation, pu.partPatronale, pu.partSalariale, pu.montantMajoration) != 0 || pu.effectif != nil {
			effectifs[pu.siret] = append(effectifs[pu.siret], pu.effectif)
			periodes[pu.siret] = append(periodes[pu.siret], pu.periode)
			if roles.containsRole("urssaf") && (e.Etablissements[pu.siret].Visible && e.Etablissements[pu.siret].Alert || e.Etablissements[pu.siret].Followed) {
				cotisations[pu.siret] = append(cotisations[pu.siret], pu.cotisation)
				partPatronales[pu.siret] = append(partPatronales[pu.siret], pu.partPatronale)
				partSalariales[pu.siret] = append(partSalariales[pu.siret], pu.partSalariale)
				montantMajorations[pu.siret] = append(montantMajorations[pu.siret], pu.montantMajoration)
			}
		}
	}

	for k, periode := range periodes {
		etablissement := e.Etablissements[k]
		etablissement.PeriodeUrssaf.Cotisation = cotisations[k]
		etablissement.PeriodeUrssaf.PartPatronale = partPatronales[k]
		etablissement.PeriodeUrssaf.PartSalariale = partSalariales[k]
		etablissement.PeriodeUrssaf.MontantMajorations = montantMajorations[k]
		etablissement.PeriodeUrssaf.Effectif = effectifs[k]
		etablissement.PeriodeUrssaf.Periode = periode
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
		err := (*rows).Scan(&siren, &bd.Annee, &bd.ArreteBilan, &bd.DelaiFournisseur, &bd.FinancierCourtTerme,
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

func (e *Etablissements) loadAPDemande(rows *pgx.Rows, roles scope) error {
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
		if len(e.Etablissements[siret].Scores) > 0 && roles.containsRole("dgefp") {
			apdemandes[siret] = append(apdemandes[siret], ap)
		}
	}
	for k, v := range apdemandes {
		etablissement := e.Etablissements[k]
		etablissement.APDemande = v
		e.Etablissements[k] = etablissement
	}
	return nil
}

func (e *Etablissements) loadAPConso(rows *pgx.Rows, roles scope) error {
	var apconsos = make(map[string][]EtablissementAPConso)
	for (*rows).Next() {
		var ap EtablissementAPConso
		var siret string
		err := (*rows).Scan(&siret, &ap.IDConso, &ap.HeureConsomme, &ap.Montant, &ap.Effectif, &ap.Periode)
		if err != nil {
			return err
		}
		if len(e.Etablissements[siret].Scores) > 0 && roles.containsRole("dgefp") {
			apconsos[siret] = append(apconsos[siret], ap)
		}
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
			&et.Sirene.NumeroVoie, &et.Sirene.TypeVoie,
			&et.Sirene.Voie, &et.Sirene.IndiceRepetition, &et.Sirene.DistributionSpeciale,
			&et.Sirene.CodePostal, &et.Sirene.Commune, &et.Sirene.CodeDept,
			&et.Sirene.Dept, &et.Sirene.Region,
			&et.Sirene.Latitude, &et.Sirene.Longitude, &et.VisiteFCE,
			&et.Sirene.NAF.LibelleSecteur,
			&et.Sirene.NAF.CodeSecteur,
			&et.Sirene.NAF.LibelleActivite,
			&et.Sirene.NAF.CodeActivite,
			&et.Sirene.NAF.LibelleN2,
			&et.Sirene.NAF.LibelleN3,
			&et.Sirene.NAF.LibelleN4,
			&et.Sirene.NAF.NomenActivite,
			&et.Followed,
			&et.Visible,
			&et.Alert,
			&en.Sirene.Prenom1,
			&en.Sirene.Prenom2,
			&en.Sirene.Prenom3,
			&en.Sirene.Prenom4,
			&en.Sirene.Nom,
			&en.Sirene.NomUsage,
			&en.Sirene.Creation,
		)

		if err != nil {
			return err
		}
		e.Etablissements[et.Siret] = et
		e.Entreprises[en.Siren] = en
	}
	return nil
}

func (e *Etablissements) load(roles scope, username string) error {
	tx, err := db.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Commit(context.Background())
	batch := e.getBatch(roles, username)
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

	// scores
	rows, err = b.Query()
	if err != nil {
		return err
	}
	err = e.loadScore(&rows)
	if err != nil {
		return err
	}

	// apconso
	rows, err = b.Query()
	if err != nil {
		return err
	}
	err = e.loadAPConso(&rows, roles)
	if err != nil {
		return err
	}

	// apdemande
	rows, err = b.Query()
	if err != nil {
		return err
	}
	err = e.loadAPDemande(&rows, roles)
	if err != nil {
		return err
	}

	// periode_urssaf
	rows, err = b.Query()
	if err != nil {
		return err
	}
	err = e.loadPeriodeUrssaf(&rows, roles)
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

	return nil
}

func getSiegeFromSiren(siren string) (string, error) {
	sqlSiege := `select siret from etablissement0
	where siren = $1 and siege`
	var siret string
	err := db.QueryRow(context.Background(), sqlSiege, siren).Scan(&siret)
	return siret, err
}

type searchParams struct {
	search      string
	page        int
	ignoreRoles bool
	ignoreZone  bool
	roles       scope
	username    string
}

type searchResult struct {
	From    int     `json:"from"`
	To      int     `json:"to"`
	Total   int     `json:"total"`
	PageMax int     `json:"pageMax"`
	Page    int     `json:"page"`
	Results []Score `json:"results"`
}

func searchEtablissementHandler(c *gin.Context) {
	var params searchParams
	var err error

	params.username = c.GetString("username")

	if search := c.Param("search"); len(search) >= 3 {
		params.search = search
	} else {
		c.JSON(400, "search string length < 3")
	}

	if page, ok := c.GetQuery("page"); ok {
		params.page, err = strconv.Atoi(page)
		if err != nil {
			c.JSON(400, "page has to be integer >= 0")
			return
		}
	}

	if ignoreRoles, ok := c.GetQuery("ignoreroles"); ok {
		params.ignoreRoles, err = strconv.ParseBool(ignoreRoles)
		if err != nil {
			c.JSON(400, "france is either `true` or `false`")
			return
		}
	}

	if ignoreZone, ok := c.GetQuery("ignorezone"); ok {
		params.ignoreZone, err = strconv.ParseBool(ignoreZone)
		if err != nil {
			c.JSON(400, "roles is either `true` or `false`")
			return
		}
	}

	params.roles = scopeFromContext(c)

	result, Jerr := searchEtablissement(params)
	if Jerr != nil {
		c.JSON(Jerr.Code(), Jerr.Error())
		return
	}
	c.JSON(200, result)

}

func searchEtablissement(params searchParams) (searchResult, Jerror) {
	sqlSearch := `select 
		et.siret,
		et.siren,
		en.raison_sociale, 
		et.commune, 
		d.libelle, 
		d.code,
		case when (r.roles && $1 and vs.siret is not null) or f.id is not null then s.score else null end,
		case when (r.roles && $1 and vs.siret is not null) or f.id is not null then s.diff else null end,
		di.chiffre_affaire,
		di.arrete_bilan,
		di.variation_ca,
		di.resultat_expl,
		ef.effectif,
		n.libelle,
		n1.libelle,
		et.code_activite,
		coalesce(ep.last_procol, 'in_bonis') as last_procol,
		case when 'dgefp' = any($1) and ((r.roles && $1 and vs.siret is not null) or f.id is not null) then coalesce(ap.ap, false) else null end as activite_partielle ,
		case when 'urssaf' = any($1) and ((r.roles && $1 and vs.siret is not null) or f.id is not null) then 
			case when u.dette[0] > u.dette[1] or u.dette[1] > u.dette[2] then true else false end 
		else null end as hausseUrssaf,
		case when 'detection' = any($1) and ((r.roles && $1 and vs.siret is not null) or f.id is not null) then s.alert else null end,
		count(*) over (),
		coalesce(r.roles && $1 and vs.siret is not null, false) as visible,
		coalesce(et.departement = any($2), false) as in_zone,
		f.id is not null as followed
		from etablissement0 et
		inner join v_roles r on et.siren = r.siren
		inner join entreprise0 en on en.siren = r.siren
		inner join departements d on d.code = et.departement
		inner join naf n on n.code = et.code_activite
		inner join naf n1 on n.id_n1 = n1.id 
		left join score s on et.siret = s.siret
		left join v_score vs on vs.siret = et.siret
		left join etablissement_follow f on f.siret = et.siret and f.active and f.username = $10
		left join v_last_effectif ef on ef.siret = et.siret
		left join v_hausse_urssaf u on u.siret = et.siret
		left join v_apdemande ap on ap.siret = et.siret
		left join v_last_procol ep on ep.siret = et.siret
		left join v_diane_variation_ca di on di.siren = s.siren
		where (et.siret ilike $6 or en.raison_sociale ilike $7)
		and (r.roles && $1 or $8)
		and (et.departement=any($2) or $9)
		and coalesce(s.libelle_liste, $5) = $5
		order by en.raison_sociale, siret
		limit $3 offset $4
		;`

	liste, err := findAllListes()
	if err != nil {
		return searchResult{}, errorToJSON(500, err)
	}
	zoneGeo := params.roles.zoneGeo()
	limit := viper.GetInt("searchPageLength")

	rows, err := db.Query(context.Background(), sqlSearch,

		zoneGeo,
		zoneGeo,
		limit,
		params.page*limit,
		liste[0].ID,
		params.search+"%",
		"%"+params.search+"%",
		params.ignoreRoles,
		params.ignoreZone,
		params.username,
	)

	if err != nil {
		return searchResult{}, errorToJSON(500, err)
	}

	var search searchResult
	for rows.Next() {
		var r Score
		err := rows.Scan(&r.Siret,
			&r.Siren,
			&r.RaisonSociale,
			&r.Commune,
			&r.LibelleDepartement,
			&r.Departement,
			&r.Score,
			&r.Diff,
			&r.DernierCA,
			&r.ArreteBilan,
			&r.VariationCA,
			&r.DernierREXP,
			&r.DernierEffectif,
			&r.LibelleActivite,
			&r.LibelleActiviteN1,
			&r.CodeActivite,
			&r.EtatProcol,
			&r.ActivitePartielle,
			&r.HausseUrssaf,
			&r.Alert,
			&search.Total,
			&r.Visible,
			&r.InZone,
			&r.Followed,
		)
		if err != nil {
			return searchResult{}, errorToJSON(500, err)
		}
		search.Results = append(search.Results, r)
	}

	if viper.GetInt("searchPageLength")*params.page >= search.Total {
		return searchResult{}, newJSONerror(204, "empty page")
	}

	search.From = limit*params.page + 1
	search.To = limit*params.page + len(search.Results)
	search.Page = params.page
	search.PageMax = (search.Total - 1) / limit

	return search, nil
}

func getEntrepriseViewers(c *gin.Context) {
	sqlViewers := `select username, firstname, lastname from v_roles r
		inner join users u on u.roles && r.roles
		where siren = $1`

	siren := c.Param("siren")
	rows, err := db.Query(context.Background(), sqlViewers, siren)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}
	var users []user
	for rows.Next() {
		var u user
		err := rows.Scan(&u.Username, &u.FirstName, &u.LastName)
		if err != nil {
			c.JSON(500, err.Error())
			return
		}
		users = append(users, u)
	}
	if len(users) == 0 {
		c.JSON(204, "")
		return
	}
	c.JSON(200, users)
}

func getEtablissementViewers(c *gin.Context) {
	sqlViewers := `select username, firstname, lastname from v_roles r
		inner join users u on u.roles && r.roles
		where siren = $1`

	siren := c.Param("siret")[0:9]
	rows, err := db.Query(context.Background(), sqlViewers, siren)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}
	var users []user
	for rows.Next() {
		var u user
		err := rows.Scan(&u.Username, &u.FirstName, &u.LastName)
		if err != nil {
			c.JSON(500, err.Error())
			return
		}
		users = append(users, u)
	}
	if len(users) == 0 {
		c.JSON(204, "")
		return
	}
	c.JSON(200, users)
}

func sumPFloats(floats ...*float64) float64 {
	var total float64
	for _, f := range floats {
		if f == nil {
			continue
		} else {
			total += *f
		}
	}
	return total
}
