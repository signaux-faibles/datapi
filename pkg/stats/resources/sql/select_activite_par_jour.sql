-- 1 onglet avec une ligne par jour
--
--     Colonne 1 : date du jour
--     colonne 2 : nombre de recherches
--     colonne 3 : nombre de fiches vues
--     colonne 4 : nombre d’actions


WITH stat_recherches AS (SELECT count(DISTINCT
                                      CASE
                                        WHEN v_log.method = 'GET'::text THEN v_log.path
                                        ELSE v_log.body::json ->> 'search'::text
                                        END || ((v_log.tokencontent ->> 'email'::text) ||
                                                date_trunc('day'::text, v_log.date_add)::text)) AS recherches,
                                coalesce(tokencontent ->> 'email', 'EMAIL INCONNU')             as username,
                                date_trunc('day'::text, v_log.date_add)                         AS jour
                         FROM v_log
                         WHERE v_log.path ~~ '%/search%'::text
                           AND date_add >= $1
                           AND date_add < $2
                         GROUP BY (date_trunc('day'::text, v_log.date_add)),
                                  coalesce(tokencontent ->> 'email', 'EMAIL INCONNU')),
     stat_fiches AS (SELECT count(v_log.path)                                   AS fiches,
                            coalesce(tokencontent ->> 'email', 'EMAIL INCONNU') as username,
                            date_trunc('day'::text, v_log.date_add)             AS jour
                     FROM v_log
                     WHERE v_log.path ~~ '/etablissement/get/%'::text
                        OR v_log.path ~~ '/entreprise/%'::text
                       AND date_add >= $1
                       AND date_add < $2
                     GROUP BY (date_trunc('day'::text, v_log.date_add)),
                              coalesce(tokencontent ->> 'email', 'EMAIL INCONNU')),
     stat_utilisateur AS (SELECT count(v_log.path)                                   AS actions,
                                 coalesce(tokencontent ->> 'segment', '-')           as segment,
                                 coalesce(tokencontent ->> 'email', 'EMAIL INCONNU') as username,
                                 date_trunc('day'::text, v_log.date_add)             AS jour
                          FROM v_log
                          WHERE not starts_with(v_log.path, '/kanban/config')
                            AND not starts_with(path, '/campaign/stream')
                            AND date_add >= $1
                            AND date_add < $2
                          GROUP BY (date_trunc('day'::text, v_log.date_add)),
                                   coalesce(tokencontent ->> 'email', 'EMAIL INCONNU'),
                                   coalesce(tokencontent ->> 'segment', '-'))
SELECT u.jour,
       u.username,
       u.actions,
       coalesce(r.recherches, 0) as recherches,
       coalesce(f.fiches, 0)     as fiches,
       u.segment
FROM stat_utilisateur u
       LEFT JOIN stat_recherches r ON u.username = r.username AND u.jour = r.jour
       LEFT JOIN stat_fiches f ON u.username = f.username AND u.jour = f.jour;
