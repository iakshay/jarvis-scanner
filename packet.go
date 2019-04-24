package common

import (
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	//"golang.org/x/net/ipv6"
	"log"
	"net"
	"os"
	"sync"
)

func Ping(ip net.IP) bool {
	c, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		// most likely, need to run tests as root
		log.Fatal(err)
	}
	defer c.Close()

	wm := icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{
			ID: os.Getpid() & 0xffff, Seq: 1,
			Data: []byte("HELLO-R-U-THERE"),
		},
	}
	wb, err := wm.Marshal(nil)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := c.WriteTo(wb, &net.IPAddr{IP: ip}); err != nil {
		log.Fatal(err)
	}

	rb := make([]byte, 1500)
	n, peer, err := c.ReadFrom(rb)
	if err != nil {
		log.Fatal(err)
	}
	rm, err := icmp.ParseMessage(1, rb[:n])
	if err != nil {
		log.Fatal(err)
	}
	switch rm.Type {
	case ipv4.ICMPTypeEchoReply:
		log.Printf("got reflection from %v", peer)
		return true
	default:
		log.Printf("got %+v; want echo reply", rm)
	}
	return false
}

func IsAlive(ipRange IpRange) IsAliveResult {
	wg := sync.WaitGroup{}
	defer wg.Wait()
	var result IsAliveResult

	for _, ip := range ipRange.Iterate() {
		wg.Add(1)
		go func(ip net.IP) {
			defer wg.Done()
			ok := Ping(ip)

			if ok {
				log.Println("can ping")
			} else {
				log.Println("no ping")
			}
		}(ip)
	}

	return result
}
