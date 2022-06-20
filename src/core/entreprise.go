package core

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
)

// Entreprise type entreprise pour l'API
type Entreprise struct {
	Siren  string `json:"siren"`
	Sirene struct {
		RaisonSociale     *string   `json:"raisonSociale"`
		StatutJuridique   *string   `json:"statutJuridique"`
		StatutJuridiqueN1 *string   `json:"statutJuridiqueN1"`
		StatutJuridiqueN2 *string   `json:"statutJuridiqueN2"`
		Prenom1           *string   `json:"prenom1,omitempty"`
		Prenom2           *string   `json:"prenom2,omitempty"`
		Prenom3           *string   `json:"prenom3,omitempty"`
		Prenom4           *string   `json:"prenom4,omitempty"`
		Nom               *string   `json:"nom,omitempty"`
		NomUsage          *string   `json:"nomUsage,omitempty"`
		Creation          time.Time `json:"creation,omitempty"`
	}
	Paydex                *Paydex         `json:"paydex,omitempty"`
	Diane                 []diane         `json:"diane"`
	Bdf                   []bdf           `json:"-"`
	EtablissementsSummary []Summary       `json:"etablissementsSummary,omitempty"`
	Etablissements        []Etablissement `json:"etablissements,omitempty"`
	Groupe                *ellisphere     `json:"groupe,omitempty"`
	PGEActif              *bool           `json:"pge,omitempty"`
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
	Siege              *bool   `json:"siege,omitempty"`
}

// Paydex est l'index des retards de reglements de D&B
type Paydex struct {
	DateValeur []time.Time `json:"date_valeur"`
	NBJours    []int       `json:"nb_jours"`
}

// Etablissements Collection d'établissements
type Etablissements struct {
	Query struct {
		Sirets []string `json:"sirets"`
		Sirens []string `json:"sirens"`
	} `json:"-"`
	Etablissements map[string]Etablissement `json:"etablissements,omitempty"`
	Entreprises    map[string]Entreprise    `json:"entreprises,omitempty"`
}

// Etablissement type établissement pour l'API
type Etablissement struct {
	Siren      string      `json:"siren"`
	Siret      string      `json:"siret"`
	Entreprise *Entreprise `json:"entreprise,omitempty"`
	VisiteFCE  *bool       `json:"visiteFCE"`
	Siege      *bool       `json:"siege,omitempty"`
	Sirene     struct {
		ComplementAdresse    *string    `json:"-"`
		NumVoie              *string    `json:"-"`
		IndRep               *string    `json:"-"`
		TypeVoie             *string    `json:"-"`
		Voie                 *string    `json:"-"`
		Commune              *string    `json:"-"`
		CommuneEtranger      *string    `json:"-"`
		DistributionSpeciale *string    `json:"-"`
		CodeCommune          *string    `json:"-"`
		CodeCedex            *string    `json:"-"`
		Cedex                *string    `json:"-"`
		CodePaysEtranger     *string    `json:"-"`
		PaysEtranger         *string    `json:"-"`
		Adresse              *string    `json:"adresse"`
		CodeDepartement      *string    `json:"codeDepartement"`
		Departement          *string    `json:"departement"`
		Creation             *time.Time `json:"creation"`
		CodePostal           *string    `json:"codePostal"`
		Region               *string    `json:"region"`
		Latitude             *float64   `json:"latitude"`
		Longitude            *float64   `json:"longitude"`
		NAF                  struct {
			LibelleActivite *string `json:"libelleActivite"`
			LibelleSecteur  *string `json:"libelleSecteur"`
			CodeSecteur     *string `json:"codeSecteur"`
			CodeActivite    *string `json:"codeActivite"`
			LibelleN2       *string `json:"libelleN2"`
			LibelleN3       *string `json:"libelleN3"`
			LibelleN4       *string `json:"libelleN4"`
			NomenActivite   *string `json:"nomenActivite"`
		} `json:"naf"`
	} `json:"sirene"`
	PeriodeUrssaf      EtablissementPeriodeUrssaf `json:"periodeUrssaf,omitempty"`
	Delai              []EtablissementDelai       `json:"delai,omitempty"`
	APDemande          []EtablissementAPDemande   `json:"apDemande,omitempty"`
	APConso            []EtablissementAPConso     `json:"apConso,omitempty"`
	Procol             []EtablissementProcol      `json:"procol,omitempty"`
	Scores             []EtablissementScore       `json:"scores,omitempty"`
	Followed           bool                       `json:"followed"`
	FollowedEntreprise bool                       `json:"followedEntreprise"`
	Visible            bool                       `json:"visible"`
	InZone             bool                       `json:"inZone"`
	Alert              bool                       `json:"alert,omitempty"`
	TerrInd            *EtablissementTerrInd      `json:"territoireIndustrie,omitempty"`
	PermScore          bool                       `json:"permScore"`
	PermUrssaf         bool                       `json:"permUrssaf"`
	PermDGEFP          bool                       `json:"permDGEFP"`
	PermBDF            bool                       `json:"permBDF"`
	PermPGE            bool                       `json:"permPGE"`
}

