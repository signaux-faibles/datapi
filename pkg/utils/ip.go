package utils

import (
	"log/slog"
	"net"
	"strings"

	"github.com/spf13/viper"
)

func AcceptIP(ip string) bool {
	whiteList := viper.GetStringSlice("adminWhitelist")
	clientIP := net.ParseIP(ip)
	if len(whiteList) == 0 && clientIP.IsLoopback() {
		slog.Warn("pas de whitelist configurée mais appel loopback, appel accepté", slog.String("fromIp", ip))
		return true
	}
	slog.Debug("admin white liste configurée", slog.Any("adminWhitelist", whiteList))
	// parsing de la liste configurée
	ips := parseWhitelist(whiteList)
	return Contains(ips, clientIP)
}

func parseWhitelist(whitelist []string) []net.IP {
	var r []net.IP
	var malformed []any
	for _, current := range whitelist {
		ip := net.ParseIP(strings.TrimSpace(current))
		if ip == nil {
			malformed = append(malformed, current)
		} else {
			r = append(r, ip)
		}
	}
	if len(malformed) > 0 {
		slog.Warn("IP whitelistées non prises en compte", slog.Any("whitelist", whitelist))
	}
	return r
}
