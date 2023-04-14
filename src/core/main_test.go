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

const loopbackIPv4 = "127.0.0.1"
const loopbackIPv6 = "::1"

func init() {
	fake = faker.New()
}

func TestAdminAuthMiddleware_withoutWhitelist_acceptOnlyLoopback(t *testing.T) {
	ass := assert.New(t)
	for _, ip := range generateRandomIPs() {
		ginContext := buildGinContextWithIP(ip)
		AdminAuthMiddleware(ginContext)
		ass.True(ginContext.IsAborted())
		ass.Equal(http.StatusForbidden, ginContext.Writer.Status())
	}

	// test loopback ip
	for _, loopback := range []string{loopbackIPv4, loopbackIPv6} {
		ginContext := buildGinContextWithIP(loopback)
		AdminAuthMiddleware(ginContext)
		ass.False(ginContext.IsAborted())
		ass.Equal(http.StatusOK, ginContext.Writer.Status())
	}
}

func TestAdminAuthMiddleware_withWhitelist_acceptOnlyWhitelistedIPs(t *testing.T) {
	ass := assert.New(t)

	whitelist := generateRandomIPs()
	configureAdminWhitelist(whitelist...)

	for _, ip := range whitelist {
		ginContext := buildGinContextWithIP(ip)
		AdminAuthMiddleware(ginContext)
		ass.False(ginContext.IsAborted())
		ass.Equal(http.StatusOK, ginContext.Writer.Status())
	}

	for _, loopback := range append(generateRandomIPs(), loopbackIPv6, loopbackIPv4) {
		ginContext := buildGinContextWithIP(loopback)
		AdminAuthMiddleware(ginContext)
		ass.True(ginContext.IsAborted())
		ass.Equal(http.StatusForbidden, ginContext.Writer.Status())
	}
}

func buildGinContextWithIP(ip string) *gin.Context {
	ginContext, _ := gin.CreateTestContext(httptest.NewRecorder())
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

func configureAdminWhitelist(ips ...string) {
	whitelistProperty := strings.Join(ips, ", ")
	viper.GetViper().SetDefault("adminWhitelist", whitelistProperty)
	log.Println("la propriété `adminWhitelist` ->", whitelistProperty)
}

func generateRandomIPs() []string {
	size := fake.Int()%5 + 1
	var r []string
	for i := 0; i < size; i++ {
		if fake.Bool() {
			r = append(r, fake.Internet().Ipv4())
		} else {
			r = append(r, fake.Internet().Ipv6())
		}
	}
	return r
}
