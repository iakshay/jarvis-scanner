package common

import (
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	//"golang.org/x/net/ipv6"
	"log"
	"net"
	"os"
	//"sync"
	"time"
)

func IsAlive(ipRange IpRange) IsAliveResult {
	var result = make(map[string]IpStatus)
	c, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		// most likely, need to run tests as root
		log.Fatal(err)
	}
	defer c.Close()

	// set read / write deadline on the connection
	err = c.SetDeadline(time.Now().Add(3 * time.Second))
	if err != nil {
		log.Fatal(err)
	}

	// send echo packets
	ips := ipRange.Iterate()
	for _, ip := range ips {
		result[ip.String()] = IpDead
		wm := icmp.Message{
			Type: ipv4.ICMPTypeEcho, Code: 0,
			Body: &icmp.Echo{
				ID: os.Getpid() & 0xffff, Seq: 1,
				Data: []byte(ip.String()),
			},
		}
		wb, err := wm.Marshal(nil)
		if err != nil {
			log.Fatal(err)
		}
		if _, err := c.WriteTo(wb, &net.IPAddr{IP: ip}); err != nil {
			log.Fatal(err)
		}
	}
	for {
		rb := make([]byte, 1500)
		n, peer, err := c.ReadFrom(rb)
		if err != nil {
			log.Println(err)
			break
		}
		rm, err := icmp.ParseMessage(1, rb[:n])
		if err != nil {
			log.Println(err)
		}
		switch rm.Type {
		case ipv4.ICMPTypeEchoReply:
			ip := peer.String()
			log.Printf("got reflection from %v", peer)
			if _, ok := result[ip]; ok {
				result[ip] = IpAlive
				log.Println("set alive")
			}
		default:
			log.Printf("got %+v; want echo reply", rm)
		}
	}

	var isAliveResult IsAliveResult
	for k, v := range result {
		log.Println(k, v)
		isAliveResult.Result = append(isAliveResult.Result, IpResult{net.ParseIP(k), v})
	}
	return isAliveResult
}
