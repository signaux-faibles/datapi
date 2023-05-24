package utils

import (
	"github.com/spf13/viper"
	"log"
	"net"
	"strings"
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
		log.Printf("Warning : Appel loopback depuis %s\n", ip)
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
		log.Printf("Warning : IP whitelistées non prises en compte %s\n", malformed)
	}
	return r
}
