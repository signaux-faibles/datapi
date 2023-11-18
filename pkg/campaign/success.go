package campaign

import (
	"datapi/pkg/core"
	"github.com/gin-gonic/gin"
)

func successHandler(ctx *gin.Context, kanbanService core.KanbanService) {
	actionHandlerFunc("success", kanbanService)(ctx)
}
