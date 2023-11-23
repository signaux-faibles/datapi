drop table if exists tmp_cotisation;
create table tmp_cotisation
(
    siret text,
    numéro_compte text,
    période_start date,
    période_end date,
    encaissé real,
    du real
);