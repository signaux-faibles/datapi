-- View: public.v_alert_entreprise
DROP MATERIALIZED VIEW IF EXISTS public.v_summaries;

DROP MATERIALIZED VIEW IF EXISTS public.v_alert_entreprise;

CREATE MATERIALIZED VIEW IF NOT EXISTS public.v_alert_entreprise
TABLESPACE pg_default
AS
 WITH first_lists AS (
         SELECT DISTINCT sc.siren,
            first(sc.libelle_liste) OVER (window_siren) AS first_list,
            first(sc.libelle_liste) FILTER (WHERE sc.alert = 'Alerte seuil F1'::text) OVER (window_siren) AS first_red_list
           FROM score0 sc
          WHERE sc.alert = ANY (ARRAY['Alerte seuil F1'::text, 'Alerte seuil F2'::text])
          WINDOW window_siren AS (PARTITION BY sc.siren ORDER BY sc.batch)
        )
 SELECT first_lists.siren,
    max(first_lists.first_list) as first_list,
    max(first_lists.first_red_list) as first_red_list
  FROM first_lists
  WHERE first_lists.first_list IS NOT NULL
  group by siren
WITH DATA;

CREATE INDEX idx_v_alert_entreprise_siren
    ON public.v_alert_entreprise USING btree
    (siren COLLATE pg_catalog."default")
    TABLESPACE pg_default;
    
CREATE MATERIALIZED VIEW IF NOT EXISTS public.v_summaries
TABLESPACE pg_default
AS
 WITH last_liste AS (
         SELECT first(liste.libelle ORDER BY liste.batch DESC, liste.algo DESC) AS last_liste
           FROM liste
        )
 SELECT r.roles,
    et.siret,
    et.siren,
    en.raison_sociale,
    et.commune,
    d.libelle AS libelle_departement,
    et.departement AS code_departement,
    s.score AS valeur_score,
    s.detail AS detail_score,
    coalesce(aen.first_list = l.last_liste or aen.first_red_list = l.last_liste, false) AS first_alert,
    aen.first_list AS first_list_entreprise,
    aen.first_red_list AS first_red_list_entreprise,
    aet.last_list,
    aet.last_alert,
    l.last_liste AS liste,
    di.chiffre_affaire,
    di.prev_chiffre_affaire,
    di.arrete_bilan,
    di.exercice_diane,
    di.variation_ca,
    di.resultat_expl,
    di.prev_resultat_expl,
    di.excedent_brut_d_exploitation,
    di.prev_excedent_brut_d_exploitation,
    ef.effectif,
    sum(ef.effectif) FILTER (WHERE COALESCE(et.etat_administratif, 'A'::text) = 'A'::text) OVER (PARTITION BY en.siren) AS effectif_entreprise,
    max(ef.periode) FILTER (WHERE COALESCE(et.etat_administratif, 'A'::text) <> 'A'::text) OVER (PARTITION BY en.siren) AS date_entreprise,
    ef.periode AS date_effectif,
    n.libelle_n5,
    n.libelle_n1,
    et.code_activite,
    COALESCE(ep.last_procol, 'in_bonis'::text) AS last_procol,
    ep.date_effet AS date_last_procol,
    COALESCE(ap.ap, false) AS activite_partielle,
    ap.heure_consomme AS apconso_heure_consomme,
    ap.montant AS apconso_montant,
    u.dette[1] > u.dette[2] OR u.dette[2] > u.dette[3] AS hausse_urssaf,
    u.dette[1] AS dette_urssaf,
    u.periode_urssaf,
    u.presence_part_salariale,
    s.alert,
    et.siege,
    g.raison_sociale AS raison_sociale_groupe,
    ti.code_commune IS NOT NULL AS territoire_industrie,
    ti.code_terrind AS code_territoire_industrie,
    ti.libelle_terrind AS libelle_territoire_industrie,
    cj.libelle AS statut_juridique_n3,
    cj2.libelle AS statut_juridique_n2,
    cj1.libelle AS statut_juridique_n1,
    et.creation AS date_ouverture_etablissement,
    en.creation AS date_creation_entreprise,
    COALESCE(sco.secteur_covid, 'nonSecteurCovid'::text) AS secteur_covid,
        CASE
            WHEN COALESCE(en.etat_administratif, 'A'::text) = 'C'::text THEN 'F'::text
            ELSE COALESCE(et.etat_administratif, 'A'::text)
        END AS etat_administratif,
    en.etat_administratif AS etat_administratif_entreprise
   FROM last_liste l
     JOIN etablissement0 et ON true
     JOIN entreprise0 en ON en.siren::text = et.siren::text
     JOIN v_etablissement_raison_sociale etrs ON etrs.id_etablissement = et.id
     JOIN v_roles r ON et.siren::text = r.siren::text
     JOIN departements d ON d.code = et.departement
     LEFT JOIN v_naf n ON n.code_n5 = et.code_activite
     LEFT JOIN secteurs_covid sco ON sco.code_activite = et.code_activite
     LEFT JOIN score0 s ON et.siret::text = s.siret AND s.libelle_liste = l.last_liste
     LEFT JOIN v_alert_etablissement aet ON aet.siret = et.siret::text
     LEFT JOIN v_alert_entreprise aen ON aen.siren = et.siren::text
     LEFT JOIN v_last_effectif ef ON ef.siret::text = et.siret::text
     LEFT JOIN v_hausse_urssaf u ON u.siret::text = et.siret::text
     LEFT JOIN v_apdemande ap ON ap.siret::text = et.siret::text
     LEFT JOIN v_last_procol ep ON ep.siren::text = et.siren::text
     LEFT JOIN v_diane_variation_ca di ON di.siren::text = et.siren::text
     LEFT JOIN entreprise_ellisphere0 g ON g.siren::text = et.siren::text
     LEFT JOIN terrind ti ON ti.code_commune = et.code_commune
     LEFT JOIN categorie_juridique cj ON cj.code = en.statut_juridique
     LEFT JOIN categorie_juridique cj2 ON "substring"(cj.code, 0, 3) = cj2.code
     LEFT JOIN categorie_juridique cj1 ON "substring"(cj.code, 0, 2) = cj1.code
