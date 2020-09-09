create view entreprise0 as
select * from entreprise where version = 0;

create view entreprise_bdf0 as
select * from entreprise_bdf where version = 0;

create view entreprise_diane0 as
select * from entreprise_diane where version = 0;

create view etablissement0 as
select * from etablissement where version = 0;

create view etablissement_apconso0 as
select * from etablissement_apconso where version = 0;

create view etablissement_apdemande0 as
select * from etablissement_apdemande where version = 0;

create view etablissement_delai0 as
select * from etablissement_delai where version = 0;

create view etablissement_periode_urssaf0 as
select * from etablissement_periode_urssaf where version = 0;

create view etablissement_procol0 as
select * from etablissement_procol where version = 0;

create view score0 as
select * from score where version = 0;

create view v_naf as
select 
  n1.id id_n1,
  n2.id id_n2,
  n3.id id_n3,
  n4.id in_n4,
  n5.id id_n5,
  n1.code code_n1,
  n2.code code_n2,
  n3.code code_n3,
  n4.code code_n4,
  n5.code code_n5,
  n1.libelle libelle_n1,
  n2.libelle libelle_n2,
  n3.libelle libelle_n3,
  n4.libelle libelle_n4,
  n5.libelle libelle_n5
from naf n5
  inner join naf n4 on n5.id_parent = n4.id
  inner join naf n3 on n4.id_parent = n3.id
  inner join naf n2 on n3.id_parent = n2.id
  inner join naf n1 on n2.id_parent = n1.id;

create materialized view v_roles as 
select siren, array_agg(distinct departement) as roles
  from etablissement
  where version = 0
  group by siren;

create index idx_v_roles_siret
  on v_roles (siren);
create index idx_v_roles_roles on v_roles using gin (roles);

create materialized view v_diane_variation_ca as
with diane as (SELECT
	siren,
	arrete_bilan_diane,
	lag(arrete_bilan_diane,1) OVER (partition by siren order by arrete_bilan_diane) prev_arrete_bilan_diane,
	chiffre_affaire,
	lag(chiffre_affaire,1) OVER (partition by siren order by arrete_bilan_diane) prev_chiffre_affaire,
  resultat_expl
from entreprise_diane0
order by siren, arrete_bilan_diane desc)
select siren, 
first(arrete_bilan_diane) as arrete_bilan, 
first(prev_chiffre_affaire) as prev_chiffre_affaire,
first(chiffre_affaire) as chiffre_affaire, 
first(chiffre_affaire / prev_chiffre_affaire) as variation_ca,
first(resultat_expl) as resultat_expl
from diane
where coalesce(prev_chiffre_affaire,0) != 0 and  coalesce(chiffre_affaire,0) != 0 and
arrete_bilan_diane - '1 year'::interval = prev_arrete_bilan_diane 
group by siren;

create unique index idx_v_diane_variation_ca_siret
  on v_diane_variation_ca (siren);

create materialized view v_last_effectif as
select siret, last(effectif order by periode) as effectif 
  from etablissement_periode_urssaf0 where effectif is not null
  group by siret;

create unique index idx_v_last_effectif_siret
  on v_last_effectif (siret);
create index idx_v_last_effectif_effectif 
  on v_last_effectif (effectif);

create materialized view v_last_procol as
select siret, last(action_procol order by date_effet) as last_procol 
  from etablissement_procol0
  group by siret;

create unique index idx_v_last_procol_siret
  on v_last_procol (siret);

create materialized view v_hausse_urssaf as 
select siret, (array_agg(part_patronale + part_salariale order by periode desc))[0:3] as dette
  from etablissement_periode_urssaf0
  group by siret;

create unique index idx_v_hausse_urssaf_siret
  on v_hausse_urssaf (siret);

create materialized view v_apdemande as select siret, true as ap 
  from etablissement_apdemande0
  where periode_start + '1 year'::interval > current_timestamp
  group by siret;

create unique index idx_v_apdemande_siret 
  on v_apdemande (siret);

create materialized view v_score as
select distinct siren, siret
from score
where alert in ('Alerte seuil F1', 'Alerte seuil F2');
create unique index idx_v_score_siret
  on v_score (siret);
create unique index idx_v_score_siren 
  on v_score (siren);