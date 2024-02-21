package refresh

import (
	_ "embed"
)

//go:embed sql/populate_v_tables.sql
var SQLPopulateVTables string

//go:embed sql/aggUrssaf.sql
var SQLAggregationUrssaf string
