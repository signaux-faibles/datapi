package kanban

import _ "embed"

//go:embed sql/createTmpFollowWekan.sql
var sqlCreateTmpFollowWekan string

//go:embed sql/followFromTmp.sql
var sqlFollowFromTmp string

//go:embed sql/getCards.sql
var SqlGetCards string

//go:embed sql/getFollow.sql
var SqlGetFollow string

//go:embed sql/getDbExport.sql
var sqlGetDbExport string

//go:embed sql/getDbExportWithoutCards.sql
var sqlGetDbExportWithoutCards string
