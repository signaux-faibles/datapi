package core

import (
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"datapi/pkg/test"
	"datapi/pkg/utils"
)

var fake = test.NewFaker()

const loopbackIPv4 = "127.0.0.1"
const loopbackIPv6 = "::1"

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
	uerel, _ := url.Parse(fake.Internet().URL())
	req := &http.Request{
		URL:    uerel,
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
	return generateRandomIPsNotIn(loopbackIPv6, loopbackIPv4)
}

func generateRandomIPsNotIn(forbidden ...string) []string {
	size := fake.Int()%5 + 1
	var r []string
	for i := 0; i < size; i++ {
		r = append(r, generateIPNotIn(forbidden...))
	}
	return r
}

func generateIPNotIn(forbidden ...string) string {
	var r string
	for next := true; next; next = utils.Contains(forbidden, r) {
		if fake.Bool() {
			r = fake.Internet().Ipv4()
		} else {
			r = fake.Internet().Ipv6()
		}
	}
	return r
}
