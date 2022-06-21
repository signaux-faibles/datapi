-- ajoute la table concernant les infos PGE

create table if not exists public.entreprise_pge
(
    siren text,
    actif boolean
);

create index if not exists idx_entreprise_pge_siren on public.entreprise_pge(siren);

-- ajuste la fonction `permissions` pour utiliser la nouvelle permission `pge`
-- on précise les arguments parce pgsql accepte le polymorphisme
drop function if exists public.permissions(roles_user text[],
                                           roles_enterprise text[],
                                           first_alert_entreprise text,
                                           departement text,
                                           followed boolean,
                                           OUT visible boolean,
                                           OUT in_zone boolean,
                                           OUT score boolean,
                                           OUT urssaf boolean,
                                           OUT dgefp boolean,
                                           OUT bdf boolean);

create or replace function public.permissions(roles_user text[],
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
                                              OUT pge boolean)
    returns record
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

drop function if exists public.f_etablissement_permissions(roles_user text[], username text);
create or replace function public.f_etablissement_permissions(roles_user text[], username text)
    returns TABLE
            (
                id       integer,
                siren    character varying,
                siret    character varying,
                followed boolean,
                visible  boolean,
                in_zone  boolean,
                score    boolean,
                urssaf   boolean,
                dgefp    boolean,
                bdf      boolean,
                pge      boolean
            )
    immutable
    language sql
as
$$
select e.id,
       e.siren,
       e.siret,
       f.siren is not null as followed,
       (permissions($1, r.roles, a.first_list, e.departement, f.siren is not null)).*
from etablissement0 e
         inner join v_roles r on r.siren = e.siren
         left join v_entreprise_follow f on f.siren = e.siren and f.username = $2
         left join v_alert_entreprise a on a.siren = e.siren
$$;
