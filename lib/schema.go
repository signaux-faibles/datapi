package dalib

var schema = `-- Settings

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;
SET default_tablespace = '';
SET default_with_oids = false;

-- Extensions

CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;
COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';

CREATE EXTENSION IF NOT EXISTS hstore WITH SCHEMA public;
COMMENT ON EXTENSION hstore IS 'data type for storing sets of (key, value) pairs';

CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;
COMMENT ON EXTENSION pgcrypto IS 'cryptographic functions';

-- Functions

CREATE FUNCTION public.last_agg(anyelement, anyelement) RETURNS anyelement
    LANGUAGE sql IMMUTABLE STRICT
    AS $_$
    SELECT $2;
	$_$;

CREATE FUNCTION public.url_decode(data text) RETURNS bytea
    LANGUAGE sql
    AS $$
WITH t AS (SELECT translate(data, '-_', '+/') AS trans),
     rem AS (SELECT length(t.trans) % 4 AS remainder FROM t) -- compute padding size
    SELECT decode(
        t.trans ||
        CASE WHEN rem.remainder > 0
           THEN repeat('=', (4 - rem.remainder))
           ELSE '' END,
    'base64') FROM t, rem;
$$;

-- Aggregates

CREATE AGGREGATE public.array_cat_agg(anyarray) (
    SFUNC = array_cat,
    STYPE = anyarray
);

CREATE AGGREGATE public.last(anyelement) (
    SFUNC = public.last_agg,
    STYPE = anyelement
);

CREATE SEQUENCE public.bucket_id
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE public.bucket (
    id bigint DEFAULT nextval('public.bucket_id'::regclass) NOT NULL,
    key public.hstore,
    add_date timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    release_date timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    scope text[],
    value jsonb,
    author text
);

ALTER TABLE ONLY public.bucket
    ADD CONSTRAINT bucket_pkey PRIMARY KEY (id);

CREATE INDEX bucket_key_btree ON public.bucket USING btree (key);

CREATE INDEX bucket_key_gin ON public.bucket USING gin (key);

CREATE SEQUENCE public.logs_id
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE public.logs (
    id bigint DEFAULT nextval('public.logs_id'::regclass) NOT NULL,
    reader text,
    read_date timestamp without time zone,
    query jsonb
);

ALTER TABLE ONLY public.logs
    ADD CONSTRAINT logs_pkey PRIMARY KEY (id);

CREATE TABLE public.cache_listing (
    scope text[],
    hash text NOT NULL
);

ALTER TABLE ONLY public.cache_listing
    ADD CONSTRAINT cache_listing_pkey PRIMARY KEY (hash);
`
