---
--- Logs database creation
---

CREATE DATABASE datapilogs OWNER postgres;

\c datapilogs

CREATE SEQUENCE public.logs_id_seq;

CREATE TABLE public.logs (
    id integer NOT NULL DEFAULT nextval('public.logs_id_seq'::regclass),
    date_add timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    path text,
    method text,
    body text,
    token text
);

ALTER TABLE public.logs OWNER TO postgres;