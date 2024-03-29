package core

import (
	"fmt"
)

type ForbiddenError struct {
	Reason string
}

func (e ForbiddenError) Error() string {
	return fmt.Sprintf("accès interdit: %s", e.Reason)
}

type UnknownBoardError struct {
	BoardIdentifier string
}

func (e UnknownBoardError) Error() string {
	return fmt.Sprintf("aucun tableau avec l'identifiant `%s`", e.BoardIdentifier)
}

type UnknownListError struct {
	ListIdentifier string
}

func (e UnknownListError) Error() string {
	return fmt.Sprintf("aucune liste trouvée avec l'identifiant `%s`", e.ListIdentifier)
}

type UnknownCardError struct {
	CardIdentifier string
}

func (e UnknownCardError) Error() string {
	return fmt.Sprintf("aucune carte trouvée avec l'identifiant `%s`", e.CardIdentifier)
}

type DatabaseExecutionError struct {
	QueryIdentifier string
}

func (e DatabaseExecutionError) Error() string {
	return fmt.Sprintf("la requête `%s` a rencontré une erreur", e.QueryIdentifier)
}
