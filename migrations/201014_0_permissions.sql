create function permissions (
	in roles_user text[], 
	in roles_enterprise text[],
	in first_alert_entreprise text, 
	in first_alert_etablissement text, 
	in departement text, 
	in followed boolean,
	out visible boolean,
	out in_zone boolean,
	out score boolean,
	out urssaf boolean,
	out dgefp boolean,
	out bdf boolean
) as $$
select 
	$2 && $1 as visible,
	$5 = any($1) as in_zone,
	(($2 && $1 and $3 is not null) or $6) and 'score' = any($1) score,
	(($2 && $1 and $3 is not null) or $6) and 'urssaf' = any($1) urssaf,
	(($2 && $1 and $3 is not null) or $6) and 'dgefp' = any($1) dgefp,
	(($2 && $1 and $3 is not null) or $6) and 'bdf' = any($1) bdf
$$ language sql immutable

