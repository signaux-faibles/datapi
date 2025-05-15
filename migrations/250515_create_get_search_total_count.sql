DROP FUNCTION IF EXISTS public.get_search_total_count(
    text[], int4, int4, text, text, text, bool, bool, text, bool, text, bool,
    text[], text[], bool, int4, int4, text[], text[], int4, int4, int4, int4, text[], text
);

CREATE OR REPLACE FUNCTION public.get_search_total_count(
    roles_users text[], nblimit integer, nboffset integer, libelle_liste text,
    siret_expression text, raison_sociale_expression text, ignore_roles boolean,
    ignore_zone boolean, username text, siege_uniquement boolean, order_by text,
    alert_only boolean, last_procol text[], departements text[], suivi boolean,
    effectif_min integer, effectif_max integer, sirens text[], activites text[],
    effectif_min_entreprise integer, effectif_max_entreprise integer, ca_min integer,
    ca_max integer, exclude_secteurs_covid text[], etat_administratif text
)
RETURNS TABLE(total_count bigint)
LANGUAGE sql
IMMUTABLE
AS $function$
SELECT
    COUNT(*)
FROM v_summaries s
    LEFT JOIN etablissement_follow f ON f.active AND f.siret = s.siret AND f.username = $9
    LEFT JOIN v_entreprise_follow fe ON fe.siren = s.siren AND fe.username = $9
    LEFT JOIN v_naf n ON n.code_n5 = s.code_activite
WHERE
    (s.raison_sociale ILIKE $6 OR s.siret ILIKE $5)
    AND (s.roles && $1 OR $7)
    AND (s.code_departement = ANY($1) OR $8)
    AND (s.code_departement = ANY($14) OR $14 IS NULL)
    AND (s.effectif >= $16 OR $16 IS NULL)
    AND (s.effectif_entreprise >= $20 OR $20 IS NULL)
    AND (s.effectif_entreprise <= $21 OR $21 IS NULL)
    AND (s.chiffre_affaire >= $22 OR $22 IS NULL)
    AND (s.chiffre_affaire <= $23 OR $23 IS NULL)
    AND (n.code_n1 = ANY($19) OR $19 IS NULL)
    AND (s.siege OR NOT $10)
    AND (NOT (s.secteur_covid = ANY($24)) OR $24 IS NULL)
    AND (s.etat_administratif = $25 OR $25 IS NULL)
;
$function$;
