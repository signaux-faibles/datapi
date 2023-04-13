package core

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/signaux-faibles/datapi/src/utils"
	"github.com/spf13/viper"
	"log"
	"net"
	"net/http"
	"strings"
)

var errorAdminWhitelistNotConfigured = errors.New("la white list des IP n'est pas configurée, voir config.toml, entrée `adminWhitelist`")

// AdminAuthMiddleware stoppe la requête si l'ip client n'est pas contenue dans la whitelist
func AdminAuthMiddleware(c *gin.Context) {
	clientIP := net.ParseIP(c.ClientIP())
	if clientIP.IsLoopback() {
		log.Printf("Warning : Appel loopback depuis %s\n", c.ClientIP())
		return
	}
	var whitelist = viper.GetString("adminWhitelist")
	if whitelist == "" {
		c.AbortWithError(http.StatusInternalServerError, errorAdminWhitelistNotConfigured)
		return
	}
	whitelistedIPs := strings.Split(whitelist, ",")
	if !utils.Contains(whitelistedIPs, c.ClientIP()) {
		log.Printf("Erreur : Une tentative de connexion depuis %s, ce qui n'est pas autorisé dans `adminWhitelist`, voir config.toml\n", c.ClientIP())
		c.AbortWithStatus(http.StatusForbidden)
	}
}
