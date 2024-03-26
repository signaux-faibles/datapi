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

truncate table etablissement_delai;
insert into etablissement_delai
    (siret, siren, action, annee_creation, date_creation,
     date_echeance, denomination, duree_delai, indic_6m, montant_echeancier,
     numero_compte, numero_contentieux, stade)
    select siret, substring(siret from 1 for 9), action, année_création, date_création,
           date_échéance, dénomination, durée_délai, case when indic_6mois then 'SUP' else 'INF' end, montant_échéancier,
           numéro_compte, numéro_contentieux, stade
    from tmp_delai;

drop table if exists tmp_effectif_agg;
create table tmp_effectif_agg as
select siret, période, sum(effectif) as effectif from tmp_effectif
where période > (current_date - '2 years 2 month'::interval)
group by siret, période;

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
         left join tmp_debit_agg d on d.periode = ca.période and s.siret = d.siret
         left join tmp_cotisation_agg c on s.siret = c.siret and ca.période = c.période
         left join tmp_effectif_agg e on s.siret = e.siret and ca.période = e.période;

-- insertion des données
truncate table etablissement_periode_urssaf;
insert into etablissement_periode_urssaf (siret, siren, periode, cotisation, part_patronale, part_salariale, effectif)
select siret, substring(siret from 1 for 9), période, cotisation, part_patronale, part_ouvrière, effectif
from tmp_etablissement_periode_urssaf
where coalesce(part_patronale, part_ouvrière, cotisation, effectif) is not null;

truncate table etablissement_procol;
insert into etablissement_procol (siret, siren, date_effet, action_procol, stade_procol)
select siret, substring(siret from 1 for 9), date_effet, action_procol, stade_procol
from tmp_procol;
