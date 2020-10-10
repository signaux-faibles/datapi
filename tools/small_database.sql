-- choisir au hasard 200 code siren
create table limit_entreprise as
select siren from entreprise 
order by id
limit 300
offset 1000;

delete from entreprise
where siren not in (select siren from limit_entreprise);

delete from entreprise_bdf
where siren not in (select siren from limit_entreprise);

delete from entreprise_diane
where siren not in (select siren from limit_entreprise);

delete from entreprise_ellisphere
where siren not in (select siren from limit_entreprise);

delete from etablissement
where siren not in (select siren from limit_entreprise);

delete from etablissement_apconso
where siren not in (select siren from limit_entreprise);

delete from etablissement_apdemande
where siren not in (select siren from limit_entreprise);

delete from etablissement_comments
where siren not in (select siren from limit_entreprise);

delete from etablissement_delai
where siren not in (select siren from limit_entreprise);

delete from etablissement_follow
where siren not in (select siren from limit_entreprise);

delete from etablissement_periode_urssaf
where siren not in (select siren from limit_entreprise);

delete from etablissement_procol
where siren not in (select siren from limit_entreprise);

delete from score
where siren not in (select siren from limit_entreprise);

refresh materialized view v_roles;
refresh materialized view v_diane_variation_ca;
refresh materialized view v_last_effectif;
refresh materialized view v_last_procol;
refresh materialized view v_hausse_urssaf;
refresh materialized view v_apdemande;
refresh materialized view v_alert_etablissement;
refresh materialized view v_alert_entreprise;

drop table limit_entreprise;