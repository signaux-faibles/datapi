package campaign

import (
	"datapi/pkg/core"
	"github.com/gin-gonic/gin"
)

func cancelHandler(ctx *gin.Context, kanbanService core.KanbanService) {
	actionHandlerFunc("cancel", kanbanService)(ctx)
}
