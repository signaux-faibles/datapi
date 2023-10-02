with
    sirets as (select unnest($2::text[])::text as siret),
    zones as (select key as slug, ARRAY(SELECT jsonb_array_elements_text(value)) as zone
               from jsonb_each($3::jsonb)),
     zone as (select flatmap(z.zone) as zone
                   from campaign c
                            inner join zones z on z.slug ~ c.wekan_domain_regexp
                   where c.id = $1)
select distinct s.siret,
       case when v.siret is null then 'nonSiret'
            when not(v.code_departement = any (z.zone)) then 'horsZone'
            when ce.id is not null then 'present'
            else 'ok'
       end,
       v.raison_sociale,
       v.code_departement
from sirets s
         left join v_summaries v on v.siret = s.siret
         left join zone z on true
         left join campaign_etablissement ce on ce.siret = s.siret and ce.id_campaign = $1
where s.siret = any ($2)