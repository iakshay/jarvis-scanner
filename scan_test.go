package common

import "testing"
import "log"
import "net"
import "github.com/google/gopacket/routing"
import "github.com/google/gopacket/examples/util"

func TestScan(t *testing.T) {
	defer util.Run()()
	router, err := routing.New()
	if err != nil {
		log.Fatal("routing error:", err)
	}
	ip := net.ParseIP("10.0.0.11")
	ip = ip.To4()
	if ip == nil {
		return
	}
	// Note:  newScanner creates and closes a pcap Handle once for
	// every scan target.  We could do much better, were this not an
	// example ;)
	scantype := FinScan
	s, err := newScanner(ip, router, scantype)
	if err != nil {
		t.Fatalf("unable to create scanner for %v: %v", ip, err)
	}
	if err := s.scan(); err != nil {
		log.Printf("unable to scan %v: %v", ip, err)
	}
	s.close()
}
