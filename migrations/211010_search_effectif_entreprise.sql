-- create index idx_etablissement_periode_urssaf_siret_siren on etablissement_periode_urssaf (siret, siren, version)
drop materialized view v_summaries;
drop materialized view v_last_effectif;

create materialized view v_last_effectif as
with last_effectif as (select siret, siren, last(effectif order by periode) as effectif,
last(periode order by periode) as periode
  from etablissement_periode_urssaf0 where effectif is not null
  group by siret, siren)
select siret, siren, effectif, sum(effectif) OVER (PARTITION BY siren) as effectif_entreprise, periode
from last_effectif;


create unique index idx_v_last_effectif_siret
  on v_last_effectif (siret);
create index idx_v_last_effectif_effectif 
  on v_last_effectif (effectif);

create materialized view v_summaries as
  with last_liste as (select first(libelle order by batch desc, algo desc) as last_liste from liste)
  select 
    r.roles, et.siret, et.siren, en.raison_sociale, et.commune,
    d.libelle as libelle_departement, et.departement as code_departement,
    s.score as valeur_score,
    s.detail as detail_score,
    aet.first_list = l.last_liste as first_alert,
    aen.first_list as first_list_entreprise,
  	l.last_liste as liste,
    di.chiffre_affaire, 
	di.prev_chiffre_affaire,
	di.arrete_bilan,
	di.exercice_diane, 
	di.variation_ca, 
	di.resultat_expl, 
	di.prev_resultat_expl,
	di.excedent_brut_d_exploitation,
	di.prev_excedent_brut_d_exploitation,
	ef.effectif, ef.effectif_entreprise, ef.periode as date_effectif,
    n.libelle_n5, n.libelle_n1, et.code_activite, 
	coalesce(ep.last_procol, 'in_bonis') as last_procol,
	ep.date_effet as date_last_procol,
    coalesce(ap.ap, false) as activite_partielle,
    ap.heure_consomme as apconso_heure_consomme,
    ap.montant as apconso_montant,
    u.dette[1] > u.dette[2] or u.dette[2] > u.dette[3] as hausse_urssaf,
    u.dette[1] as dette_urssaf,
	u.periode_urssaf,
	u.presence_part_salariale,
    s.alert,
    et.siege, g.raison_sociale as raison_sociale_groupe,
    ti.code_commune is not null as territoire_industrie,
	ti.code_terrind as code_territoire_industrie,
	ti.libelle_terrind as libelle_territoire_industrie,
	cj.libelle statut_juridique_n3,
	cj2.libelle statut_juridique_n2,
	cj1.libelle statut_juridique_n1,
	et.creation as date_ouverture_etablissement,
	en.creation as date_creation_entreprise,
	aet.last_list,
	aet.last_alert
  from last_liste l
  	inner join etablissement0 et on true
    inner join entreprise0 en on en.siren = et.siren
    inner join v_etablissement_raison_sociale etrs on etrs.id_etablissement = et.id
    inner join v_roles r on et.siren = r.siren
    inner join departements d on d.code = et.departement
    left join v_naf n on n.code_n5 = et.code_activite
    left join score0 s on et.siret = s.siret and s.libelle_liste = l.last_liste
    left join v_alert_etablissement aet on aet.siret = et.siret
    left join v_alert_entreprise aen on aen.siren = et.siren
    left join v_last_effectif ef on ef.siret = et.siret
    left join v_hausse_urssaf u on u.siret = et.siret
    left join v_apdemande ap on ap.siret = et.siret
    left join v_last_procol ep on ep.siret = et.siret
    left join v_diane_variation_ca di on di.siren = et.siren
    left join entreprise_ellisphere0 g on g.siren = et.siren
    left join terrind ti on ti.code_commune = et.code_commune
	left join categorie_juridique cj on cj.code = en.statut_juridique
	left join categorie_juridique cj2 on substring(cj.code from 0 for 3) = cj2.code
	left join categorie_juridique cj1 on substring(cj.code from 0 for 2) = cj1.code;

create index idx_v_summaries_raison_sociale on v_summaries using gin (siret gin_trgm_ops, raison_sociale gin_trgm_ops);
create index idx_v_summaries_siren on v_summaries (siren);
create index idx_v_summaries_siret on v_summaries (siret);
create index idx_v_summaries_order_raison_sociale on v_summaries (raison_sociale, siret);
create index idx_v_summaries_roles on v_summaries using gin (roles);
create index idx_v_summaries_score on v_summaries (
  valeur_score desc,
  siret,
  siege,
  effectif,
  effectif_entreprise,
  code_departement,
  last_procol,
  chiffre_affaire) where alert != 'Pas d''alerte';