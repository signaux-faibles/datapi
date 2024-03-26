package imports

import (
	_ "embed"

	"datapi/pkg/ops/scripts"
)

//go:embed sql/aggUrssaf.sql
var sqlAggregationUrssaf string

var ExecuteAggregationURSSAF = scripts.Script{
	Label: "aggrège les données temporaires URSSAF",
	SQL:   sqlAggregationUrssaf,
}
