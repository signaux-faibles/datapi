create temporary table tmp_follow_wekan on commit drop as
with sirets as (select unnest($1::text[]) as siret),
     follow as (select siret from etablissement_follow f where f.username = $2 and active)
select case when f.siret is null then 'follow' else 'unfollow' end as todo,
       coalesce(f.siret, s.siret) as siret
from follow f
       full join sirets s on s.siret = f.siret
where f.siret is null or s.siret is null;
