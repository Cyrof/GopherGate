package ippool

import (
	"os"
	"strings"
)

const (
	EnvCIDR     = "WG_IP_POOL_CIDR"
	EnvStart    = "WG_IP_POOL_START"
	EnvEnd      = "WG_IP_POOL_END"
	EnvReserved = "WG_IP_POOL_RESERVED"
)

// FromEnv returns nil when no pool CIDR is configured
func FromEnv() (*Pool, error) {
	cidr := strings.TrimSpace(os.Getenv(EnvCIDR))
	if cidr == "" {
		return nil, nil
	}

	var reserved []string
	if raw := strings.TrimSpace(os.Getenv(EnvReserved)); raw != "" {
		reserved = strings.Split(raw, ",")
	}

	return New(cidr, os.Getenv(EnvStart), os.Getenv(EnvEnd), reserved)
}
