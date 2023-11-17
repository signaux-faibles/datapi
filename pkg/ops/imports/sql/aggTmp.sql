drop table if exists tmp_debit_agg;
create table tmp_debit_agg as
with calendar as (select date_trunc('month', current_date) - generate_series(1, 24) * '1 month'::interval as periode),
     débits as (select c.periode,
                       siret,
                       période_start,
                       période_end,
                       numéro_compte,
                       numéro_écart_négatif,
                       first(part_ouvrière order by numéro_historique_écart_négatif desc)  as part_ouvrière,
                       first(part_patronale order by numéro_historique_écart_négatif desc) as part_patronale
                from tmp_debit d
                         inner join calendar c on d.date_traitement <= c.periode - '1 month'::interval + '20 days'::interval
                group by c.periode, siret, période_start, période_end, numéro_compte, numéro_écart_négatif)
select siret, periode, sum(part_ouvrière) as part_ouvrière, sum(part_patronale) as part_patronale
from débits
group by siret, periode;

-- drop table if exists tmp_debit_agg;
-- create table tmp_debit_agg as
-- with calendar as (select date_trunc('month', current_date) - generate_series(1, 24) * '1 month'::interval as période),
--      debits as (select substring(siret from 1 for 9),
--                        siret,
--                        numéro_compte,
--                        numéro_écart_négatif,
--                        période,
--                        période_start,
--                        période_end,
--                        date_traitement,
--                        first(part_ouvrière) over debit as part_ouvrière,
--                        first(part_patronale) over debit as part_patronale,
--                        numéro_historique_écart_négatif,
--                        first(numéro_historique_écart_négatif) over debit as last_numéro_historique_écart_négatif
--                 from tmp_debit
--                          inner join calendar c on date_traitement <= c.période + '20 days'::interval
--                 window debit as (partition by période, numéro_compte, numéro_écart_négatif, période_start, période_end ORDER BY numéro_historique_écart_négatif DESC))
-- select siret, période, sum(part_ouvrière) as part_ouvrière, sum(part_patronale) as part_patronale from debits
-- where last_numéro_historique_écart_négatif = numéro_historique_écart_négatif
-- group by  siret, période
-- order by période, siret;


drop table if exists tmp_cotisation_agg;
create table tmp_cotisation_agg as
with calendar as (select date_trunc('month', current_date) - generate_series(1, 24) * '1 month'::interval as période)
select siret,
       sum(du / (extract(year from age(co.période_end, co.période_start)) * 12 +
                 extract(month from age(co.période_end, co.période_start)))) as du_période,
       ca.période
from tmp_cotisation co
         inner join calendar ca on ca.période >= co.période_start and période < co.période_end
group by siret, ca.période
order by période;

-- transformation des sets au format hérité de datapi (etablissement_periode_urssaf)
drop table if exists tmp_etablissement_periode_urssaf;
create table tmp_etablissement_periode_urssaf as
with calendar as (select date_trunc('month', current_date) - generate_series(1, 24) * '1 month'::interval as période),
     sirets as (select siret
                from tmp_debit_agg
                union
                select siret
                from tmp_cotisation_agg
                union
                select siret
                from tmp_effectif)
select ca.période, s.siret, d.part_patronale, d.part_ouvrière, c.du_période as cotisation, e.effectif
from calendar ca
         inner join sirets s on true
         inner join etablissement et on et.siret = s.siret
         left join tmp_debit_agg d on d.période = ca.période and s.siret = d.siret
         left join tmp_cotisation_agg c on s.siret = c.siret and ca.période = c.période
         left join tmp_effectif e on s.siret = e.siret and ca.période = e.période;

-- insertion des données
delete
from etablissement_periode_urssaf;
insert into etablissement_periode_urssaf (siret, siren, periode, cotisation, part_patronale, part_salariale, effectif)
select siret, substring(siret from 1 for 9), période, cotisation, part_patronale, part_ouvrière, effectif
from tmp_etablissement_periode_urssaf
where coalesce(part_patronale, part_ouvrière, cotisation, effectif) is not null;

delete
from etablissement_procol;
insert into etablissement_procol (siret, siren, date_effet, action_procol, stade_procol)
select siret, substring(siret from 1 for 9), date_effet, action_procol, stade_procol
from tmp_procol;

delete
from etablissement_apdemande;
insert into etablissement_apdemande (siret, siren, id_demande, effectif_entreprise, effectif, date_statut,
                                     periode_start, periode_end, hta, mta, effectif_autorise, motif_recours_se,
                                     heure_consomme, montant_consomme, effectif_consomme)
select etab_siret,
       substring(etab_siret from 0 for 9),
       id_da,
       eff_ent,
       eff_etab,
       date_statut::date,
       date_deb::date,
       date_fin::date,
       hta,
       mta,
       eff_auto,
       motif_recours_se,
       s_heure_consom_tot,
       s_montant_consom_tot,
       s_eff_consom_tot
from demande_ap;

delete
from etablissement_apconso;
insert into etablissement_apconso (siret, siren, id_conso, heure_consomme, montant, effectif, periode)
select etab_siret, substring(etab_siret from 1 for 9), id_da, heures, montants, effectifs, mois::date
from consommation_ap;

-- testing only
delete
from etablissement
where siren not in (select distinct siren
                from score
                where batch = '2312' and alert != 'Pas d''alerte');

delete from entreprise
where siren not in (select distinct siren
                    from score
                    where batch = '2312' and alert != 'Pas d''alerte');
delete
from etablissement_periode_urssaf
where siren not in (select distinct siren
                from score
                where batch = '2312' and alert != 'Pas d''alerte');

delete
from entreprise_bce
where siren not in (select distinct siren
                from score
                where batch = '2312' and alert != 'Pas d''alerte');

delete
from entreprise_ellisphere
where siren not in (select distinct siren
                from score
                where batch = '2312' and alert != 'Pas d''alerte');

delete
from entreprise_diane
where siren not in (select distinct siren
                from score
                where batch = '2312' and alert != 'Pas d''alerte');

delete
from entreprise_paydex
where siren not in (select distinct siren
                from score
                where batch = '2312' and alert != 'Pas d''alerte');

delete
from etablissement_apconso
where siren not in (select distinct siren
                from score
                where batch = '2312' and alert != 'Pas d''alerte');

delete
from etablissement_apdemande
where siren not in (select distinct siren
                from score
                where batch = '2312' and alert != 'Pas d''alerte');

delete
from etablissement_delai
where siren not in (select distinct siren
                from score
                where batch = '2312' and alert != 'Pas d''alerte');

delete
from etablissement_procol
where siren not in (select distinct siren
                from score
                where batch = '2312' and alert != 'Pas d''alerte');

delete
from etablissement_periode_urssaf
where siren not in (select distinct siren
                from score
                where batch = '2312' and alert != 'Pas d''alerte');

refresh materialized view v_alert_entreprise;
refresh materialized view v_alert_etablissement;
refresh materialized view v_apdemande;
refresh materialized view v_diane_variation_ca;
refresh materialized view v_etablissement_raison_sociale;
refresh materialized view v_hausse_urssaf;
refresh materialized view v_last_effectif;
refresh materialized view v_last_procol;
refresh materialized view v_naf;
refresh materialized view v_roles;
refresh materialized view v_summaries;
