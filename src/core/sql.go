package core

func (p summaryParams) toSQLCurrentScoreParams() []interface{} {
	var expressionSiret *string
	var expressionRaisonSociale *string

	if p.filter != nil {
		eSiret := *p.filter + "%"
		eRaisonSociale := "%" + *p.filter + "%"
		expressionSiret = &eSiret
		expressionRaisonSociale = &eRaisonSociale
	}
	return []interface{}{
		p.zoneGeo,
		p.limit,
		p.offset,
		expressionSiret,
		expressionRaisonSociale,
		p.ignoreRoles,
		p.ignoreZone,
		p.userName,
		p.siegeUniquement,
		p.etatsProcol,
		p.departements,
		p.suivi,
		p.effectifMin,
		p.effectifMax,
		p.activites,
		p.effectifMinEntreprise,
		p.effectifMaxEntreprise,
		p.caMin,
		p.caMax,
		p.excludeSecteursCovid,
		p.etatAdministratif,
		p.firstAlert,
	}
}

// sqlCurrentScore récupère les derniers scores connus
// output: siret text, siren text, raison_sociale text, commune text, libelle_departement text,
// code_departement text, valeur_score real, detail_score jsonb, first_alert boolean,
// chiffre_affaire real, arrete_bilan date, exercice_diane integer, variation_ca real,
// resultat_expl real, effectif real, effectif_entreprise real, libelle_n5 text, libelle_n1 text, code_activite text,
// last_procol text, activite_partielle boolean, apconso_heure_consomme integer,
// apconso_montant integer, hausse_urssaf boolean, dette_urssaf real, alert text,
// nb_total bigint, nb_f1 bigint, nb_f2 bigint, visible boolean, in_zone boolean,
// followed boolean, followed_enterprise boolean, siege boolean, raison_sociale_groupe text,
// territoire_industrie boolean, comment text, category text, since timestamp without time zone,
// urssaf boolean, dgefp boolean, score boolean, bdf boolean, secteur_covid text, excedent_brut_d_exploitation real,
// etat_administratif text, etat_administratif_entreprise text
var sqlCurrentScore = `select 
  s.siret, s.siren, s.raison_sociale, s.commune,
  s.libelle_departement, s.code_departement,
  case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score then s.valeur_score end as valeur_score,
  case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score then s.detail_score end as detail_score,
  case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score then s.first_alert end as first_alert,
  s.chiffre_affaire, s.arrete_bilan, s.exercice_diane, s.variation_ca, s.resultat_expl, s.effectif, s.effectif_entreprise,
  s.libelle_n5, s.libelle_n1, s.code_activite, s.last_procol,
  case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).dgefp then s.activite_partielle end as activite_partielle,
  case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).dgefp then s.apconso_heure_consomme end as apconso_heure_consomme,
  case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).dgefp then s.apconso_montant end as apconso_montant,
  case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).urssaf then s.hausse_urssaf end as hausse_urssaf,
  case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).urssaf then s.dette_urssaf end as dette_urssaf,
  case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score then s.alert end,
  count(*) over () as nb_total,
  count(case when s.alert='Alerte seuil F1' and (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score then 1 end) over () as nb_f1,
  count(case when s.alert='Alerte seuil F2' and (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score then 1 end) over () as nb_f2,
  (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).visible, 
  (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).in_zone, 
  f.id is not null as followed_etablissement,
  fe.siren is not null as followed_entreprise,
  s.siege, s.raison_sociale_groupe, territoire_industrie,
  f.comment, f.category, f.since,
  (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).urssaf,
  (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).dgefp,
  (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score,
  (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).bdf,
  s.secteur_covid, s.excedent_brut_d_exploitation, s.etat_administratif, s.etat_administratif_entreprise
from v_summaries s
  left join v_naf n on n.code_n5 = s.code_activite
  left join etablissement_follow f on f.active and f.siret = s.siret and f.username = $8
  left join v_entreprise_follow fe on fe.siren = s.siren and fe.username = $8
  where 
  (s.roles && $1 or $6)
  and s.alert != 'Pas d''alerte'
  and (s.code_departement=any($1) or $7)
  and (s.siege or not $9)
  and (s.last_procol = any($10) or $10 is null)
  and (s.code_departement=any($11) or $11 is null)
  and (s.effectif >= $13 or $13 is null)
  and (s.effectif <= $14 or $14 is null)
  and (s.effectif_entreprise >= $16 or $16 is null)
  and (s.effectif_entreprise <= $17 or $17 is null)
  and (s.chiffre_affaire >= $18 or $18 is null)
  and (s.chiffre_affaire <= $19 or $19 is null)
  and (fe.siren is not null = $12 or $12 is null)
  and (s.raison_sociale ilike $5 or s.siret ilike $4 or coalesce($4, $5) is null)
  and 'score' = any($1)
  and (not (s.secteur_covid = any($20)) or $20 is null)
  and (n.code_n1 = any($15) or $15 is null)
  and (s.etat_administratif = $21 or $21 is null)
  and (s.first_alert = $22 or $22 is null)
order by s.alert, s.valeur_score desc, s.siret
limit $2 offset $3`

