package common

import (
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

func TestSubnetSplit(t *testing.T) {
	results := SubnetSplit("192.168.2.1/24", 4)
	assert.Equal(t, len(results), 4)
	assert.Equal(t, results[0].Start, net.ParseIP("192.168.2.0").To4())
	assert.Equal(t, results[1].Start, net.ParseIP("192.168.2.64").To4())
	assert.Equal(t, results[2].Start, net.ParseIP("192.168.2.128").To4())
	assert.Equal(t, results[3].Start, net.ParseIP("192.168.2.192").To4())
}
