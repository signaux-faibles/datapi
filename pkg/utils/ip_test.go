package utils

import (
	"github.com/jaswdr/faker"
	"net"
	"reflect"
	"testing"
)

const loopbackIPv4 = "127.0.0.1"
const loopbackIPv6 = "::1"

var fake faker.Faker

func init() {
	fake = faker.New()
}

func Test_parseWhitelist(t *testing.T) {
	type args struct {
		whitelist string
	}
	tests := []struct {
		name string
		args args
		want []net.IP
	}{
		{"whitelist vide", args{""}, nil},
		{"whitelist sans espace", args{"127.0.0.1,2345:0425:2CA1:0000:0000:0567:5673:23b5"}, []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("2345:425:2ca1::567:5673:23b5")}},
		{"whitelist avec espaces", args{" 127.0.0.1 ,\t2345:0425:2CA1:0000:0567:0000:5673:23b5,::1"}, []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("2345:425:2ca1:0:567:0:5673:23b5"), net.ParseIP("::1")}},
		{"whitelist avec ipv6 malformées", args{" 2345:0425:2CA1:0000:0567:5673:23b5 , ::1"}, []net.IP{net.ParseIP("::1")}},
		{"whitelist avec ipv4 malformées", args{" 256.255.255.255,127.0.0.1 "}, []net.IP{net.ParseIP("127.0.0.1")}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseWhitelist(tt.args.whitelist); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseWhitelist() -> got %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isIPWhitelisted(t *testing.T) {
	type args struct {
		whitelist []net.IP
		ip        string
	}
	knownIP := generateRandomIP()
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"cas pas de whitelist (nil) : refusé", args{nil, generateRandomIP()}, false},
		{"cas pas de whitelist (vide) : refusé", args{[]net.IP{}, generateRandomIP()}, false},
		{"cas pas de whitelist et vient du loopback en ipv4 : accepté", args{[]net.IP{}, loopbackIPv4}, true},
		{"cas pas de whitelist et vient du loopback en ipv6 : accepté", args{nil, loopbackIPv6}, true},
		{"cas whitelist et vient du loopback en ipv4 : refusé", args{[]net.IP{net.ParseIP(generateRandomIP())}, loopbackIPv4}, false},
		{"cas whitelist et vient du loopback en ipv6 : refusé", args{[]net.IP{net.ParseIP(generateRandomIP())}, loopbackIPv6}, false},
		{"cas whitelist et vient d'une adresse non whitelistée : refusé", args{toIP(generateRandomIPs()), generateRandomIP()}, false},
		{"cas whitelist et vient d'une adresse whitelistée : accepté", args{toIP(append(generateRandomIPs(), knownIP)), knownIP}, true},
		{"whitelist ipv6 réécrite : accepté", args{[]net.IP{net.ParseIP("2345:425:2ca1::567:5673:23b5")}, "2345:425:2ca1::567:5673:23b5"}, true},
		{"whitelist ipv6 non réécrite : accepté", args{[]net.IP{net.ParseIP("2345:425:2ca1::567:5673:23b5")}, "2345:0425:2CA1:0000:0000:0567:5673:23b5"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isIPWhitelisted(tt.args.whitelist, tt.args.ip); got != tt.want {
				t.Errorf("isIPWhitelisted() = %v, want %v", got, tt.want)
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