// EtablissementTerrInd …
type EtablissementTerrInd struct {
	Code    string `json:"code"`
	Libelle string `json:"libelle"`
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

// EtablissementScoreExplSelection …
type EtablissementScoreExplSelection struct {
	SelectConcerning [][2]string `json:"selectConcerning"`
	SelectReassuring [][2]string `json:"-"`
}

// EtablissementScore …
type EtablissementScore struct {
	IDListe               string                           `json:"idListe"`
	Batch                 string                           `json:"batch"`
	Algo                  string                           `json:"algo"`
	Periode               time.Time                        `json:"periode"`
	Score                 float64                          `json:"score"`
	Diff                  float64                          `json:"diff"`
	Alert                 string                           `json:"alert"`
	MacroRadar            map[string]float64               `json:"macroRadar,omitempty"`
	ExplSelection         *EtablissementScoreExplSelection `json:"explSelection,omitempty"`
	MacroExpl             map[string]float64               `json:"-"`
	MicroExpl             map[string]float64               `json:"-"`
	AlertPreRedressements string                           `json:"alertPreRedressements"`
	Redressements         []string                         `json:"redressements"`
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

func (e *Etablissements) intoBatch(roles scope, username string) *pgx.Batch {
	var batch pgx.Batch
	listes, _ := findAllListes()
	lastListe := listes[0].ID

	batch.Queue(
		`select 
		et.siret, et.siren, et.siren,	en.raison_sociale, j.libelle, j2.libelle, j1.libelle,
		et.complement_adresse, et.numero_voie, et.indice_repetition, et.type_voie, et.voie,  
		et.commune, et.commune_etranger, et.distribution_speciale, et.code_commune,
		et.code_cedex, et.cedex, et.code_pays_etranger, et.pays_etranger,
		et.code_postal, et.departement, d.libelle, r.libelle, 
		coalesce(et.nomen_activite, 'NAFRev2'), et.creation,	et.latitude, et.longitude, 
		et.visite_fce, n.libelle_n1, n.code_n1, n.libelle_n5, et.code_activite, n.libelle_n2,
		n.libelle_n3, n.libelle_n4,	
		f.id is not null as followed,
		followed as followed_entreprise, visible,
		s.siren is not null as alert,	en.prenom1, en.prenom2, en.prenom3, en.prenom4, 
		en.nom, en.nom_usage, en.creation, et.siege,
		coalesce(g.code, ''), 
		coalesce(g.refid, ''), 
		coalesce(g.raison_sociale, ''), 
		coalesce(g.adresse, ''), 
		coalesce(g.personne_pou_m, ''), 
		coalesce(g.niveau_detention, 0), 
		coalesce(g.part_financiere, 0), 
		coalesce(g.code_filiere, ''),
		coalesce(g.refid_filiere, ''),
		coalesce(g.personne_pou_m_filiere, ''),
		coalesce(ti.code_terrind,''), coalesce(ti.libelle_terrind,''),
		p.score as permScore,
		p.urssaf as permUrssaf,
		p.dgefp as permDGEFP,
		p.bdf as permBDF,
		p.in_zone as in_zone
		from etablissement0 et
		inner join f_etablissement_permissions($1, $2) p on p.id = et.id
		inner join departements d on d.code = et.departement
		inner join regions r on d.id_region = r.id
		inner join v_roles ro on ro.siren = et.siren
		left join v_naf n on n.code_n5 = et.code_activite
		left join etablissement_follow f on f.siret = et.siret and f.active and f.username = $2
		left join entreprise0 en on en.siren = et.siren
		left join categorie_juridique j on j.code = en.statut_juridique
		left join categorie_juridique j2 on substring(j.code from 0 for 3) = j2.code
		left join categorie_juridique j1 on substring(j.code from 0 for 2) = j1.code
		left join v_alert_entreprise s on s.siren = et.siren
		left join entreprise_ellisphere0 g on g.siren = et.siren
		left join terrind ti on ti.code_commune = et.code_commune
		where (et.siret=any($3) or et.siren=any($4));
	`, roles.zoneGeo(), username, e.Query.Sirets, e.Query.Sirens)

	batch.Queue(`select siren, arrete_bilan_diane, achat_marchandises, achat_matieres_premieres, autonomie_financiere, 
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
				taux_marge_commerciale,	taux_valeur_ajoutee, valeur_ajoutee
			from entreprise_diane0 en
			where en.siren=any($1)
			order by en.arrete_bilan_diane;`,
		e.sirensFromQuery())

	// A remettre en service lorsque nous aurons des données BDF
	// batch.Queue(`select en.siren, annee_bdf, arrete_bilan_bdf, delai_fournisseur, financier_court_terme, poids_frng,
	// 	dette_fiscale, frais_financier, taux_marge
	// 	from entreprise_bdf0 en
	// 	left join v_alert_entreprise s on s.siren = en.siren
	// 	left join (select distinct siren from etablissement_follow where username=$3 and active) f on f.siren = en.siren
	// 	inner join v_roles ro on ro.siren = en.siren and $2 @> array['bdf'] and (s.siren is not null and (ro.roles && $2) or f.siren is not null)
	// 	where en.siren=any($1)
	// 	order by arrete_bilan_bdf;`,
	// 	e.sirensFromQuery(),
	// 	roles.zoneGeo(),
	// 	username,
	// )

	batch.Queue(`select s.siret, s.libelle_liste, s.batch, s.algo, s.periode, s.score, s.diff, s.alert, 
		case when s.libelle_liste = $5 then s.expl_selection_concerning else '[]' end,			
		case when s.libelle_liste = $5 then s.expl_selection_reassuring else '[]' end, 
		case when s.libelle_liste = $5 then s.macro_expl else '{}' end, 
		case when s.libelle_liste = $5 then s.micro_expl else '{}' end,
		case when s.libelle_liste = $5 then s.macro_radar else '{}' end,
		s.alert_pre_redressements,
		s.redressements
		from score0 s
		inner join f_etablissement_permissions($1, $2) p on p.siret = s.siret and p.score 
		where (p.siret=any($3) or p.siren=any($4))
		order by s.siret, s.batch desc, s.score desc;`,
		roles.zoneGeo(), username, e.Query.Sirets, e.Query.Sirens, lastListe)

	batch.Queue(`select e.siret, id_conso, heure_consomme, montant, effectif, periode
		from etablissement_apconso0 e
		inner join f_etablissement_permissions($1, $2) p on p.siret = e.siret and dgefp
		where (e.siret=any($3) or e.siren=any($4))
		order by siret, periode;`,
		roles.zoneGeo(), username, e.Query.Sirets, e.Query.Sirens)

	batch.Queue(`select e.siret, id_demande, effectif_entreprise, effectif, date_statut, periode_start, 
		periode_end, hta, mta, effectif_autorise, motif_recours_se, heure_consomme, montant_consomme, effectif_consomme
		from etablissement_apdemande0 e
		inner join f_etablissement_permissions($1, $2) p on p.siret = e.siret and dgefp
		where e.siret=any($3) or e.siren=any($4)
		order by siret, periode_start;`,
		roles.zoneGeo(), username, e.Query.Sirets, e.Query.Sirens)

	batch.Queue(`select e.siret, e.periode, 
		case when urssaf and cotisation != 0 then cotisation end, 
		case when urssaf then part_patronale end, 
		case when urssaf then part_salariale end, 
		case when urssaf then montant_majorations end, 
		effectif
		from etablissement_periode_urssaf0 e
		inner join f_etablissement_permissions($1, $2) p on p.siret = e.siret
		where e.siret=any($3) or e.siren=any($4)
		order by siret, periode;`,
		roles.zoneGeo(), username, e.Query.Sirets, e.Query.Sirens)

	batch.Queue(`select e.siret, action, annee_creation, date_creation, date_echeance, denomination,
		duree_delai, indic_6m, montant_echeancier, numero_compte, numero_contentieux, stade
		from etablissement_delai0 e
		inner join f_etablissement_permissions($1, $2) p on p.siret = e.siret and urssaf
		where e.siret=any($3) or e.siren=any($4)
		order by e.siret, date_creation;`,
		roles.zoneGeo(), username, e.Query.Sirets, e.Query.Sirens)

	batch.Queue(`
	SELECT p.siren,
		min(date_effet) as date_effet,
		first(action_procol order by date_effet) as action_procol,
		first(stade_procol order by date_effet) as stade_procol
		FROM etablissement_procol0 p
		where siren = any($1)
		group by siren, CASE
			WHEN 'liquidation'::text = p.action_procol THEN 'liquidation'::text
			WHEN 'redressement'::text = p.action_procol THEN
				CASE
					WHEN 'redressement_plan_continuation'::text = p.action_procol || '_'::text || p.stade_procol THEN 'plan_continuation'::text
					ELSE 'redressement'::text
					END
					WHEN 'sauvegarde'::text = p.action_procol THEN
					CASE
							WHEN 'sauvegarde_plan_continuation'::text =p.action_procol || '_'::text || p.stade_procol THEN 'plan_sauvegarde'::text
							ELSE 'sauvegarde'::text
					END
					ELSE NULL::text
			END
		 order by siren, min(date_effet)`,
		e.sirensFromQuery())

	batch.Queue(`select e.siren, e.date_valeur, e.nb_jours
	from entreprise_paydex0 e
	where e.siren=any($1) and date_valeur + '24 month'::interval >= current_date
	order by e.siren, date_valeur;`,
		e.sirensFromQuery())

	batch.Queue(`select * from get_brother($1, null, null, $2, null, null, true, true, $3, false, 'effectif_desc', false, null, null, null, null, null, $4, null, null, null, null, null, null, null) as brothers;`,
		roles.zoneGeo(), listes[0].ID, username, e.sirensFromQuery(),
	)

	e.addPGEsSelection(&batch, roles, username)

	return &batch
}

func (e *Etablissements) loadEtablissements(rows *pgx.Rows) error {
	var cousins = make(map[string][]Summary)
	for (*rows).Next() {
		var e Summary
		var throwAway interface{}
		err := (*rows).Scan(
			&e.Siret,
			&e.Siren,
			&e.RaisonSociale,
			&e.Commune,
			&e.LibelleDepartement,
			&e.CodeDepartement,
			&e.ValeurScore,
			&e.DetailScore,
			&e.FirstAlert,
			&e.ChiffreAffaire,
			&e.ArreteBilan,
			&e.ExerciceDiane,
			&e.VariationCA,
			&e.ResultatExploitation,
			&e.Effectif,
			&e.EffectifEntreprise,
			&e.LibelleActivite,
			&e.LibelleActiviteN1,
			&e.CodeActivite,
			&e.EtatProcol,
			&e.ActivitePartielle,
			&e.APHeureConsommeAVG12m,
			&e.APMontantAVG12m,
			&e.HausseUrssaf,
			&e.DetteUrssaf,
			&e.Alert,
			&throwAway,
			&throwAway,
			&throwAway,
			&e.Visible,
			&e.InZone,
			&e.Followed,
			&e.FollowedEntreprise,
			&e.Siege,
			&e.Groupe,
			&e.TerrInd,
			&throwAway,
			&throwAway,
			&throwAway,
			&e.PermUrssaf,
			&e.PermDGEFP,
			&e.PermScore,
			&e.PermBDF,
			&e.SecteurCovid,
			&e.ExcedentBrutDExploitation,
			&e.EtatAdministratif,
			&e.EtatAdministratifEntreprise,
		)
		if err != nil {
			return err
		}
		cousins[e.Siret[0:9]] = append(cousins[e.Siret[0:9]], e)
	}

	for k, v := range cousins {
		entreprise := e.Entreprises[k]
		entreprise.EtablissementsSummary = v
		e.Entreprises[k[0:9]] = entreprise
	}
	return nil
}

func (e *Etablissements) loadScore(rows *pgx.Rows) error {
	var scores = make(map[string][]EtablissementScore)
	for (*rows).Next() {
		var sc EtablissementScore
		var explSelection EtablissementScoreExplSelection
		var siret string
		err := (*rows).Scan(&siret, &sc.IDListe, &sc.Batch, &sc.Algo, &sc.Periode, &sc.Score, &sc.Diff, &sc.Alert,
			&explSelection.SelectConcerning, &explSelection.SelectReassuring, &sc.MacroExpl, &sc.MicroExpl, &sc.MacroRadar,
			&sc.AlertPreRedressements, &sc.Redressements)
		if err != nil {
			return err
		}
		if len(explSelection.SelectConcerning) > 0 {
			sc.ExplSelection = &explSelection
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
			if e.Etablissements[pu.siret].PermUrssaf {
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
		err := (*rows).Scan(&siret, &dl.Action, &dl.AnneeCreation, &dl.DateCreation, &dl.DateEcheance, &dl.Denomination,
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
		var siren string
		err := (*rows).Scan(&siren, &pc.DateEffet, &pc.Action, &pc.Stade)
		if err != nil {
			return err
		}
		procol[siren] = append(procol[siren], pc)
	}
	for siren, procol := range procol {
		for siret, etablissement := range e.Etablissements {
			if siret[0:9] == siren {
				etablissement.Procol = procol
				e.Etablissements[siret] = etablissement
			}
		}
	}
	return nil
}

// À réactiver lorsque nous aurons plus de données BDF
// func (e *Etablissements) loadBDF(rows *pgx.Rows) error {
// 	var bdfs = make(map[string][]bdf)
// 	for (*rows).Next() {
// 		var bd bdf
// 		var siren string
// 		err := (*rows).Scan(&siren, &bd.Annee, &bd.ArreteBilan, &bd.DelaiFournisseur, &bd.FinancierCourtTerme,
// 			&bd.PoidsFrng, &bd.DetteFiscale, &bd.FraisFinancier, &bd.TauxMarge,
// 		)
// 		if err != nil {
// 			return err
// 		}
// 		bdfs[siren] = append(bdfs[siren], bd)
// 	}
// 	for k, v := range bdfs {
// 		entreprise := e.Entreprises[k]
// 		entreprise.Bdf = v
// 		e.Entreprises[k] = entreprise
// 	}
// 	return nil
// }

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
		apdemandes[siret] = append(apdemandes[siret], ap)
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

		err := (*rows).Scan(&siren, &di.ArreteBilan, &di.AchatMarchandises, &di.AchatMatieresPremieres, &di.AutonomieFinanciere,
			&di.AutresAchatsChargesExternes, &di.AutresProduitsChargesReprises, &di.BeneficeOuPerte, &di.CAExportation,
			&di.CapaciteAutofinancement, &di.CapaciteRemboursement, &di.CAparEffectif, &di.ChargeExceptionnelle, &di.ChargePersonnel,
			&di.ChargesFinancieres, &di.ChiffreAffaire, &di.ConcesBrevEtDroitsSim, &di.ConcoursBancaireCourant, &di.Consommation,
			&di.CouvertureCaBesoinFdr, &di.CouvertureCaFdr, &di.CreditClient, &di.CreditFournisseur, &di.DegreImmoCorporelle,
			&di.DetteFiscaleEtSociale, &di.DotationAmortissement, &di.EffectifConsolide, &di.EfficaciteEconomique, &di.Endettement,
			&di.EndettementGlobal, &di.EquilibreFinancier, &di.ExcedentBrutDExploitation, &di.Exercice, &di.Exportation,
			&di.FinancementActifCirculant, &di.FraisDeRetD, &di.ImpotBenefice, &di.ImpotsTaxes, &di.IndependanceFinanciere, &di.Interets,
			&di.LiquiditeGenerale, &di.LiquiditeReduite, &di.MargeCommerciale, &di.NombreEtabSecondaire, &di.NombreFiliale, &di.NombreMois,
			&di.OperationsCommun, &di.PartAutofinancement, &di.PartEtat, &di.PartPreteur, &di.PartSalaries, &di.ParticipationSalaries,
			&di.Performance, &di.PoidsBFRExploitation, &di.ProcedureCollective, &di.Production, &di.ProductiviteCapitalFinancier,
			&di.ProductiviteCapitalInvesti, &di.ProductivitePotentielProduction, &di.ProduitExceptionnel, &di.ProduitsFinanciers,
			&di.RendementBrutFondsPropres, &di.RendementCapitauxPropres, &di.RendementRessourcesDurables, &di.RentabiliteEconomique,
			&di.RentabiliteNette, &di.ResultatAvantImpot, &di.ResultatExploitation, &di.RotationStocks, &di.StatutJuridique, &di.SubventionsDExploitation,
			&di.TailleCompoGroupe, &di.TauxDInvestissementProductif, &di.TauxEndettement, &di.TauxInteretFinancier, &di.TauxInteretSurCA,
			&di.TauxMargeCommerciale, &di.TauxValeurAjoutee, &di.ValeurAjoutee,
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

func (e *Etablissements) loadPaydex(rows *pgx.Rows) error {
	var dateValeurs = make(map[string][]time.Time)
	var nbJours = make(map[string][]int)

	for (*rows).Next() {
		var pd paydex // resultat DB
		err := (*rows).Scan(&pd.Siren, &pd.DateValeur, &pd.NBJours)
		if err != nil {
			return err
		}
		dateValeurs[pd.Siren] = append(dateValeurs[pd.Siren], pd.DateValeur)
		nbJours[pd.Siren] = append(nbJours[pd.Siren], pd.NBJours)
	}
	for siren, v := range dateValeurs {
		entreprise := e.Entreprises[siren]
		p := Paydex{
			DateValeur: v,
			NBJours:    nbJours[siren],
		}
		entreprise.Paydex = &p
		e.Entreprises[siren] = entreprise
	}

	return nil
}

func (e *Etablissements) loadSirene(rows *pgx.Rows) error {
	for (*rows).Next() {
		var et Etablissement
		var en Entreprise
		var ti EtablissementTerrInd
		var el ellisphere
		err := (*rows).Scan(&et.Siret, &et.Siren, &en.Siren,
			&en.Sirene.RaisonSociale, &en.Sirene.StatutJuridique, &en.Sirene.StatutJuridiqueN2, &en.Sirene.StatutJuridiqueN1,
			&et.Sirene.ComplementAdresse, &et.Sirene.NumVoie, &et.Sirene.IndRep, &et.Sirene.TypeVoie, &et.Sirene.Voie,
			&et.Sirene.Commune, &et.Sirene.CommuneEtranger, &et.Sirene.DistributionSpeciale, &et.Sirene.CodeCommune,
			&et.Sirene.CodeCedex, &et.Sirene.Cedex, &et.Sirene.CodePaysEtranger, &et.Sirene.PaysEtranger,
			&et.Sirene.CodePostal, &et.Sirene.CodeDepartement, &et.Sirene.Departement, &et.Sirene.Region,
			&et.Sirene.NAF.NomenActivite, &et.Sirene.Creation, &et.Sirene.Latitude, &et.Sirene.Longitude,
			&et.VisiteFCE, &et.Sirene.NAF.LibelleSecteur, &et.Sirene.NAF.CodeSecteur, &et.Sirene.NAF.LibelleActivite,
			&et.Sirene.NAF.CodeActivite, &et.Sirene.NAF.LibelleN2, &et.Sirene.NAF.LibelleN3, &et.Sirene.NAF.LibelleN4,
			&et.Followed, &et.FollowedEntreprise, &et.Visible, &et.Alert, &en.Sirene.Prenom1, &en.Sirene.Prenom2, &en.Sirene.Prenom3,
			&en.Sirene.Prenom4, &en.Sirene.Nom, &en.Sirene.NomUsage, &en.Sirene.Creation, &et.Siege,
			&el.CodeGroupe, &el.RefIDGroupe, &el.RaisocGroupe, &el.AdresseGroupe,
			&el.PersonnePouMGroupe, &el.NiveauDetention, &el.PartFinanciere,
			&el.CodeFiliere, &el.RefIDFiliere, &el.PersonnePouMFiliere, &ti.Code, &ti.Libelle,
			&et.PermScore, &et.PermUrssaf, &et.PermDGEFP, &et.PermBDF, &et.InZone,
		)

		if err != nil {
			return err
		}
		et.setAdresse()
		if ti.Code != "" {
			et.TerrInd = &ti
		}
		if el.CodeGroupe != "" {
			en.Groupe = &el
		}
		e.Etablissements[et.Siret] = et
		e.Entreprises[en.Siren] = en
	}
	return nil
}

func (e *Etablissements) load(roles scope, username string) error {
	tx, err := Db().Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Commit(context.Background())
	batch := e.intoBatch(roles, username)
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

	// // bdf
	// rows, err = b.Query()
	// if err != nil {
	// 	return err
	// }
	// err = e.loadBDF(&rows)
	// if err != nil {
	// 	return err
	// }

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

	// paydex
	rows, err = b.Query()
	if err != nil {
		return err
	}
	err = e.loadPaydex(&rows)
	if err != nil {
		return err
	}

	// brothers
	rows, err = b.Query()
	if err != nil {
		return err
	}
	err = e.loadEtablissements(&rows)
	if err != nil {
		return err
	}

	// pge
	rows, err = b.Query()
	if err != nil {
		return err
	}
	err = e.loadPGE(&rows)
	if err != nil {
		return err
	}
	// end
	return nil
}

func (e *Etablissement) setAdresse() {
	space := regexp.MustCompile(`\s+`)
	adresse := []string{}
	if e.Sirene.ComplementAdresse != nil {
		adresse = append(adresse, str(e.Sirene.ComplementAdresse))
	}
	voie := fmt.Sprintf("%s %s %s", str(e.Sirene.NumVoie), str(e.Sirene.TypeVoie), str(e.Sirene.Voie))
	voie = strings.Trim(space.ReplaceAllString(voie, " "), " ")
	if voie != "" {
		adresse = append(adresse, voie)
	}
	if e.Sirene.DistributionSpeciale != nil {
		adresse = append(adresse, str(e.Sirene.DistributionSpeciale))
	}
	var ville string
	if e.Sirene.CodeCedex != nil && e.Sirene.Cedex != nil {
		ville = str(e.Sirene.CodeCedex) + " " + str(e.Sirene.Cedex)
	} else {
		ville = str(e.Sirene.CodePostal) + " " + str(e.Sirene.Commune)
	}
	adresse = append(adresse, ville)
	if e.Sirene.CodePaysEtranger != nil {
		adresse = append(adresse, str(e.Sirene.CodePaysEtranger))
	}
	adresseStr := strings.Join(adresse, "\n")
	e.Sirene.Adresse = &adresseStr
}

func getSiegeFromSiren(siren string) (string, error) {
	sqlSiege := `select siret from etablissement0
	where siren = $1 and siege`
	var siret string
	err := Db().QueryRow(context.Background(), sqlSiege, siren).Scan(&siret)
	return siret, err
}

func getEntrepriseViewers(c *gin.Context) {
	sqlViewers := `select username, firstname, lastname from v_roles r
		inner join users u on u.roles && r.roles
		where siren = $1`

	siren := c.Param("siren")
	rows, err := Db().Query(context.Background(), sqlViewers, siren)
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
	rows, err := Db().Query(context.Background(), sqlViewers, siren)
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

func str(o ...*string) string {
	for _, s := range o {
		if s != nil {
			return *s
		}
	}
	return ""
}

func getDeptFromSiret(siret string) (string, error) {
	sql := "select departement from etablissement0 where siret = $1"
	row := Db().QueryRow(context.Background(), sql, siret)
	var dept string
	err := row.Scan(&dept)
	return dept, err
}
