// Copyright 2012 Google, Inc. All rights reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file in the root of the source
// tree.

// synscan implements a TCP syn scanner on top of pcap.
// It's more complicated than arpscan, since it has to handle sending packets
// outside the local network, requiring some routing and ARP work.
//
// Since this is just an example program, it aims for simplicity over
// performance.  It doesn't handle sending packets very quickly, it scans IPs
// serially instead of in parallel, and uses gopacket.Packet instead of
// gopacket.DecodingLayerParser for packet processing.  We also make use of very
// simple timeout logic with time.Since.
//
// Making it blazingly fast is left as an exercise to the reader.
package common

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/routing"
)

// scanner handles scanning a single IP address.
type scanner struct {
	// iface is the interface to send packets on.
	iface *net.Interface
	// destination, gateway (if applicable), and source IP addresses to use.
	dst, gw, src net.IP

	handle *pcap.Handle

	// opts and buf allow us to easily serialize packets in the send()
	// method.
	opts gopacket.SerializeOptions
	buf  gopacket.SerializeBuffer
}

// newScanner creates a new scanner for a given destination IP address, using
// router to determine how to route packets to that IP.
func NewScanner(ip net.IP, router routing.Router) (*scanner, error) {
	s := &scanner{
		dst: ip.To4(),
		opts: gopacket.SerializeOptions{
			FixLengths:       true,
			ComputeChecksums: true,
		},
		buf: gopacket.NewSerializeBuffer(),
	}
	// Figure out the route to the IP.
	iface, gw, src, err := router.Route(ip)
	if err != nil {
		return nil, err
	}
	//log.Printf("scanning, ip %v with interface %v, gateway %v, src %v", ip, iface.Name, gw, src)
	s.gw, s.src, s.iface = gw, src, iface

	// Open the handle for reading/writing.
	// Note we could very easily add some BPF filtering here to greatly
	// decrease the number of packets we have to look at when getting back
	// scan results.
	s.handle, err = pcap.OpenLive(iface.Name, 65536, true, pcap.BlockForever)
	if err != nil {
		return nil, err
	}
	/*
		  TODO - investigate why BPF filter causes timeout
		  err = s.handle.SetBPFFilter("arp or (tcp and port 54321)")
			if err != nil {
				log.Fatal(err)
			}*/
	return s, nil
}

// close cleans up the handle.
func (s *scanner) Close() {
	s.handle.Close()
}

// getHwAddr is a hacky but effective way to get the destination hardware
// address for our packets.  It does an ARP request for our gateway (if there is
// one) or destination IP (if no gateway is necessary), then waits for an ARP
// reply.  This is pretty slow right now, since it blocks on the ARP
// request/reply.
func (s *scanner) getHwAddr() (net.HardwareAddr, error) {
	start := time.Now()
	arpDst := s.dst
	if s.gw != nil {
		arpDst = s.gw
	}
	// Prepare the layers to send for an ARP request.
	eth := layers.Ethernet{
		SrcMAC:       s.iface.HardwareAddr,
		DstMAC:       net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		EthernetType: layers.EthernetTypeARP,
	}
	arp := layers.ARP{
		AddrType:          layers.LinkTypeEthernet,
		Protocol:          layers.EthernetTypeIPv4,
		HwAddressSize:     6,
		ProtAddressSize:   4,
		Operation:         layers.ARPRequest,
		SourceHwAddress:   []byte(s.iface.HardwareAddr),
		SourceProtAddress: []byte(s.src),
		DstHwAddress:      []byte{0, 0, 0, 0, 0, 0},
		DstProtAddress:    []byte(arpDst),
	}
	// Send a single ARP request packet (we never retry a send, since this
	// is just an example ;)
	if err := s.send(&eth, &arp); err != nil {
		return nil, err
	}

	// Wait 3 seconds for an ARP reply.
	for {
		if time.Since(start) > time.Second*3 {
			return nil, errors.New("timeout getting ARP reply")
		}
		data, _, err := s.handle.ReadPacketData()
		if err == pcap.NextErrorTimeoutExpired {
			continue
		} else if err != nil {
			return nil, err
		}
		packet := gopacket.NewPacket(data, layers.LayerTypeEthernet, gopacket.NoCopy)
		if arpLayer := packet.Layer(layers.LayerTypeARP); arpLayer != nil {
			arp := arpLayer.(*layers.ARP)
			if net.IP(arp.SourceProtAddress).Equal(net.IP(arpDst)) {
				return net.HardwareAddr(arp.SourceHwAddress), nil
			}
		}
	}
}

