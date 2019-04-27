package common

import (
	"log"
	"net"
)

func SubnetSplit(ipBlock string, count int) []IpRange {
	var results []IpRange
	ip, ipnet, err := net.ParseCIDR(ipBlock)
	if err != nil {
		log.Fatal(err)
	}
	a, b := ipnet.Mask.Size()
	blockSize := (1 << uint(b-a)) / count
	i := 0
	var start net.IP
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		if i == 0 {
			start = make(net.IP, len(ip))
			copy(start, ip)
		}
		if i == blockSize-1 {
			i = 0
			results = append(results, IpRange{start, ip})
			continue
		}
		i++
	}
	return results
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
