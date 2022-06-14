-- ajoute la table concernant les infos PGE
drop table if exists entreprise_pge;

create table entreprise_pge
(
    siren text,
    actif boolean
);

-- alter table entreprise_pge
--     owner to datapi;

-- ajuste la fonction `permissions` pour utiliser la nouvelle permission `pge`
drop function if exists permissions;
create or replace function permissions(roles_user text[],
                                       roles_enterprise text[],
                                       first_alert_entreprise text,
                                       departement text,
                                       followed boolean,
                                       OUT visible boolean,
                                       OUT in_zone boolean,
                                       OUT score boolean,
                                       OUT urssaf boolean,
                                       OUT dgefp boolean,
                                       OUT bdf boolean,
                                       OUT pge boolean) returns record
    immutable
    language sql
as
$$
select coalesce($2, '{}'::text[]) && coalesce($1, '{}'::text[]) as visible,
       coalesce($4 = any (coalesce($1, '{}'::text[])), false)   as in_zone,
       ((coalesce($2, '{}'::text[]) && coalesce($1, '{}'::text[]) and $3 is not null) or coalesce($5, false)) and
       'score' = any (coalesce($1, '{}'::text[]))                  score,
       ((coalesce($2, '{}'::text[]) && coalesce($1, '{}'::text[]) and $3 is not null) or coalesce($5, false)) and
       'urssaf' = any (coalesce($1, '{}'::text[]))                 urssaf,
       ((coalesce($2, '{}'::text[]) && coalesce($1, '{}'::text[]) and $3 is not null) or coalesce($5, false)) and
       'dgefp' = any (coalesce($1, '{}'::text[]))                  dgefp,
       ((coalesce($2, '{}'::text[]) && coalesce($1, '{}'::text[]) and $3 is not null) or coalesce($5, false)) and
       'bdf' = any (coalesce($1, '{}'::text[]))                    bdf,
       ((coalesce($2, '{}'::text[]) && coalesce($1, '{}'::text[]) and $3 is not null) or coalesce($5, false)) and
       'pge' = any (coalesce($1, '{}'::text[]))                    pge
$$;

-- alter function permissions(
--     text[],
--     text[],
--     text,
--     text,
--     boolean,
--     out boolean,
--     out boolean,
--     out boolean,
--     out boolean,
--     out boolean,
--     out boolean,
--     out boolean) owner to datapi;

