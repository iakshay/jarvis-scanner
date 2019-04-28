package common

import (
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

func TestSubnetSplit(t *testing.T) {
	results, _ := SubnetSplit("192.168.2.1/24", 4)
	assert.Equal(t, len(results), 4)
	assert.Equal(t, results[0].Start.To4(), net.ParseIP("192.168.2.0").To4())
	assert.Equal(t, results[1].Start.To4(), net.ParseIP("192.168.2.64").To4())
	assert.Equal(t, results[2].Start.To4(), net.ParseIP("192.168.2.128").To4())
	assert.Equal(t, results[3].Start.To4(), net.ParseIP("192.168.2.192").To4())
}

func TestIpSplit(t *testing.T) {
	results, _ := SubnetSplit("192.168.2.1", 4)
	assert.Equal(t, len(results), 1)
	assert.Equal(t, results[0].Start.To4(), net.ParseIP("192.168.2.1").To4())
	assert.Equal(t, results[0].End.To4(), net.ParseIP("192.168.2.1").To4())
}

func TestPortRangeSplit(t *testing.T) {
	results := PortRangeSplit(PortRange{0, 10}, 4)
	assert.Equal(t, len(results), 4)
	assert.Equal(t, results[0].Start, uint16(0))
	assert.Equal(t, results[0].End, uint16(2))
	assert.Equal(t, results[1].Start, uint16(3))
	assert.Equal(t, results[1].End, uint16(5))
	assert.Equal(t, results[2].Start, uint16(6))
	assert.Equal(t, results[2].End, uint16(8))
	assert.Equal(t, results[3].Start, uint16(9))
	assert.Equal(t, results[3].End, uint16(10))

	results = PortRangeSplit(PortRange{0, 0}, 4)
	assert.Equal(t, len(results), 1)

	results = PortRangeSplit(PortRange{0, 65535}, 19)
	assert.Equal(t, len(results), 19)
}
