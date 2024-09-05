package islazy

import (
	"encoding/binary"
	"net"
)

// IpsInCIDR returns a list of usable IP addresses in a given CIDR block
// excluding network and broadcast addresses for CIDRs larger than /31.
func IpsInCIDR(cidr string) ([]string, error) {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	mask := binary.BigEndian.Uint32(ipnet.Mask)
	start := binary.BigEndian.Uint32(ipnet.IP)
	end := (start & mask) | (mask ^ 0xFFFFFFFF)

	var ips []string
	ip := make(net.IP, 4) // Preallocate buffer

	// Iterate over the range of IPs
	for i := start; i <= end; i++ {
		// Exclude network and broadcast addresses in larger CIDR ranges
		if !(i&0xFF == 255 || i&0xFF == 0) || ipnet.Mask[3] >= 30 {
			binary.BigEndian.PutUint32(ip, i)
			ips = append(ips, ip.String())
		}
	}

	return ips, nil
}
