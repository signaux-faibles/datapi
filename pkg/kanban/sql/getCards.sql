-- SqlGetCards $1 = roles.ZoneGeo, $2 = username, $3 = sirets
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
       s.secteur_covid, s.excedent_brut_d_exploitation, s.etat_administratif, s.etat_administratif_entreprise, null
from v_summaries s
       left join etablissement_follow f on f.active and f.siret = s.siret and f.username = $2
       left join v_entreprise_follow fe on fe.siren = s.siren and fe.username = $2
where s.siret = any($3) and (s.code_departement = any($4) or coalesce($4, '{}') = '{}')  and (s.raison_sociale ilike $6 or $6 is null)
limit $5;
