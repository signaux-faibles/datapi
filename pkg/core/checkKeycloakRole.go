package core

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"datapi/pkg/utils"
)

type MandatoryRoleChecker struct {
	mandatoryRoles []string
	needAll        bool
}

func NewSimpleMandatoryRoleChecker(roles ...string) MandatoryRoleChecker {
	return MandatoryRoleChecker{mandatoryRoles: roles, needAll: false}
}

func (mrc MandatoryRoleChecker) CheckRoleMiddleware(c *gin.Context) {
	roles := scopeFromContext(c)
	if roles == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, "erreur, les claims n'ont pas été correctement défini dans le contexte")
	}
	containsAll := true
	for _, current := range mrc.mandatoryRoles {
		contains := utils.Contains(roles, current)
		if contains && !mrc.needAll {
			return
		}
		containsAll = containsAll && contains
	}
	if containsAll {
		return
	}
	c.AbortWithStatusJSON(http.StatusForbidden, "erreur, il manque un rôle")
}
