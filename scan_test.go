package common

import "testing"
import "log"
import "net"
import "time"
import "github.com/stretchr/testify/assert"
import "github.com/google/gopacket/routing"

//import "github.com/google/gopacket/examples/util"

func TestSynScan(t *testing.T) {
	//defer util.Run()()
	router, err := routing.New()
	if err != nil {
		log.Fatal("routing error:", err)
	}
	// assuming we are running simple http server at port 8000
	ip := net.ParseIP("10.0.0.11")
	ip = ip.To4()
	if ip == nil {
		return
	}
	// Note:  NewScanner creates and closes a pcap Handle once for
	// every scan target.  We could do much better, were this not an
	// example ;)
	s, err := NewScanner(ip, router)
	if err != nil {
		t.Fatalf("unable to create scanner for %v: %v", ip, err)
	}
	defer s.Close()
	response, err := s.Scan(SynScan, PortRange{7990, 8009})
	assert.Equal(t, err, nil)

	assert.Equal(t, 20, len(response))

	for port, result := range response {
		if port != 8000 {
			assert.Equal(t, result.Status, PortClosed)
		} else {
			assert.Equal(t, result.Status, PortOpen)
		}
	}
}

func TestFinScan(t *testing.T) {
	//defer util.Run()()
	router, err := routing.New()
	if err != nil {
		log.Fatal("routing error:", err)
	}
	// assuming we are running simple http server at port 8000
	ip := net.ParseIP("10.0.0.11")
	ip = ip.To4()
	if ip == nil {
		return
	}
	// Note:  NewScanner creates and closes a pcap Handle once for
	// every scan target.  We could do much better, were this not an
	// example ;)
	s, err := NewScanner(ip, router)
	if err != nil {
		t.Fatalf("unable to create scanner for %v: %v", ip, err)
	}
	response, err := s.Scan(FinScan, PortRange{7990, 8009})
	assert.Equal(t, err, nil)
	assert.Equal(t, 20, len(response))

	for port, result := range response {
		if port != 8000 {
			assert.Equal(t, result.Status, PortClosed)
		} else {
			assert.Equal(t, result.Status, PortOpen|PortFiltered)
		}
	}
}

func TestMultipleScan(t *testing.T) {
	//defer util.Run()()
	router, err := routing.New()
	if err != nil {
		log.Fatal("routing error:", err)
	}
	// assuming we are running simple http server at port 8000
	ip := net.ParseIP("10.0.0.11")
	ip = ip.To4()
	if ip == nil {
		return
	}
	// Note:  NewScanner creates and closes a pcap Handle once for
	// every scan target.  We could do much better, were this not an
	// example ;)
	s, err := NewScanner(ip, router)
	if err != nil {
		t.Fatalf("unable to create scanner for %v: %v", ip, err)
	}
	response, err := s.Scan(SynScan, PortRange{7990, 8009})
	assert.Equal(t, err, nil)

	assert.Equal(t, 20, len(response))

	for port, result := range response {
		if port != 8000 {
			assert.Equal(t, result.Status, PortClosed)
		} else {
			assert.Equal(t, result.Status, PortOpen)
		}
	}
	s.Close()

	s, err = NewScanner(ip, router)
	response, err = s.Scan(FinScan, PortRange{7990, 8009})
	assert.Equal(t, err, nil)
	s.Close()

	for port, result := range response {
		if port != 8000 {
			assert.Equal(t, result.Status, PortClosed)
		} else {
			assert.Equal(t, result.Status, PortOpen|PortFiltered)
		}
	}

	s, err = NewScanner(ip, router)
	response, err = s.Scan(SynScan, PortRange{7000, 7009})
	assert.Equal(t, err, nil)
	s.Close()

	assert.Equal(t, 10, len(response))

	for port, result := range response {
		if port != 7000 {
			assert.Equal(t, result.Status, PortClosed)
		} else {
			assert.Equal(t, result.Status, PortOpen)
		}
	}
}

func TestNormalScan(t *testing.T) {
	result := NormalPortScan("10.0.0.11", PortRange{7000, 7009}, 3*time.Second)
	assert.Equal(t, len(result), 10)
	for port, result := range result {
		if port != 7000 {
			assert.Equal(t, result.Status, PortClosed)
		} else {
			assert.Equal(t, result.Status, PortOpen)
		}
	}
}
