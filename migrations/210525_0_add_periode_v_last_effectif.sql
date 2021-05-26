drop materialized view v_summaries;
drop materialized view v_last_effectif;
drop materialized view v_diane_variation_ca;
drop materialized view v_last_procol; 
drop materialized view v_alert_etablissement;
drop materialized view v_hausse_urssaf;

create materialized view v_alert_etablissement as
select siret, min(libelle_liste) as first_list,
max(libelle_liste) as last_list,
first(alert order by batch desc) as last_alert
from score0
where alert in ('Alerte seuil F1', 'Alerte seuil F2')
group by siret;

create unique index idx_v_alert_etablissement_siret
  on v_alert_etablissement (siret);
  
create materialized view v_last_procol as
select siret, last(action_procol order by date_effet) as last_procol,
last(stade_procol) as last_stade_procol,
jsonb_agg(
	jsonb_build_object('date_effet', date_effet, 'stade', stade_procol, 'action', action_procol) order by date_effet desc
) as procol_history
from etablissement_procol0
group by siret;

create unique index idx_v_last_procol
  on v_last_procol (siret);

create materialized view v_diane_variation_ca as
with diane as (SELECT
	siren,
	arrete_bilan_diane,
  exercice_diane::int,
	lag(arrete_bilan_diane,1) OVER (partition by siren order by arrete_bilan_diane) prev_arrete_bilan_diane,
	chiffre_affaire,
	lag(chiffre_affaire,1) OVER (partition by siren order by arrete_bilan_diane) prev_chiffre_affaire,
  resultat_expl,
  lag(resultat_expl,1) OVER (partition by siren order by arrete_bilan_diane) prev_resultat_expl,
  excedent_brut_d_exploitation,
  lag(excedent_brut_d_exploitation,1) OVER (partition by siren order by arrete_bilan_diane) prev_excedent_brut_d_exploitation
from entreprise_diane0
order by siren, arrete_bilan_diane desc)
select siren, 
first(arrete_bilan_diane) as arrete_bilan,
first(exercice_diane) as exercice_diane,
first(prev_chiffre_affaire) as prev_chiffre_affaire,
first(chiffre_affaire) as chiffre_affaire, 
first(chiffre_affaire / prev_chiffre_affaire) as variation_ca,
first(resultat_expl) as resultat_expl,
first(prev_resultat_expl) as prev_resultat_expl,
first(excedent_brut_d_exploitation) as excedent_brut_d_exploitation,
first(prev_excedent_brut_d_exploitation) as prev_excedent_brut_d_exploitation
from diane
where coalesce(prev_chiffre_affaire,0) != 0 and  coalesce(chiffre_affaire,0) != 0 and
arrete_bilan_diane - '1 year'::interval = prev_arrete_bilan_diane 
group by siren;

create unique index idx_v_diane_variation_ca_siren
  on v_diane_variation_ca (siren);

create materialized view v_last_effectif as
select siret, last(effectif order by periode) as effectif,
last(periode order by periode) as periode
  from etablissement_periode_urssaf0 where effectif is not null
  group by siret;

create unique index idx_v_last_effectif_siret
  on v_last_effectif (siret);
create index idx_v_last_effectif_effectif 
  on v_last_effectif (effectif);

create materialized view v_hausse_urssaf as 
select siret, (array_agg(part_patronale + part_salariale order by periode desc))[0:3] as dette,
  case when first(part_salariale order by periode desc) > 100 then true else false end as presence_part_salariale
  from etablissement_periode_urssaf0
  group by siret;

create unique index idx_v_hausse_urssaf_siret
  on v_hausse_urssaf (siret);

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
    di.chiffre_affaire, di.arrete_bilan, di.exercice_diane, di.variation_ca, di.resultat_expl, ef.effectif,
    n.libelle_n5, n.libelle_n1, et.code_activite, coalesce(ep.last_procol, 'in_bonis') as last_procol,
    coalesce(ap.ap, false) as activite_partielle,
    ap.heure_consomme as apconso_heure_consomme,
    ap.montant as apconso_montant,
    u.dette[1] > u.dette[2] or u.dette[2] > u.dette[3] as hausse_urssaf,
    u.dette[1] as dette_urssaf,
    s.alert,
    et.siege, g.raison_sociale as raison_sociale_groupe,
    ti.code_commune is not null as territoire_industrie
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
    left join terrind ti on ti.code_commune = et.code_commune;

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
  code_departement,
  last_procol) where alert != 'Pas d''alerte';