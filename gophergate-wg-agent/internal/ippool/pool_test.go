package ippool

import (
	"errors"
	"net/netip"
	"testing"
)

func TestPoolNextAvailable(t *testing.T) {
	p, err := New("10.13.13.0/24", "10.13.13.2", "10.13.13.5", []string{"10.13.13.3"})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := p.Capacity(), uint64(3); got != want {
		t.Fatalf("capacity=%d want=%d", got, want)
	}

	used := map[netip.Addr]struct{}{
		netip.MustParseAddr("10.13.13.2"): {},
		netip.MustParseAddr("10.13.13.4"): {},
	}
	got, err := p.NextAvailable(used)
	if err != nil {
		t.Fatal(err)
	}
	if want := netip.MustParseAddr("10.13.13.5"); got != want {
		t.Fatalf("next=%s want=%s", got, want)
	}
}

func TestPoolExhausted(t *testing.T) {
	p, err := New("10.0.0.0/30", "", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	used := map[netip.Addr]struct{}{
		netip.MustParseAddr("10.0.0.1"): {},
		netip.MustParseAddr("10.0.0.2"): {},
	}
	_, err = p.NextAvailable(used)
	if !errors.Is(err, ErrExhausted) {
		t.Fatalf("err=%v want ErrExhausted", err)
	}
}

func TestPoolRejectsOutOfRangeReservedIP(t *testing.T) {
	_, err := New("10.0.0.0/24", "10.0.0.10", "10.0.0.20", []string{"10.0.0.1"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestPoolUsesDefaultHostRange(t *testing.T) {
	p, err := New("192.0.2.0/30", "", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := p.Start(), netip.MustParseAddr("192.0.2.1"); got != want {
		t.Fatalf("start=%s want=%s", got, want)
	}
	if got, want := p.End(), netip.MustParseAddr("192.0.2.2"); got != want {
		t.Fatalf("end=%s want=%s", got, want)
	}
	if got, want := p.Capacity(), uint64(2); got != want {
		t.Fatalf("capacity=%d want=%d", got, want)
	}
}

func TestPoolRejectsNetworkAndBroadcastAddresses(t *testing.T) {
	tests := []struct {
		name  string
		start string
		end   string
	}{
		{name: "network", start: "10.0.0.0", end: "10.0.0.10"},
		{name: "broadcast", start: "10.0.0.10", end: "10.0.0.255"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := New("10.0.0.0/24", tt.start, tt.end, nil); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestPoolRejectsIPv6(t *testing.T) {
	if _, err := New("fd00::/64", "", "", nil); err == nil {
		t.Fatal("expected error")
	}
}
