package common

import "testing"
import "log"
import "net"
import "github.com/stretchr/testify/assert"
import "github.com/google/gopacket/routing"
import "github.com/google/gopacket/examples/util"

func TestSynScan(t *testing.T) {
	defer util.Run()()
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
	// Note:  newScanner creates and closes a pcap Handle once for
	// every scan target.  We could do much better, were this not an
	// example ;)
	scantype := SynScan
	portRange := PortRange{7990, 8009}
	s, err := newScanner(ip, router, scantype, portRange)
	if err != nil {
		t.Fatalf("unable to create scanner for %v: %v", ip, err)
	}
	if err := s.scan(); err != nil {
		log.Printf("unable to scan %v: %v", ip, err)
	}
	s.close()

	response := s.response()
	assert.Equal(t, 20, len(response))

	for port, status := range response {
		if port != 8000 {
			assert.Equal(t, status, PortClosed)
		} else {
			assert.Equal(t, status, PortOpen)
		}
	}
}

func TestFinScan(t *testing.T) {
	defer util.Run()()
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
	// Note:  newScanner creates and closes a pcap Handle once for
	// every scan target.  We could do much better, were this not an
	// example ;)
	scantype := FinScan
	portRange := PortRange{7990, 8009}
	s, err := newScanner(ip, router, scantype, portRange)
	if err != nil {
		t.Fatalf("unable to create scanner for %v: %v", ip, err)
	}
	if err := s.scan(); err != nil {
		log.Printf("unable to scan %v: %v", ip, err)
	}
	s.close()

	response := s.response()
	assert.Equal(t, 20, len(response))

	for port, status := range response {
		if port != 8000 {
			assert.Equal(t, status, PortClosed)
		} else {
			assert.Equal(t, status, PortOpen|PortFiltered)
		}
	}
}
