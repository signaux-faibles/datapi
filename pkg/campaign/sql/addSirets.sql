with sirets as (select unnest($2::text[])::text as siret),
     zones as (select key as slug, ARRAY(SELECT jsonb_array_elements_text(value)) as zone
               from jsonb_each($3::jsonb)),
     zone as (select flatmap(z.zone) as zone
              from campaign c
                       inner join zones z on z.slug ~ c.wekan_domain_regexp
              where c.id = $1),
     checkedSirets as (select distinct s.siret,
                                       case
                                           when v.siret is null then 'nonSiret'
                                           when not (v.code_departement = any (z.zone)) then 'horsZone'
                                           when ce.id is not null then 'present'
                                           else 'ok'
                                           end as status,
                                       v.raison_sociale,
                                       v.code_departement
                       from sirets s
                                left join v_summaries v on v.siret = s.siret
                                left join zone z on true
                                left join campaign_etablissement ce on ce.siret = s.siret
                                left join campaign c on c.id = ce.id_campaign and c.id = $1
                       where s.siret = any ($2)),
     insertedSirets as (insert into campaign_etablissement (id_campaign, siret, username)
         select $1, siret, $4
         from checkedSirets
         where status = 'ok'
         returning siret)
select cs.siret,
       cs.status,
       cs.raison_sociale,
       cs.code_departement,
       case
           when ins.siret is not null
               then 'inserted'
           else 'not inserted'
           end as outcome
from checkedSirets cs
         left join insertedSirets ins on ins.siret = cs.siret
