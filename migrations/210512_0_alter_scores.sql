alter table score add expl_selection_concerning jsonb;
alter table score add expl_selection_reassuring jsonb;
alter table score add macro_explain jsonb;
alter table score add micro_explain jsonb;
alter table score add macro_radar jsonb;
create or replace view score0 as
select * from score where version = 0;
