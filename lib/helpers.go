package lib

import (
	"encoding/binary"
	"net"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// ScreenshotPath determines a full path and file name for a screenshot image
func ScreenshotPath(destination string, url *url.URL, path string) string {

	var fname, dst string
	if destination == "" {
		fname = SafeFileName(url.String())
		dst = filepath.Join(path, fname)
	} else {
		fname = destination
		if filepath.IsAbs(fname) {
			dst = fname
		} else {
			dst = filepath.Join(path, fname)
		}
	}

	return dst
}

// SafeFileName return a safe string that can be used in file names
func SafeFileName(str string) string {

	name := strings.ToLower(str)
	name = strings.Trim(name, " ")

	separators, err := regexp.Compile(`[ &_=+:/]`)
	if err == nil {
		name = separators.ReplaceAllString(name, "-")
	}

	legal, err := regexp.Compile(`[^[:alnum:]-.]`)
	if err == nil {
		name = legal.ReplaceAllString(name, "")
	}

	for strings.Contains(name, "--") {
		name = strings.Replace(name, "--", "-", -1)
	}

	return name + `.png`
}

// PortsFromString returns a slice of ports parsed from a string
func PortsFromString(ports string) ([]int, error) {

	parsed := strings.Split(ports, ",")

	var m = make(map[int]bool)
	var r []int

	for _, port := range parsed {

		p, err := strconv.Atoi(port)
		if err != nil {
			continue
		}

		// uniq
		if m[p] {
			continue
		}

		r = append(r, p)
		m[p] = true
	}

	return r, nil
}

// HostsInCIDR returns the IP's from a provided CIDR
func HostsInCIDR(cidr string) (ips []string, err error) {

	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	mask := binary.BigEndian.Uint32(ipnet.Mask)
	start := binary.BigEndian.Uint32(ipnet.IP)
	end := (start & mask) | (mask ^ 0xFFFFFFFF)

	for i := start; i <= end; i++ {
		if !(i&0xFF == 255 || i&0xFF == 0) {
			ip := make(net.IP, 4)
			binary.BigEndian.PutUint32(ip, i)
			ips = append(ips, ip.String())
		}
	}
	return
}

// SliceContainsInt checks if a slice has an int
func SliceContainsInt(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}

	return false
}

// SliceContainsString checks if a slice has a string
func SliceContainsString(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}

	return false
}
