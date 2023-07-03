package campaign

import "github.com/gin-gonic/gin"

func successHandler(ctx *gin.Context) {
	actionHandlerFunc("success")(ctx)
}
