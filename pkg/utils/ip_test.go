package utils

import (
	"math/rand"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/jaswdr/faker"
	"github.com/spf13/viper"
)

const loopbackIPv4 = "127.0.0.1"
const loopbackIPv6 = "::1"

var fake faker.Faker

func init() {
	fake = faker.NewWithSeed(rand.NewSource(time.Now().UnixMicro()))
	ConfigureLogLevel("debug")
}

func Test_AcceptIP(t *testing.T) {
	knownIP := generateRandomIP()
	type args struct {
		whitelist []string
		ip        string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"cas pas de whitelist (nil) : refusé", args{nil, generateRandomIP()}, false},
		{"cas pas de whitelist (vide) : refusé", args{[]string{}, generateRandomIP()}, false},
		{"cas pas de whitelist et vient du loopback en ipv4 : accepté", args{[]string{}, loopbackIPv4}, true},
		{"cas pas de whitelist et vient du loopback en ipv6 : accepté", args{nil, loopbackIPv6}, true},
		{"cas whitelist et vient du loopback en ipv4 : refusé", args{[]string{generateRandomIP()}, loopbackIPv4}, false},
		{"cas whitelist et vient du loopback en ipv6 : refusé", args{[]string{generateRandomIP()}, loopbackIPv6}, false},
		{"cas whitelist et vient d'une adresse non whitelistée : refusé", args{generateRandomIPs(), generateRandomIP()}, false},
		{"cas whitelist et vient d'une adresse whitelistée : accepté", args{append(generateRandomIPs(), knownIP), knownIP}, true},
		{"whitelist ipv6 réécrite : accepté", args{[]string{"2345:425:2ca1::567:5673:23b5"}, "2345:425:2ca1::567:5673:23b5"}, true},
		{"whitelist ipv6 non réécrite : accepté", args{[]string{string("2345:425:2ca1::567:5673:23b5")}, "2345:0425:2CA1:0000:0000:0567:5673:23b5"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Set("adminWhiteList", tt.args.whitelist)
			if got := AcceptIP(tt.args.ip); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseWhitelist() -> got %v, want %v", got, tt.want)
			}
		})
	}
}

func generateRandomIPs() []string {
	size := fake.Int()%5 + 1
	var r []string
	for i := 0; i < size; i++ {
		r = append(r, generateRandomIP())
	}
	return r
}

func toIP(input []string) []net.IP {
	if input == nil {
		return nil
	}
	return Convert(input, net.ParseIP)
}

func generateRandomIP() string {
	if fake.Bool() {
		return fake.Internet().Ipv4()
	}
	return fake.Internet().Ipv6()
}
