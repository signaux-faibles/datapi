--
-- Name: logs_id; Type: SEQUENCE
--

CREATE SEQUENCE IF NOT EXISTS logs_id
  START WITH 1
  INCREMENT BY 1
  NO MINVALUE
  NO MAXVALUE
  CACHE 1;

--
-- Name: logs; Type: TABLE
--

CREATE TABLE IF NOT EXISTS logs (
  id integer DEFAULT nextval('logs_id'::regclass) NOT NULL,
  date_add timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
  path text,
  method text,
  body text,
  token text
);

--
-- Name: v_log; Type: VIEW
--

create or replace view v_log(id, date_add, path, method, body, token, tokencontent) as
SELECT logs.id,
       logs.date_add,
       logs.path,
       logs.method,
       logs.body,
       logs.token,
       convert_from(decode(rpad(translate(split_part(logs.token, '.'::text, 2), '-_'::text, '+/'::text),
                                4 * ((length(split_part(logs.token, '.'::text, 2)) + 3) / 4), '='::text),
                           'base64'::text), 'UTF-8'::name)::jsonb AS tokencontent
FROM logs
WHERE logs.token <> 'fakeKeycloak'::text;

--
-- Name: v_log; Type: MATERIALIZED VIEW
--

create materialized view if not exists v_stats as
WITH stat_users AS (SELECT count(DISTINCT v_log.tokencontent ->> 'email'::text) AS users,
                           date_trunc('month'::text, v_log.date_add)            AS mois
                    FROM v_log
                    GROUP BY (date_trunc('month'::text, v_log.date_add))),
     stat_recherches AS (SELECT count(DISTINCT
                                      CASE
                                        WHEN v_log.method = 'GET'::text THEN v_log.path
                                        ELSE v_log.body::json ->> 'search'::text
                                        END || ((v_log.tokencontent ->> 'email'::text) ||
                                                date_trunc('day'::text, v_log.date_add)::text)) AS recherches,
                                date_trunc('month'::text, v_log.date_add)                         AS mois
                         FROM v_log
                         WHERE v_log.path ~~ '%/search%'::text
                         GROUP BY (date_trunc('month'::text, v_log.date_add))),
     stat_fiches AS (SELECT count(v_log.path)                         AS fiches,
                            date_trunc('month'::text, v_log.date_add) AS mois
                     FROM v_log
                     WHERE v_log.path ~~ '/etablissement/get/%'::text
                        OR v_log.path ~~ '/entreprise/%'::text
                     GROUP BY (date_trunc('month'::text, v_log.date_add))),
     stat_visites AS (SELECT count(DISTINCT date_trunc('hour'::text, v_log.date_add)::text ||
                                            (v_log.tokencontent ->> 'email'::text)) AS visites,
                             date_trunc('month'::text, v_log.date_add)              AS mois
                      FROM v_log
                      GROUP BY (date_trunc('month'::text, v_log.date_add))
                      LIMIT 100)
SELECT u.mois,
       u.users,
       r.recherches,
       f.fiches,
       v.visites
FROM stat_users u
       JOIN stat_recherches r ON r.mois = u.mois
       JOIN stat_fiches f ON f.mois = u.mois
       JOIN stat_visites v ON v.mois = u.mois;

alter materialized view v_stats owner to postgres;




