package campaign

import "github.com/gin-gonic/gin"

func cancelHandler(ctx *gin.Context) {
	actionHandlerFunc("cancel")(ctx)
}
