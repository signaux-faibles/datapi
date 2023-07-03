package kanban

import (
	"context"
	"datapi/pkg/core"
	"datapi/pkg/db"
	"github.com/signaux-faibles/libwekan"
)

func followSiretsFromWekan(ctx context.Context, username libwekan.Username, sirets []string) error {
	tx, err := db.Get().Begin(ctx)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, sqlCreateTmpFollowWekan, sirets, username); err != nil {
		return core.DatabaseExecutionError{QueryIdentifier: "sqlCreateTmpFollowWekan"}
	}
	if _, err := tx.Exec(ctx, sqlFollowFromTmp, username); err != nil {
		return core.DatabaseExecutionError{QueryIdentifier: "sqlFollowFromTmp"}
	}
	return tx.Commit(ctx)
}
