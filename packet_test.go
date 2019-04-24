package common

import (
	"net"
	"testing"
)

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Fatalf("%s != %s", a, b)
	}
}

func TestIsAlive(t *testing.T) {
	assertEqual(t, true, Ping(net.ParseIP("10.0.0.11")))
	assertEqual(t, false, Ping(net.ParseIP("10.0.0.12")))
}

func TestIpRange(t *testing.T) {
	ipRange := IpRange{net.ParseIP("10.0.0.0"), net.ParseIP("10.0.0.255")}
	assertEqual(t, len(ipRange.Iterate()), 256)
	ipRange = IpRange{net.ParseIP("10.0.0.0"), net.ParseIP("10.0.255.255")}
	assertEqual(t, len(ipRange.Iterate()), 256*256)
	ipRange = IpRange{net.ParseIP("10.0.0.64"), net.ParseIP("10.0.255.255")}
	assertEqual(t, len(ipRange.Iterate()), 256*(256-64))
}
