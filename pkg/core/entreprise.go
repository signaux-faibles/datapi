package core

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"datapi/pkg/db"
	"datapi/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
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
		NAF               NAF       `json:"naf,omitempty"`
	}
	Paydex *Paydex `json:"paydex,omitempty"`
	Diane  []Diane `json:"diane"`
	//Bdf                   []Bdf           `json:"-"`
	EtablissementsSummary []Summary       `json:"etablissementsSummary,omitempty"`
	Etablissements        []Etablissement `json:"etablissements,omitempty"`
	Groupe                *Ellisphere     `json:"groupe,omitempty"`
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
	Paydex              []*string    `json:"paydex"`
	JoursRetard         []*int       `json:"joursRetard"`
	Fournisseurs        []*int       `json:"fournisseurs"`
	Encours             []*float64   `json:"encours"`
	ExperiencesPaiement []*int       `json:"experiencesPaiement"`
	FPI30               []*int       `json:"fpi30"`
	FPI90               []*int       `json:"fpi90"`
	DateValeur          []*time.Time `json:"dateValeur"`
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

type NAF struct {
	LibelleActivite *string `json:"libelleActivite,omitempty"`
	LibelleSecteur  *string `json:"libelleSecteur,omitempty"`
	CodeSecteur     *string `json:"codeSecteur,omitempty"`
	CodeActivite    *string `json:"codeActivite,omitempty"`
	LibelleN2       *string `json:"libelleN2,omitempty"`
	LibelleN3       *string `json:"libelleN3,omitempty"`
	LibelleN4       *string `json:"libelleN4,omitempty"`
	NomenActivite   *string `json:"nomenActivite,omitempty"`
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
		NAF                  NAF        `json:"naf"`
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
	IDDemande          *string    `json:"idDemande"`
	EffectifEntreprise *int       `json:"effectifEntreprise"`
	Effectif           *int       `json:"effectif"`
	EffectifAutorise   *int       `json:"effectifAutorise"`
	DateStatut         *time.Time `json:"dateStatut"`
	PeriodeStart       *time.Time `json:"debut"`
	PeriodeEnd         *time.Time `json:"fin"`
	HTA                *float64   `json:"hta"`
	MTA                *float64   `json:"mta"`
	MotifRecoursSE     *int       `json:"motifRecoursSE"`
	HeureConsomme      *float64   `json:"heureConsomme"`
	MontantConsomme    *float64   `json:"montantConsomme"`
	EffectifConsomme   *int       `json:"effectifConsomme"`
}

// EtablissementAPConso …
type EtablissementAPConso struct {
	IDConso       *string    `json:"idConso"`
	HeureConsomme *float64   `json:"heureConsomme"`
	Montant       *float64   `json:"montant"`
	Effectif      *int       `json:"effectif"`
	Periode       *time.Time `json:"date"`
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
	MacroExpl             map[string]float64               `json:"macroExpl,omitempty"`
	MicroExpl             map[string]interface{}           `json:"microExpl,omitempty"`
	AlertPreRedressements string                           `json:"alertPreRedressements"`
	Redressements         []string                         `json:"redressements"`
}

