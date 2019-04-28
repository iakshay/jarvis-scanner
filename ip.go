package common

import (
	"log"
	"net"
)

func SubnetSplit(ipBlock string, count int) ([]IpRange, error) {
	var results []IpRange
	ip := net.ParseIP(ipBlock)
	if ip != nil {
		results = append(results, IpRange{ip, ip})
		return results, nil
	}
	ip, ipnet, err := net.ParseCIDR(ipBlock)
	if err != nil {
		return results, err
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
	return results, nil
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func PortRangeSplit(portRange PortRange, count int) []PortRange {
	var results []PortRange
	var rangeLength uint32 = uint32(portRange.End-portRange.Start) + 1
	quotientWork := (rangeLength / uint32(count))
	if quotientWork >= 0 {
		quotientWork--
	}
	remainderWork := rangeLength % uint32(count)
	currStart := uint16(portRange.Start)
	log.Printf("Range: %d Quotient: %d Remainder: %d", rangeLength, quotientWork, remainderWork)
	var currEnd uint16 = 0
	for i := 0; i < count; i++ {
		currEnd = currStart + uint16(quotientWork)
		if remainderWork > 0 {
			currEnd += 1
			remainderWork = remainderWork - 1
		}
		results = append(results, PortRange{currStart, currEnd})
		currStart = currEnd + 1
		if currEnd == portRange.End {
			break
		}
	}
	return results
}
