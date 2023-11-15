drop table if exists tmp_procol;
create table tmp_procol
(
    siret         text,
    date_effet    date,
    action_procol text,
    stade_procol  text
);