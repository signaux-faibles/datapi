package core

//type KanbanCard struct {
//	ID           libwekan.CardID     `bson:"_id"`
//	Title        string              `bson:"title"`
//	ListID       libwekan.ListID     `bson:"listId"`
//	BoardId      libwekan.BoardID    `bson:"boardId"`
//	SwimlaneID   libwekan.SwimlaneID `bson:"swimlaneId"`
//	Members      []string            `bson:"members"`
//	Assignees    []string            `bson:"assignees"`
//	Rank         float64             `bson:"sort"`
//	Description  string              `bson:"description"`
//	StartAt      time.Time           `bson:"startAt"`
//	EndAt        *time.Time          `bson:"endAt"`
//	LastActivity time.Time           `bson:"lastActivity"`
//	LabelIds     []string            `bson:"labelIds"`
//	Archived     bool                `bson:"archived"`
//	UserID       libwekan.UserID     `bson:"userId"`
//	Comments     []string            `bson:"comments"`
//	CustomFields []struct {
//		ID    string `bson:"_id"`
//		Value string `bson:"value"`
//	}
//}
//
//func getKanbanCardFromSiret(ctx context.Context, siret string) ([]libwekan.Card, error) {
//	return wekan.SelectCardsFromCustomTextField(ctx, "SIRET", siret)
//}
//
//func createKanbanCard(ctx context.Context, kanbanCard KanbanCard) error {
//	card := libwekan.BuildCard(
//		kanbanCard.BoardId,
//		kanbanCard.ListID,
//		kanbanCard.SwimlaneID,
//		kanbanCard.Title,
//		kanbanCard.Description,
//		kanbanCard.UserID,
//	)
//	return nil
//}
