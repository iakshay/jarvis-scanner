package common

import (
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

func TestIsAlive(t *testing.T) {
	result := IsAlive(IpRange{net.ParseIP("10.0.0.5"), net.ParseIP("10.0.0.10")}).Result
	assert.Equal(t, len(result), 6)
	vmIp1 := net.ParseIP("10.0.0.10")
	vmIp2 := net.ParseIP("10.0.0.11")
	for _, ipResult := range result {
		if vmIp1.Equal(ipResult.Ip) || vmIp2.Equal(ipResult.Ip) {
			assert.Equal(t, ipResult.Status, IpAlive)
		} else {
			assert.Equal(t, ipResult.Status, IpDead)
		}
	}

	result = IsAlive(IpRange{net.ParseIP("10.0.0.10"), net.ParseIP("10.0.0.11")}).Result
	assert.Equal(t, len(result), 2)
	for _, ipResult := range result {
		assert.Equal(t, ipResult.Status, IpAlive)
	}

	result = IsAlive(IpRange{net.ParseIP("127.0.0.1"), net.ParseIP("127.0.0.20")}).Result
	assert.Equal(t, len(result), 20)
	for _, ipResult := range result {
		assert.Equal(t, ipResult.Status, IpAlive)
	}

	result = IsAlive(IpRange{net.ParseIP("10.0.0.1"), net.ParseIP("10.0.0.255")}).Result
	assert.Equal(t, len(result), 255)
	for _, ipResult := range result {
		if vmIp1.Equal(ipResult.Ip) || vmIp2.Equal(ipResult.Ip) {
			assert.Equal(t, ipResult.Status, IpAlive)
		} else {
			assert.Equal(t, ipResult.Status, IpDead)
		}
	}
}

func TestIpRange(t *testing.T) {
	ipRange := IpRange{net.ParseIP("10.0.0.0"), net.ParseIP("10.0.0.255")}
	assert.Equal(t, len(ipRange.Iterate()), 256)
	ipRange = IpRange{net.ParseIP("10.0.0.0"), net.ParseIP("10.0.255.255")}
	assert.Equal(t, len(ipRange.Iterate()), 256*256)
	ipRange = IpRange{net.ParseIP("10.0.0.64"), net.ParseIP("10.0.255.255")}
	assert.Equal(t, len(ipRange.Iterate()), 256*(256-64))
}
