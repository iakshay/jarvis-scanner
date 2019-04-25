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
	// Note:  newScanner creates and closes a pcap Handle once for
	// every scan target.  We could do much better, were this not an
	// example ;)
	s, err := newScanner(ip, router)
	if err != nil {
		t.Fatalf("unable to create scanner for %v: %v", ip, err)
	}
	defer s.close()
	response, err := s.scan(SynScan, PortRange{7990, 8009})
	assert.Equal(t, err, nil)

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
	// Note:  newScanner creates and closes a pcap Handle once for
	// every scan target.  We could do much better, were this not an
	// example ;)
	s, err := newScanner(ip, router)
	if err != nil {
		t.Fatalf("unable to create scanner for %v: %v", ip, err)
	}
	response, err := s.scan(FinScan, PortRange{7990, 8009})
	assert.Equal(t, err, nil)
	assert.Equal(t, 20, len(response))

	for port, status := range response {
		if port != 8000 {
			assert.Equal(t, status, PortClosed)
		} else {
			assert.Equal(t, status, PortOpen|PortFiltered)
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
	// Note:  newScanner creates and closes a pcap Handle once for
	// every scan target.  We could do much better, were this not an
	// example ;)
	s, err := newScanner(ip, router)
	if err != nil {
		t.Fatalf("unable to create scanner for %v: %v", ip, err)
	}
	response, err := s.scan(SynScan, PortRange{7990, 8009})
	assert.Equal(t, err, nil)

	assert.Equal(t, 20, len(response))

	for port, status := range response {
		if port != 8000 {
			assert.Equal(t, status, PortClosed)
		} else {
			assert.Equal(t, status, PortOpen)
		}
	}
	s.close()

	s, err = newScanner(ip, router)
	response, err = s.scan(FinScan, PortRange{7990, 8009})
	assert.Equal(t, err, nil)
	s.close()

	for port, status := range response {
		if port != 8000 {
			assert.Equal(t, status, PortClosed)
		} else {
			assert.Equal(t, status, PortOpen|PortFiltered)
		}
	}

	s, err = newScanner(ip, router)
	response, err = s.scan(SynScan, PortRange{7000, 7009})
	assert.Equal(t, err, nil)
	s.close()

	assert.Equal(t, 10, len(response))

	for port, status := range response {
		if port != 7000 {
			assert.Equal(t, status, PortClosed)
		} else {
			assert.Equal(t, status, PortOpen)
		}
	}
}

func TestNormalScan(t *testing.T) {
	NormalPortScan("10.0.0.11", PortRange{7000, 7009}, 3*time.Second)
}
