create function permissions (
	in roles_user text[], 
	in roles_enterprise text[],
	in first_alert_entreprise text, 
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
	coalesce($2, '{}'::text[]) && coalesce($1, '{}'::text[]) as visible,
	coalesce($4 = any(coalesce($1, '{}'::text[])), false) as in_zone,
	((coalesce($2, '{}'::text[]) && coalesce($1, '{}'::text[]) and $3 is not null) or coalesce($5, false)) and 'score' = any(coalesce($1, '{}'::text[])) score,
	((coalesce($2, '{}'::text[]) && coalesce($1, '{}'::text[]) and $3 is not null) or coalesce($5, false)) and 'urssaf' = any(coalesce($1, '{}'::text[])) urssaf,
	((coalesce($2, '{}'::text[]) && coalesce($1, '{}'::text[]) and $3 is not null) or coalesce($5, false)) and 'dgefp' = any(coalesce($1, '{}'::text[])) dgefp,
	((coalesce($2, '{}'::text[]) && coalesce($1, '{}'::text[]) and $3 is not null) or coalesce($5, false)) and 'bdf' = any(coalesce($1, '{}'::text[])) bdf
$$ language sql immutable
