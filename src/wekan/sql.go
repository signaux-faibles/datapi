package wekan

// SqlGetCards: $1 = roles.ZoneGeo, $2 = username, $3 = sirets
const SqlGetCards = `
select s.siret, s.siren, s.raison_sociale, s.commune,
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
left join etablissement_follow f on f.active and f.siret = s.siret and f.username = $2
left join v_entreprise_follow fe on fe.siren = s.siren and fe.username = $2
where s.siret = any($3) and (s.code_departement = any($4) or $4 is null)
limit $5`

const SqlGetFollow = `
select s.siret, s.siren, s.raison_sociale, s.commune,
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
inner join etablissement_follow f on f.active and f.siret = s.siret and f.username = $2
inner join v_entreprise_follow fe on fe.siren = s.siren and fe.username = $2
where (s.code_departement = any($3) or $3 is null) and (s.siret != any($4) or $4 is null)`

var sqlDbExport = `select v.siret, v.raison_sociale, v.code_departement, v.libelle_departement, v.commune,
coalesce(v.code_territoire_industrie, ''), coalesce(v.libelle_territoire_industrie, ''), v.siege, coalesce(v.raison_sociale_groupe, ''),
v.code_activite, coalesce(v.libelle_n5, 'norme NAF non prise en charge'), coalesce(v.libelle_n1, 'norme NAF non prise en charge'),
v.statut_juridique_n1, v.statut_juridique_n2, v.statut_juridique_n3, coalesce(v.date_ouverture_etablissement, '1900-01-01'),
coalesce(v.date_creation_entreprise, '1900-01-01'), coalesce(v.effectif_entreprise, 0), coalesce(v.date_effectif, '1900-01-01'),
coalesce(v.arrete_bilan, '0001-01-01'), coalesce(v.exercice_diane,0), coalesce(v.chiffre_affaire,0),
coalesce(v.prev_chiffre_affaire,0), coalesce(v.variation_ca, 1), coalesce(v.resultat_expl,0),
coalesce(v.prev_resultat_expl,0), coalesce(v.excedent_brut_d_exploitation,0),
coalesce(v.prev_excedent_brut_d_exploitation,0), coalesce(v.last_list,''), coalesce(v.last_alert,''),
coalesce(v.activite_partielle, false), coalesce(v.hausse_urssaf, false), coalesce(v.presence_part_salariale, false),
coalesce(v.periode_urssaf, '0001-01-01'), v.last_procol, coalesce(v.date_last_procol, '0001-01-01'), coalesce(f.since, '0001-01-01'),
coalesce(f.comment, ''), (permissions($1, v.roles, v.first_list_entreprise, v.code_departement, f.siret is not null)).in_zone
from v_summaries v
left join etablissement_follow f on f.siret = v.siret and f.username = $2 and active
where v.siret = any($3) and (v.code_departement = any($4) or coalesce($4, '{}') = '{}')
order by f.id, v.siret`

var sqlDbExportWithoutCards = `select v.siret, v.raison_sociale, v.code_departement, v.libelle_departement, v.commune,
coalesce(v.code_territoire_industrie, ''), coalesce(v.libelle_territoire_industrie, ''), v.siege, coalesce(v.raison_sociale_groupe, ''),
v.code_activite, coalesce(v.libelle_n5, 'norme NAF non prise en charge'), coalesce(v.libelle_n1, 'norme NAF non prise en charge'),
v.statut_juridique_n1, v.statut_juridique_n2, v.statut_juridique_n3, coalesce(v.date_ouverture_etablissement, '1900-01-01'),
coalesce(v.date_creation_entreprise, '1900-01-01'), coalesce(v.effectif_entreprise, 0), coalesce(v.date_effectif, '1900-01-01'),
coalesce(v.arrete_bilan, '0001-01-01'), coalesce(v.exercice_diane,0), coalesce(v.chiffre_affaire,0),
coalesce(v.prev_chiffre_affaire,0), coalesce(v.variation_ca, 1), coalesce(v.resultat_expl,0),
coalesce(v.prev_resultat_expl,0), coalesce(v.excedent_brut_d_exploitation,0),
coalesce(v.prev_excedent_brut_d_exploitation,0), coalesce(v.last_list,''), coalesce(v.last_alert,''),
coalesce(v.activite_partielle, false), coalesce(v.hausse_urssaf, false), coalesce(v.presence_part_salariale, false),
coalesce(v.periode_urssaf, '0001-01-01'), v.last_procol, coalesce(v.date_last_procol, '0001-01-01'), coalesce(f.since, '0001-01-01'),
coalesce(f.comment, ''), (permissions($1, v.roles, v.first_list_entreprise, v.code_departement, f.siret is not null)).in_zone
from v_summaries v
inner join etablissement_follow f on f.siret = v.siret and f.username = $2 and active
where (v.code_departement = any($4) or coalesce($4, '{}') = '{}' is null) and v.siret != any($3)
order by f.id, v.siret`

var sqlDbExportSingle = `select v.siret, v.raison_sociale, v.code_departement, v.libelle_departement, v.commune,
coalesce(v.code_territoire_industrie, ''), coalesce(v.libelle_territoire_industrie, ''), v.siege, coalesce(v.raison_sociale_groupe, ''),
v.code_activite, coalesce(v.libelle_n5, 'norme NAF non prise en charge'), coalesce(v.libelle_n1, 'norme NAF non prise en charge'),
v.statut_juridique_n1, v.statut_juridique_n2, v.statut_juridique_n3, coalesce(v.date_ouverture_etablissement, '1900-01-01'),
coalesce(v.date_creation_entreprise, '1900-01-01'), coalesce(v.effectif_entreprise, 0), coalesce(v.date_effectif, '1900-01-01'),
coalesce(v.arrete_bilan, '0001-01-01'), coalesce(v.exercice_diane,0), coalesce(v.chiffre_affaire,0),
coalesce(v.prev_chiffre_affaire,0), coalesce(v.variation_ca, 1), coalesce(v.resultat_expl,0),
coalesce(v.prev_resultat_expl,0), coalesce(v.excedent_brut_d_exploitation,0),
coalesce(v.prev_excedent_brut_d_exploitation,0), coalesce(v.last_list,''), coalesce(v.last_alert,''),
coalesce(v.activite_partielle, false), coalesce(v.hausse_urssaf, false), coalesce(v.presence_part_salariale, false),
coalesce(v.periode_urssaf, '0001-01-01'), v.last_procol, coalesce(v.date_last_procol, '0001-01-01'), coalesce(f.since, '0001-01-01'),
coalesce(f.comment, ''), (permissions($1, v.roles, v.first_list_entreprise, v.code_departement, f.siret is not null)).in_zone
from v_summaries v
left join etablissement_follow f on f.siret = v.siret and f.username = $2 and active
where v.siret = $3
order by f.id, v.siret`