// scan scans the dst IP address of this scanner.
func (s *scanner) Scan(scanType PortScanType, portRange PortRange) (PortScanResult, error) {
	// First off, get the MAC address we should be sending packets to.
	log.Print("get hw addr")
	hwaddr, err := s.getHwAddr()
	if err != nil {
		return nil, err
	}
	result := make(PortScanResult)
	for i := portRange.Start; i <= portRange.End; i++ {
		// keeping the default status as filtered
		result[i] = PortResult{PortFiltered, ""}
	}
	// Construct all the network layers we need.
	eth := layers.Ethernet{
		SrcMAC:       s.iface.HardwareAddr,
		DstMAC:       hwaddr,
		EthernetType: layers.EthernetTypeIPv4,
	}
	ip4 := layers.IPv4{
		SrcIP:    s.src,
		DstIP:    s.dst,
		Version:  4,
		TTL:      64,
		Protocol: layers.IPProtocolTCP,
	}
	tcp := layers.TCP{
		SrcPort: 54321,
		DstPort: layers.TCPPort(portRange.Start), // will be incremented during the scan
	}

	if scanType == SynScan {
		tcp.SYN = true
	} else if scanType == FinScan {
		tcp.FIN = true
	}
	tcp.SetNetworkLayerForChecksum(&ip4)

	// Create the flow we expect returning packets to have, so we can check
	// against it and discard useless packets.
	ipFlow := gopacket.NewFlow(layers.EndpointIPv4, s.dst, s.src)
	// Send one packet per loop iteration until we've sent packets
	// to all of ports [1, 65535].
	for ; tcp.DstPort <= layers.TCPPort(portRange.End); tcp.DstPort++ {
		log.Printf("sending %d", tcp.DstPort)
		if err := s.send(&eth, &ip4, &tcp); err != nil {
			log.Printf("error sending to port %v: %v", tcp.DstPort, err)
		}
	}

	packetSource := gopacket.NewPacketSource(s.handle, s.handle.LinkType())
	packets := packetSource.Packets()
	ticker := time.Tick(time.Second * 4)
	for {
		select {
		case packet := <-packets:
			// use err and reply
			// Find the packets we care about, and print out logging
			// information about them.  All others are ignored.
			if net := packet.NetworkLayer(); net == nil {
				// log.Printf("packet has no network layer")
			} else if net.NetworkFlow() != ipFlow {
				// log.Printf("packet does not match our ip src/dst")
			} else if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer == nil {
				// log.Printf("packet has not tcp layer")
			} else if tcp, ok := tcpLayer.(*layers.TCP); !ok {
				// We panic here because this is guaranteed to never
				// happen.
				panic("tcp layer is not tcp layer :-/")
			} else if tcp.DstPort != 54321 {
				// log.Printf("dst port %v does not match", tcp.DstPort)
			} else if tcp.RST {
				log.Printf("  port %v closed", tcp.SrcPort)
				result[uint16(tcp.SrcPort)] = PortResult{PortClosed, ""}
			} else if scanType == SynScan && tcp.SYN && tcp.ACK {
				log.Printf("  port %v open", tcp.SrcPort)
				result[uint16(tcp.SrcPort)] = PortResult{PortOpen, ""}
			} else {
				// log.Printf("ignoring useless packet")
			}

		case <-ticker:
			// call timed out
			log.Println("timeout")
			if scanType == FinScan {
				for port, _ := range result {
					if result[port].Status == PortFiltered {
						result[port] = PortResult{PortOpen | PortFiltered, ""}
					}
				}
			}
			log.Printf("end scan")
			return result, nil
		}
	}
	return result, nil
}

// send sends the given layers as a single packet on the network.
func (s *scanner) send(l ...gopacket.SerializableLayer) error {
	if err := gopacket.SerializeLayers(s.buf, s.opts, l...); err != nil {
		return err
	}
	return s.handle.WritePacketData(s.buf.Bytes())
}

func verifyPeerCert(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
	return nil
}

func HandleWebPort(port uint16, addr string, result PortScanResult) {

	var prefix string
	switch port {
	case 80:
		prefix = "http://"
		break
	case 443:
		prefix = "https://"
	}

	var resp *http.Response
	var err error

	switch port {
	case 80:
		resp, err = http.Head(prefix + addr + "/")
		if err != nil {
			log.Println(err)
			result[port] = PortResult{PortOpen, ""}
			return
		}
		break
	case 443:
		transport := &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify:    true,
				VerifyPeerCertificate: verifyPeerCert,
			},
		}
		client := &http.Client{Transport: transport}
		resp, err = client.Head(prefix + addr + "/")
		if err != nil {
			log.Println(err)
			result[port] = PortResult{PortOpen, ""}
			return
		}
	}

	header := resp.Header
	for key := range header {
		log.Printf("%s: %s\n", key, header.Get(key))
	}
	result[port] = PortResult{PortOpen, header.Get("server")}

}

func NormalPortScan(ip net.IP, portRange PortRange, timeout time.Duration) PortScanResult {
	wg := sync.WaitGroup{}
	var mu sync.Mutex
	defer wg.Wait()
	result := make(PortScanResult)

	for port := portRange.Start; port <= portRange.End; port++ {
		wg.Add(1)
		go func(port uint16) {
			defer wg.Done()
			var status PortStatus
			addr := fmt.Sprintf("%s:%d", ip.String(), port)
			conn, err := net.DialTimeout("tcp", addr, timeout)
			if err != nil {
				status = PortClosed
			} else {
				status = PortOpen
			}
			mu.Lock()
			if status == PortOpen {
				if (port == 80) || (port == 443) {
					HandleWebPort(port, addr, result)
				} else {
					conn.SetReadDeadline(time.Now().Add(5 * time.Second))
					byteArray := make([]byte, 256)
					if _, e := conn.Read(byteArray); e == nil {
						fmt.Printf("%s\n", string(byteArray))
						byteArray = bytes.Trim(byteArray, "\x00")
						result[port] = PortResult{status, string(byteArray)}
					}
				}
			} else {
				result[port] = PortResult{status, ""}
			}
			mu.Unlock()
		}(port)
	}

	return result
}