func fixConfidentialiteRedressements(redressements []string, confidentiels []string) []string {
	newRedressements := []string{}
	confidentiel := false
	for _, redressement := range redressements {
		if utils.Contains(confidentiels, redressement) {
			confidentiel = true
		} else {
			newRedressements = append(newRedressements, redressement)
		}
	}
	if confidentiel {
		newRedressements = append(newRedressements, "confidentiel")
	}
	return newRedressements
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

func (e *Etablissements) intoBatch(roles Scope, username string) *pgx.Batch {
	var batch pgx.Batch
	now := time.Now()
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
		p.pge as permPGE,
		p.in_zone as in_zone,
		ne.libelle_n1 as libelle_n1_entreprise,
        ne.code_n1 as code_n1_entreprise,
        ne.libelle_n5 as libelle_n5_entreprise,
        ne.code_n5 as code_n5_entreprise,
        ne.libelle_n2 as libelle_n2_entreprise,
		ne.libelle_n3 as libelle_n3_entreprise,
		ne.libelle_n4 as libelle_n4_entreprise,
		en.nomen_activite as nomen_activite_entreprise
		from etablissement0 et
		inner join f_etablissement_permissions($1, $2) p on p.id = et.id
		inner join departements d on d.code = et.departement
		inner join regions r on d.id_region = r.id
		inner join v_roles ro on ro.siren = et.siren
		left join v_naf n on n.code_n5 = et.code_activite
		left join etablissement_follow f on f.siret = et.siret and f.active and f.username = $2
		left join entreprise0 en on en.siren = et.siren
		left join v_naf ne on ne.code_n5 = en.code_activite
		left join categorie_juridique j on j.code = en.statut_juridique
		left join categorie_juridique j2 on substring(j.code from 0 for 3) = j2.code
		left join categorie_juridique j1 on substring(j.code from 0 for 2) = j1.code
		left join v_alert_entreprise s on s.siren = et.siren
		left join entreprise_ellisphere0 g on g.siren = et.siren
		left join terrind ti on ti.code_commune = et.code_commune
		where (et.siret=any($3) or et.siren=any($4));
	`, roles, username, e.Query.Sirets, e.Query.Sirens)

	batch.Queue(`select siren, arrete_bilan_diane, chiffre_affaire, excedent_brut_d_exploitation, exercice_diane,
       benefice_ou_perte, resultat_expl
			from entreprise_diane0 en
			where en.siren=any($1)
			order by en.arrete_bilan_diane;`,
		e.sirensFromQuery())

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
		roles, username, e.Query.Sirets, e.Query.Sirens, lastListe)

	batch.Queue(`select e.siret, id_conso, heure_consomme, montant, effectif, periode
		from etablissement_apconso0 e
		inner join f_etablissement_permissions($1, $2) p on p.siret = e.siret and dgefp
		where (e.siret=any($3) or e.siren=any($4))
		order by siret, periode;`,
		roles, username, e.Query.Sirets, e.Query.Sirens)

	batch.Queue(`select e.siret, id_demande, effectif_entreprise, effectif, date_statut, periode_start,
		periode_end, hta, mta, effectif_autorise, motif_recours_se, heure_consomme, montant_consomme, effectif_consomme
		from etablissement_apdemande0 e
		inner join f_etablissement_permissions($1, $2) p on p.siret = e.siret and dgefp
		where e.siret=any($3) or e.siren=any($4)
		order by siret, periode_start;`,
		roles, username, e.Query.Sirets, e.Query.Sirens)

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
		roles, username, e.Query.Sirets, e.Query.Sirens)

	batch.Queue(`select distinct e.siret, action, annee_creation, date_creation, date_echeance, denomination,
		duree_delai, indic_6m, montant_echeancier, numero_compte, numero_contentieux, stade
		from etablissement_delai0 e
		inner join f_etablissement_permissions($1, $2) p on p.siret = e.siret and urssaf
		where (e.siret=any($3) or e.siren=any($4))
        and stade != 'PRO PR'
		order by e.siret, date_creation;`,
		roles, username, e.Query.Sirets, e.Query.Sirens)

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

	batch.Queue(`select e.siren,
               last(e.paydex order by date_valeur) as paydex,
               last(e.jours_retard order by date_valeur) as jours_retard,
               last(e.fournisseurs order by date_valeur) as fournisseurs,
               last(e.en_cours order by date_valeur) as en_cours,
               last(e.experiences_paiement order by date_valeur) as experiences_paiement,
               last(e.fpi_30 order by date_valeur) as fpi_30,
               last(e.fpi_90 order by date_valeur) as fpi_90,
               date_trunc('month', e.date_valeur) + '15 days'::interval as date_valeur
        from entreprise_paydex e
        where e.siren=any($1) and date_valeur + '24 month'::interval >= $2
        group by siren, date_trunc('month', e.date_valeur)
        order by e.siren, date_trunc('month', e.date_valeur);`,
		e.sirensFromQuery(), now)

	batch.Queue(`select * from get_brother($1, null, null, $2, null, null, true, true, $3, false, 'effectif_desc', false, null, null, null, null, null, $4, null, null, null, null, null, null, null) as brothers;`,
		roles, listes[0].ID, username, e.sirensFromQuery(),
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

func (e Entreprise) hasDiane() bool {
	for _, exercice := range e.Diane {
		if exercice.Exercice > 2020 {
			return true
		}
	}
	return false
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
		if !(e.Entreprises[siret[0:9]]).hasDiane() {
			sc.Redressements = fixConfidentialiteRedressements(sc.Redressements, []string{"solvabilité_faible", "k_propres_négatifs", "rentabilité_faible"})
		} else {
			sc.Redressements = fixConfidentialiteRedressements(sc.Redressements, []string{"rentabilité_faible"})
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
	var dianes = make(map[string][]Diane)
	for (*rows).Next() {
		var di Diane
		var siren string

		err := (*rows).Scan(&siren, &di.ArreteBilan, &di.ChiffreAffaire, &di.ExcedentBrutDExploitation,
			&di.Exercice, &di.BeneficeOuPerte, &di.ResultatExploitation)

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
	allPaydex := make(map[string][]*string)
	allJoursRetard := make(map[string][]*int)
	allFournisseurs := make(map[string][]*int)
	allEncours := make(map[string][]*float64)
	allExperiencesPaiement := make(map[string][]*int)
	allFpi30 := make(map[string][]*int)
	allFpi90 := make(map[string][]*int)
	allDateValeurs := make(map[string][]*time.Time)

	for (*rows).Next() {
		var siren string
		var paydex *string
		var joursRetard *int
		var fournisseurs *int
		var encours *float64
		var experiencesPaiement *int
		var fpi30 *int
		var fpi90 *int
		var dateValeur *time.Time

		err := (*rows).Scan(&siren, &paydex, &joursRetard, &fournisseurs, &encours, &experiencesPaiement, &fpi30, &fpi90, &dateValeur)
		if err != nil {
			return err
		}

		allPaydex[siren] = append(allPaydex[siren], paydex)
		allJoursRetard[siren] = append(allJoursRetard[siren], joursRetard)
		allFournisseurs[siren] = append(allFournisseurs[siren], fournisseurs)
		allEncours[siren] = append(allEncours[siren], encours)
		allExperiencesPaiement[siren] = append(allExperiencesPaiement[siren], experiencesPaiement)
		allFpi30[siren] = append(allFpi30[siren], fpi30)
		allFpi90[siren] = append(allFpi90[siren], fpi90)
		allDateValeurs[siren] = append(allDateValeurs[siren], dateValeur)
	}

	for siren, _ := range allPaydex {
		entreprise := e.Entreprises[siren]
		p := Paydex{
			Paydex:              allPaydex[siren],
			JoursRetard:         allJoursRetard[siren],
			Fournisseurs:        allFournisseurs[siren],
			Encours:             allEncours[siren],
			ExperiencesPaiement: allExperiencesPaiement[siren],
			FPI30:               allFpi30[siren],
			FPI90:               allFpi90[siren],
			DateValeur:          allDateValeurs[siren],
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
		var el Ellisphere
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
			&et.PermScore, &et.PermUrssaf, &et.PermDGEFP, &et.PermBDF, &et.PermPGE, &et.InZone,
			&en.Sirene.NAF.LibelleSecteur, &en.Sirene.NAF.CodeSecteur, &en.Sirene.NAF.LibelleActivite, &en.Sirene.NAF.CodeActivite,
			&en.Sirene.NAF.LibelleN2, &en.Sirene.NAF.LibelleN3, &en.Sirene.NAF.LibelleN4, &en.Sirene.NAF.NomenActivite,
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

// EtablissementData type pour les données de l'établissement à faire figurer sur la carte Wekan
type EtablissementData struct {
	Siret           string
	RaisonSociale   string
	Departement     string
	Region          string
	Effectif        int
	CodeActivite    string
	LibelleActivite string
}

func (e *Etablissements) load(roles Scope, username string) error {
	tx, err := db.Get().Begin(context.Background())
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

	rows, err = b.Query()
	if err != nil {
		return err
	}
	err = e.loadDiane(&rows)
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
	var adresse []string
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
	err := db.Get().QueryRow(context.Background(), sqlSiege, siren).Scan(&siret)
	return siret, err
}

func getEntrepriseViewers(c *gin.Context) {
	sqlViewers := `select username, firstname, lastname from v_roles r
		inner join users u on u.roles && r.roles
		where siren = $1`

	siren := c.Param("siren")
	rows, err := db.Get().Query(context.Background(), sqlViewers, siren)
	defer rows.Close()
	if err != nil {
		c.JSON(500, err.Error())
		return
	}
	var users []keycloakUser
	for rows.Next() {
		var u keycloakUser
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
	rows, err := db.Get().Query(context.Background(), sqlViewers, siren)
	defer rows.Close()
	if err != nil {
		c.JSON(500, err.Error())
		return
	}
	var users []keycloakUser
	for rows.Next() {
		var u keycloakUser
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
