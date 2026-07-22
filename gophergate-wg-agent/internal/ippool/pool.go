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

// New creates an IPv4 address pool.
//
// The start and end addresses are inclusive. When either value is empty, the
// usable host range of the supplied CIDR is used.
//
// Network and broadcast addresses cannot be included in the allocation range.
func New(cidr, start, end string, reserved []string) (*Pool, error) {
	prefix, err := netip.ParsePrefix(strings.TrimSpace(cidr))
	if err != nil {
		return nil, fmt.Errorf("parse pool CIDR: %w", err)
	}

	prefix = prefix.Masked()

	if !prefix.Addr().Is4() {
		return nil, fmt.Errorf("only IPv4 pools are currently supported")
	}

	if prefix.Bits() >= 31 {
		return nil, fmt.Errorf(
			"pool CIDR %s has no standard usable host range",
			prefix,
		)
	}

	network := addrToUint32(prefix.Addr())
	hostBits := 32 - prefix.Bits()

	// Use uint64 while shifting so a /0 prefix can safely calculate 1 << 32.
	hostMask := uint32((uint64(1) << uint(hostBits)) - 1)
	broadcast := network | hostMask

	defaultStart := uint32ToAddr(network + 1)
	defaultEnd := uint32ToAddr(broadcast - 1)

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

	startValue := addrToUint32(startAddr)
	endValue := addrToUint32(endAddr)

	if !prefix.Contains(startAddr) || !prefix.Contains(endAddr) {
		return nil, fmt.Errorf(
			"pool range %s-%s must be within %s",
			startAddr,
			endAddr,
			prefix,
		)
	}

	if startValue <= network || endValue >= broadcast {
		return nil, fmt.Errorf(
			"pool range cannot include the network or broadcast address",
		)
	}

	if startValue > endValue {
		return nil, fmt.Errorf(
			"pool start %s is after pool end %s",
			startAddr,
			endAddr,
		)
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
			return nil, fmt.Errorf(
				"reserved IP %s is outside pool range %s-%s",
				addr,
				startAddr,
				endAddr,
			)
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

// CIDR returns the parent network in CIDR notation.
func (p *Pool) CIDR() string {
	return p.prefix.String()
}

// Start returns the first allocatable address in the configured range.
func (p *Pool) Start() netip.Addr {
	return p.start
}

// End returns the last allocatable address in the configured range.
func (p *Pool) End() netip.Addr {
	return p.end
}

// Capacity returns the number of addresses that may be allocated after
// excluding reserved addresses.
func (p *Pool) Capacity() uint64 {
	return inclusiveSize(p.start, p.end) - uint64(len(p.reserved))
}

// Reserved returns the configured reserved addresses in ascending order.
func (p *Pool) Reserved() []netip.Addr {
	out := make([]netip.Addr, 0, len(p.reserved))

	for addr := range p.reserved {
		out = append(out, addr)
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].Compare(out[j]) < 0
	})

	return out
}

// Contains reports whether the address is inside the configured allocation
// range. Reserved addresses are still considered part of the range.
func (p *Pool) Contains(addr netip.Addr) bool {
	if !addr.Is4() {
		return false
	}

	addr = addr.Unmap()

	return addr.Compare(p.start) >= 0 &&
		addr.Compare(p.end) <= 0
}

// IsReserved reports whether the address has been explicitly reserved.
func (p *Pool) IsReserved(addr netip.Addr) bool {
	if !addr.Is4() {
		return false
	}

	_, ok := p.reserved[addr.Unmap()]
	return ok
}

// HostCIDR returns the address as an IPv4 host route.
func (p *Pool) HostCIDR(addr netip.Addr) string {
	return netip.PrefixFrom(addr.Unmap(), 32).String()
}

// NextAvailable returns the lowest non-reserved address that does not exist in
// the supplied used-address set.
func (p *Pool) NextAvailable(
	used map[netip.Addr]struct{},
) (netip.Addr, error) {
	first := addrToUint32(p.start)
	last := addrToUint32(p.end)

	for current := first; current <= last; current++ {
		addr := uint32ToAddr(current)

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
	bytes := addr.Unmap().As4()

	return uint32(bytes[0])<<24 |
		uint32(bytes[1])<<16 |
		uint32(bytes[2])<<8 |
		uint32(bytes[3])
}

func uint32ToAddr(value uint32) netip.Addr {
	return netip.AddrFrom4([4]byte{
		byte(value >> 24),
		byte(value >> 16),
		byte(value >> 8),
		byte(value),
	})
}