WITH DATA;


CREATE INDEX idx_v_summaries_order_raison_sociale
    ON public.v_summaries USING btree
    (raison_sociale COLLATE pg_catalog."default", siret COLLATE pg_catalog."default")
    TABLESPACE pg_default;
CREATE INDEX idx_v_summaries_raison_sociale
    ON public.v_summaries USING gin
    (siret COLLATE pg_catalog."default" gin_trgm_ops, raison_sociale COLLATE pg_catalog."default" gin_trgm_ops)
    TABLESPACE pg_default;
CREATE INDEX idx_v_summaries_roles
    ON public.v_summaries USING gin
    (roles COLLATE pg_catalog."default")
    TABLESPACE pg_default;
CREATE INDEX idx_v_summaries_score
    ON public.v_summaries USING btree
    (valeur_score DESC, siret COLLATE pg_catalog."default", siege, effectif, effectif_entreprise, code_departement COLLATE pg_catalog."default", last_procol COLLATE pg_catalog."default", chiffre_affaire, etat_administratif COLLATE pg_catalog."default", first_alert, first_list_entreprise, first_red_list_entreprise COLLATE pg_catalog."default")
    TABLESPACE pg_default
    WHERE alert <> 'Pas d''alerte'::text;
CREATE INDEX idx_v_summaries_siren
    ON public.v_summaries USING btree
    (siren COLLATE pg_catalog."default")
    TABLESPACE pg_default;
CREATE INDEX idx_v_summaries_siret
    ON public.v_summaries USING btree
    (siret COLLATE pg_catalog."default")
    TABLESPACE pg_default;