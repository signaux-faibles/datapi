-- 1 onglet avec une ligne par jour
--
--     Colonne 1 : date du jour
--     colonne 2 : nombre de recherches
--     colonne 3 : nombre de fiches vues
--     colonne 4 : nombre d’actions

CREATE INDEX IF NOT EXISTS idx_v_log_email_segment ON v_log((tokencontent ->> 'email'),(tokencontent ->> 'segment'));

WITH temp_logs AS (SELECT DATE_TRUNC('day', date_add)                     AS jour,
                          path,
                          COALESCE(tokencontent ->> 'segment'::text, '-') as segment
                   FROM v_log
                   WHERE date_add >= $1
                     and date_add < $2)
SELECT username,
       count(distinct session) AS visites,
       count(path)             AS actions,
       segment
FROM temp_logs
GROUP BY username, segment;
