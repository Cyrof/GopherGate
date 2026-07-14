package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Cyrof/GopherGate/gophergate-wg-agent/internal/data"
	"github.com/Cyrof/GopherGate/gophergate-wg-agent/internal/ippool"
)

type IPPoolStatus struct {
	CIDR            string
	RangeStart      string
	RangeEnd        string
	Capacity        uint64
	Allocated       uint64
	Available       uint64
	NextAvailableIP string
	ReservedIPs     []string
}

func BuildIPPoolStatus(ctx context.Context, repo *data.Repository, pool *ippool.Pool) (*IPPoolStatus, error) {
	if pool == nil {
		return nil, fmt.Errorf("IP pool is not configured")
	}
	if repo == nil {
		return nil, fmt.Errorf("database repository is required for IP pool status")
	}

	used, err := repo.IPAddressesInRange(ctx, pool.Start(), pool.End())
	if err != nil {
		return nil, err
	}

	var allocated uint64
	for addr := range used {
		if pool.Contains(addr) && !pool.IsReserved(addr) {
			allocated++
		}
	}
	capacity := pool.Capacity()
	if allocated > capacity {
		allocated = capacity
	}

	next := ""
	if addr, err := pool.NextAvailable(used); err == nil {
		next = addr.String()
	} else if !wgPoolExhausted(err) {
		return nil, err
	}

	reserved := pool.Reserved()
	reservedStrings := make([]string, 0, len(reserved))
	for _, addr := range reserved {
		reservedStrings = append(reservedStrings, addr.String())
	}

	return &IPPoolStatus{
		CIDR:            pool.CIDR(),
		RangeStart:      pool.Start().String(),
		RangeEnd:        pool.End().String(),
		Capacity:        capacity,
		Allocated:       allocated,
		Available:       capacity - allocated,
		NextAvailableIP: next,
		ReservedIPs:     reservedStrings,
	}, nil
}

func wgPoolExhausted(err error) bool {
	return errors.Is(err, ippool.ErrExhausted)
}
