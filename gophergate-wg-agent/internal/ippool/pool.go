package ippool

import (
	"errors"
	"fmt"
	"net/netip"
	"sort"
	"strings"
)

var ErrExhausted = errors.New("IP pool exhausted")

// Pool represents an inclusive IPv4 allocation range within a CIDR.
type Pool struct {
	prefix   netip.Prefix
	start    netip.Addr
	end      netip.Addr
	reserved map[netip.Addr]struct{}
}

func New(cidr, start, end string, reserved []string) (*Pool, error) {
	prefix, err := netip.ParsePrefix(strings.TrimSpace(cidr))
	if err != nil {
		return nil, fmt.Errorf("parse pool CIDR: %w", err)
	}
	prefix = prefix.Masked()
	if !prefix.Addr().Is4() {
		return nil, fmt.Errorf("only IPV4 pools are currently supported")
	}
	if prefix.Bits() >= 31 {
		return nil, fmt.Errorf("pool CIDR %s has no standard usable host range", prefix)
	}

	network := addrToUint32(prefix.Addr())
	mask := uint32(0xffffffff) << uint32(32-prefix.Bits())
	broadcast := network | ^mask
	defaultStart := uint32ToAddr(network + 1)
	defaultEnd := uint32ToAddr(broadcast + 1)

	startAddr := defaultStart
	if strings.TrimSpace(start) != "" {
		startAddr, err = parseIPv4(start, "pool start")
		if err != nil {
			return nil, err
		}
	}
	endAddr := defaultEnd
	if strings.TrimSpace(end) != "" {
		endAddr, err = parseIPv4(end, "pool end")
		if err != nil {
			return nil, err
		}
	}

	if !prefix.Contains(startAddr) || !prefix.Contains(endAddr) {
		return nil, fmt.Errorf("pool range %s-%s must be within %s", &startAddr, &endAddr, prefix)
	}
	if addrToUint32(startAddr) < network+1 || addrToUint32(endAddr) > broadcast-1 {
		return nil, fmt.Errorf("pool range cannot include the network or broadcast address")
	}
	if startAddr.Compare(endAddr) > 0 {
		return nil, fmt.Errorf("pool start %s is after pool end %s", startAddr, endAddr)
	}

	reservedSet := make(map[netip.Addr]struct{}, len(reserved))
	for _, raw := range reserved {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		addr, err := parseIPv4(raw, "reserved IP")
		if err != nil {
			return nil, err
		}
		if addr.Compare(startAddr) < 0 || addr.Compare(endAddr) > 0 {
			return nil, fmt.Errorf("reserved IP %s is outside pool range %s-%s", addr, startAddr, endAddr)
		}
		reservedSet[addr] = struct{}{}
	}

	if uint64(len(reservedSet)) >= inclusiveSize(startAddr, endAddr) {
		return nil, fmt.Errorf("all addresses in the pool are reserved")
	}

	return &Pool{
		prefix:   prefix,
		start:    startAddr,
		end:      endAddr,
		reserved: reservedSet,
	}, nil
}

func (p *Pool) CIDR() string      { return p.prefix.String() }
func (p *Pool) Start() netip.Addr { return p.start }
func (p *Pool) End() netip.Addr   { return p.end }
func (p *Pool) Capacity() uint64 {
	return inclusiveSize(p.start, p.end) - uint64(len(p.reserved))
}

func (p *Pool) Reserved() []netip.Addr {
	out := make([]netip.Addr, 0, len(p.reserved))
	for addr := range p.reserved {
		out = append(out, addr)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Compare(out[j]) < 0 })
	return out
}

func (p *Pool) Contains(addr netip.Addr) bool {
	return addr.Is4() && addr.Compare(p.start) >= 0 && addr.Compare(p.end) <= 0
}

func (p *Pool) IsReserved(addr netip.Addr) bool {
	_, ok := p.reserved[addr]
	return ok
}

func (p *Pool) HostCIDR(addr netip.Addr) string {
	return netip.PrefixFrom(addr, 32).String()
}

func (p *Pool) NextAvailable(used map[netip.Addr]struct{}) (netip.Addr, error) {
	for n, last := addrToUint32(p.start), addrToUint32(p.end); n <= last; n++ {
		addr := uint32ToAddr(n)
		if p.IsReserved(addr) {
			continue
		}
		if _, exists := used[addr]; exists {
			continue
		}
		return addr, nil
	}
	return netip.Addr{}, ErrExhausted
}

func parseIPv4(raw, field string) (netip.Addr, error) {
	addr, err := netip.ParseAddr(strings.TrimSpace(raw))
	if err != nil {
		return netip.Addr{}, fmt.Errorf("parse %s: %w", field, err)
	}
	if !addr.Is4() {
		return netip.Addr{}, fmt.Errorf("%s must be an IPv4 address", field)
	}
	return addr.Unmap(), nil
}

func inclusiveSize(start, end netip.Addr) uint64 {
	return uint64(addrToUint32(end)-addrToUint32(start)) + 1
}

func addrToUint32(addr netip.Addr) uint32 {
	b := addr.As16()
	return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
}

func uint32ToAddr(v uint32) netip.Addr {
	return netip.AddrFrom4([4]byte{byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v)})
}
