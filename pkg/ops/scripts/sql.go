package scripts

import (
	_ "embed"
)

type Execution struct {
	Label string
	SQL   string
}

// pour les tests
var Wait5Seconds = Execution{
	Label: "attends 5",
	SQL:   "SELECT pg_sleep(5);",
}

// pour les tests
var Fail = Execution{
	Label: "sql invalide",
	SQL:   "sql invalide",
}