func (p summaryParams) toSQLScoreParams() []interface{} {
	var expressionSiret *string
	var expressionRaisonSociale *string

	if p.filter != nil {
		eSiret := *p.filter + "%"
		eRaisonSociale := "%" + *p.filter + "%"
		expressionSiret = &eSiret
		expressionRaisonSociale = &eRaisonSociale
	}
	return []interface{}{
		p.zoneGeo,
		p.limit,
		p.offset,
		expressionSiret,
		expressionRaisonSociale,
		p.ignoreRoles,
		p.ignoreZone,
		p.userName,
		p.siegeUniquement,
		p.etatsProcol,
		p.departements,
		p.suivi,
		p.effectifMin,
		p.effectifMax,
		p.activites,
		p.effectifMinEntreprise,
		p.effectifMaxEntreprise,
		p.caMin,
		p.caMax,
		p.excludeSecteursCovid,
		p.etatAdministratif,
		p.firstAlert,
		p.libelleListe,
	}
}

// sqlScore récupère les scores de la liste passée en paramètre
// output: siret text, siren text, raison_sociale text, commune text, libelle_departement text,
//   code_departement text, valeur_score real, detail_score jsonb, first_alert boolean,
//   chiffre_affaire real, arrete_bilan date, exercice_diane integer, variation_ca real,
//   resultat_expl real, effectif real, effectif_entreprise real, libelle_n5 text, libelle_n1 text, code_activite text,
//   last_procol text, activite_partielle boolean, apconso_heure_consomme integer,
//   apconso_montant integer, hausse_urssaf boolean, dette_urssaf real, alert text,
//   nb_total bigint, nb_f1 bigint, nb_f2 bigint, visible boolean, in_zone boolean,
//   followed boolean, followed_enterprise boolean, siege boolean, raison_sociale_groupe text,
//   territoire_industrie boolean, comment text, category text, since timestamp without time zone,
//   urssaf boolean, dgefp boolean, score boolean, bdf boolean, secteur_covid text, excedent_brut_d_exploitation real,
//   etat_administratif text, etat_administratif_entreprise text
var sqlScore = `select 
  s.siret, s.siren, s.raison_sociale, s.commune,
  s.libelle_departement, s.code_departement,
  case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score then sc.score end as valeur_score,
  case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score then sc.detail end as detail_score,
  case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score then coalesce(s.first_list_etablissement = $23 or s.first_red_list_etablissement = $23, false) end as first_alert,
  s.chiffre_affaire, s.arrete_bilan, s.exercice_diane, s.variation_ca, s.resultat_expl, s.effectif, s.effectif_entreprise,
  s.libelle_n5, s.libelle_n1, s.code_activite, s.last_procol,
  case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).dgefp then s.activite_partielle end as activite_partielle,
  case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).dgefp then s.apconso_heure_consomme end as apconso_heure_consomme,
  case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).dgefp then s.apconso_montant end as apconso_montant,
  case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).urssaf then s.hausse_urssaf end as hausse_urssaf,
  case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).urssaf then s.dette_urssaf end as dette_urssaf,
  case when (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score then sc.alert end,
  count(*) over () as nb_total,
  count(case when sc.alert='Alerte seuil F1' and (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score then 1 end) over () as nb_f1,
  count(case when sc.alert='Alerte seuil F2' and (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score then 1 end) over () as nb_f2,
  (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).visible, 
  (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).in_zone, 
  f.id is not null as followed_etablissement,
  fe.siren is not null as followed_entreprise,
  s.siege, s.raison_sociale_groupe, territoire_industrie,
  f.comment, f.category, f.since,
  (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).urssaf,
  (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).dgefp,
  (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).score,
  (permissions($1, s.roles, s.first_list_entreprise, s.code_departement, fe.siren is not null)).bdf,
  s.secteur_covid, s.excedent_brut_d_exploitation, s.etat_administratif, s.etat_administratif_entreprise
  from v_summaries s
  inner join score0 sc on sc.siret = s.siret and sc.libelle_liste = $23 and sc.alert != 'Pas d''alerte'
  left join v_naf n on n.code_n5 = s.code_activite
  left join etablissement_follow f on f.active and f.siret = s.siret and f.username = $8
  left join v_entreprise_follow fe on fe.siren = s.siren and fe.username = $8
  where 
  (s.roles && $1 or $6)
  and s.alert != 'Pas d''alerte'
  and (s.code_departement=any($1) or $7)
  and (s.siege or not $9)
  and (s.last_procol = any($10) or $10 is null)
  and (s.code_departement=any($11) or $11 is null)
  and (s.effectif >= $13 or $13 is null)
  and (s.effectif <= $14 or $14 is null)
  and (s.effectif_entreprise >= $16 or $16 is null)
  and (s.effectif_entreprise <= $17 or $17 is null)
  and (s.chiffre_affaire >= $18 or $18 is null)
  and (s.chiffre_affaire <= $19 or $19 is null)
  and (fe.siren is not null = $12 or $12 is null)
  and (s.raison_sociale ilike $5 or s.siret ilike $4 or coalesce($4, $5) is null)
  and 'score' = any($1)
  and (not (s.secteur_covid = any($20)) or $20 is null)
  and (n.code_n1 = any($15) or $15 is null)
  and (s.etat_administratif = $21 or $21 is null)
  and ((s.first_list_etablissement = $23 or s.first_red_list_etablissement = $23) and $22 or $22 is null)
  order by s.alert, s.valeur_score desc, s.siret
  limit $2 offset $3`
