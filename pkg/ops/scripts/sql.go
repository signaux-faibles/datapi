package scripts

import (
	_ "embed"
)

//go:embed sql/populate_v_tables.sql
var sqlPopulateVTables string

//go:embed sql/aggUrssaf.sql
var sqlAggregationUrssaf string

type Execution struct {
	label string
	sql   string
}

var ExecuteRefreshVTables = Execution{
	label: "rafraîchit les vtables",
	sql:   sqlPopulateVTables,
}

var ExecuteAggregationURSSAF = Execution{
	label: "aggrège les données temporaires URSSAF",
	sql:   sqlAggregationUrssaf,
}

// pour les tests
var Wait5Seconds = Execution{
	label: "attends 5",
	sql:   "SELECT pg_sleep(5);",
}

// pour les tests
var Fail = Execution{
	label: "sql invalide",
	sql:   "sql invalide",
}
