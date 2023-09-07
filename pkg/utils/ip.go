package utils

import (
	"log/slog"
	"net"
	"strings"

	"github.com/spf13/viper"
)

func AcceptIP(ip string) bool {
	var whitelist = viper.GetString("adminWhitelist")
	// parsing de la liste configurée
	ips := parseWhitelist(whitelist)
	return isIPWhitelisted(ips, ip)
}

func isIPWhitelisted(whitelist []net.IP, ip string) bool {
	clientIP := net.ParseIP(ip)
	if len(whitelist) == 0 && clientIP.IsLoopback() {
		slog.Warn("Appel loopback", slog.String("fromIp", ip))
		return true
	}
	return Contains(whitelist, clientIP)
}

func parseWhitelist(whitelist string) []net.IP {
	if whitelist == "" {
		return nil
	}
	var r []net.IP
	var malformed []string
	ips := strings.Split(whitelist, ",")
	for _, current := range ips {
		ip := net.ParseIP(strings.TrimSpace(current))
		if ip == nil {
			malformed = append(malformed, current)
		} else {
			r = append(r, ip)
		}
	}
	if len(malformed) > 0 {
		slog.Warn("IP whitelistées non prises en compte", slog.String("whitelist", whitelist))
	}
	return r
}
