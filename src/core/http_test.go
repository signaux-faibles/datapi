package core

import (
	"github.com/gin-gonic/gin"
	"github.com/jaswdr/faker"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

var fake faker.Faker

func init() {
	fake = faker.New()
}

func TestAdminAuthMiddleware_acceptRequestFromLoopback(t *testing.T) {
	ass := assert.New(t)
	for _, loopback := range []string{"127.0.0.1", "::1", "127.255.31.1"} {
		ginContext := buildRequestForIP(loopback)
		AdminAuthMiddleware(ginContext)
		ass.False(ginContext.IsAborted())
		ass.Equal(http.StatusOK, ginContext.Writer.Status())
	}
}

func TestAdminAuthMiddleware_withoutWhitelist(t *testing.T) {
	ass := assert.New(t)
	for _, loopback := range []string{fake.Internet().Ipv4(), fake.Internet().Ipv6()} {
		ginContext := buildRequestForIP(loopback)
		AdminAuthMiddleware(ginContext)
		ass.True(ginContext.IsAborted())
		ass.Equal(http.StatusInternalServerError, ginContext.Writer.Status())
		ass.Len(ginContext.Errors, 1)
		ass.ErrorIs(ginContext.Errors[0], errorAdminWhitelistNotConfigured)
	}
}

func TestAdminAuthMiddleware_withNotWhitelistedIPs(t *testing.T) {
	ass := assert.New(t)
	generateWhitelist()

	randomIPs := []string{fake.Internet().Ipv4(), fake.Internet().Ipv6()}
	for _, ip := range randomIPs {
		ginContext := buildRequestForIP(ip)
		AdminAuthMiddleware(ginContext)
		ass.True(ginContext.IsAborted())
		ass.Equal(http.StatusForbidden, ginContext.Writer.Status())
	}
}

func TestAdminAuthMiddleware_withWhitelistedIPs(t *testing.T) {
	ass := assert.New(t)
	whitelist := generateWhitelist()

	for _, ip := range whitelist {
		ginContext := buildRequestForIP(ip)
		AdminAuthMiddleware(ginContext)
		ass.False(ginContext.IsAborted())
		ass.Equal(http.StatusOK, ginContext.Writer.Status())
	}
}

func generateWhitelist() []string {
	whitelist := []string{fake.Internet().Ipv4(), fake.Internet().Ipv6()}
	whitelistProperty := strings.Join(whitelist, ", ")
	viper.GetViper().SetDefault("adminWhitelist", whitelistProperty)
	log.Println("la propriété `adminWhitelist` ->", whitelistProperty)
	return whitelist
}

func buildRequestForIP(ip string) *gin.Context {
	ginContext := gin.CreateTestContextOnly(httptest.NewRecorder(), gin.Default())
	addr := net.ParseIP(ip)
	url, _ := url.Parse(fake.Internet().URL())
	req := &http.Request{
		URL:    url,
		Header: make(http.Header),
	}
	if ip4 := addr.To4(); ip4 != nil {
		req.RemoteAddr = ip4.String() + ":"
	} else {
		req.RemoteAddr = "[" + addr.String() + "]:"
	}
	ginContext.Request = req
	return ginContext
}
