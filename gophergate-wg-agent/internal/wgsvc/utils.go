package wgsvc

import (
	"net"
	"strings"
)

func pickPrimaryIP(cidrs []string) net.IP {
	for _, c := range cidrs {
		ip, ipNet, err := net.ParseCIDR(strings.TrimSpace(c))
		if err != nil {
			continue
		}
		ones, bits := ipNet.Mask.Size()
		if (ip.To4() != nil && ones == 32 && bits == 32) || (ip.To4() == nil && ones == 128 && bits == 128) {
			return ip
		}
	}
	return nil
}

func optionalString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func optionalI16(v int) *int16 {
	if v <= 0 {
		return nil
	}
	x := int16(v)
	return &x
}
