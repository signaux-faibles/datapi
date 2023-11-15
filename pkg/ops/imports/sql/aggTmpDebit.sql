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
                group by c.periode, siret, période_start, période_end, numéro_compte, numéro_écart_négatif limit 5)
select siret, periode, sum(part_ouvrière) as part_ouvrière, sum(part_patronale) as part_patronale
from débits
group by siret, periode;