drop aggregate if exists flatmap(anycompatiblearray);
create aggregate flatmap(anycompatiblearray) (
  sfunc = array_cat,
  stype = anycompatiblearray,
  initcond = '{}'
);
