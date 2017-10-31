package utils

import (
	"net"
	"strconv"
	"strings"
)

// Hosts returns the IP's from a provided CIDR
func Hosts(cidr string) ([]string, error) {

	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}

	if len(ips) > 1 {

		// remove network address and broadcast address
		return ips[1 : len(ips)-1], nil
	}

	// suppose this will only really happen with /32's
	return ips, nil
}

// helper method: https://play.golang.org/p/m8TNTtygK0
func inc(ip net.IP) {

	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// Ports returns a slice of ports parsed from a string
func Ports(ports string) ([]int, error) {

	parsed := strings.Split(ports, ",")

	var r []int

	for _, port := range parsed {

		p, err := strconv.Atoi(port)
		if err != nil {
			continue
		}

		r = append(r, p)
	}

	return r, nil
}
