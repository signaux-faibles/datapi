with base as (
	select distinct read_date::date, convert_from(url_decode(split_part(reader,'.', 2)), 'UTF-8') as jwt from logs where reader != 'system' 
)
select read_date::date, 
jwt::json->>'name' as utilisateur,
jwt::json->>'email' as login,
 ((jwt::json->>'resource_access')::json->>'signauxfaibles')::json->>'roles' as permissions,
 count(*) as nb_requetes
from base
group by read_date::date, jwt::json->>'name',jwt::json->>'email',
 ((jwt::json->>'resource_access')::json->>'signauxfaibles')::json->>'roles' 
order by read_date::date
