WITH temp_logs AS (SELECT date_bin('8 hours', date_add, TIMESTAMP '0001-01-01')     AS session,
                          path,
                          method,
                          COALESCE(tokencontent ->> 'email'::text, 'EMAIL INCONNU') as username,
                          COALESCE(tokencontent ->> 'segment'::text, '-')           as segment
                   FROM v_log
                   WHERE date_add >= $1
                     and date_add < $2)
SELECT username,
       count(distinct session) AS visites,
       count(path)             AS actions,
       segment
FROM temp_logs
GROUP BY username, segment
ORDER BY username;
