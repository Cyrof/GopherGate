package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	errPeerIPPoolMissing = errors.New("peer ip pool is not configured")
	errPeerIPPoolInvalid = errors.New("peer ip pool is invalid")
	errPeerIPInvalid     = errors.New("peer ip is invalid")
	errPeerIPOutOfRange  = errors.New("peer ip is outside configured pool")
	errPeerIPAlreadyUsed = errors.New("peer ip is already assigned")
	errPeerIPPoolFull    = errors.New("peer ip pool has no available addresses")
)

func peerActionError(action string, err error) string {
	if err == nil {
		return ""
	}

	msg := err.Error()
	code := codes.Unknown

	if st, ok := status.FromError(err); ok {
		msg = st.Message()
		code = st.Code()
	}

	lower := strings.ToLower(msg)

	switch {
	case strings.Contains(lower, "parse public key"),
		strings.Contains(lower, "illegal base64"),
		strings.Contains(lower, "base64-encoded key"):
		return "Invalid public key. Please enter a valid WireGuard public key."

	case strings.Contains(lower, "invalid allowed ip"),
		strings.Contains(lower, "invalid allowed"),
		strings.Contains(lower, "invalid cidr"),
		strings.Contains(lower, "allowed cidr"),
		strings.Contains(lower, "parse prefix"),
		strings.Contains(lower, "parseprefix"),
		strings.Contains(lower, "netip"):
		return "Invalid Allowed IP. Please enter a valid CIDR value, for example 10.8.0.25."

	case strings.Contains(lower, "uq_peers_ip_address"),
		strings.Contains(lower, "duplicate key value"),
		strings.Contains(lower, "ip_address"),
		strings.Contains(lower, "already assigned"):
		return "Allowed IP already exists. Please use a different IP or leave it blank to auto-assign one."

	case strings.Contains(lower, "invalid keepalive"):
		return "Invalid keepalive value. Please enter a number of seconds, for example 25."

	case strings.Contains(lower, "invalid endpoint"):
		return "Invalid endpoint. Leave it blank for roaming peers, or use the formate IP:port, for example 192.168.1.100:51820."

	case strings.Contains(lower, "public key already exists"),
		strings.Contains(lower, "peer already exists"):
		return "A peer with this public key already exists."

	case strings.Contains(lower, "not found"):
		return "Peer not found. It may have already been removed."
	}

	switch code {
	case codes.InvalidArgument:
		return fmt.Sprintf("Failed to %s peer. Please check the peer details and try again.", action)
	case codes.AlreadyExists:
		return fmt.Sprintf("Failed to %s peer because it already exists.", action)
	case codes.NotFound:
		return "Peer not found. It may have already been removed."
	case codes.Unavailable:
		return "GopherGate backend is currently unavailable. Please try again later."
	case codes.DeadlineExceeded:
		return "The request timed out while contacting the GopherGate backend."
	default:
		return fmt.Sprintf("Failed to %s peer. Please check the details and try again.", action)
	}
}

func peerErrorHTTPStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}

	msg := err.Error()
	code := codes.Unknown

	if st, ok := status.FromError(err); ok {
		msg = st.Message()
		code = st.Code()
	}

	lower := strings.ToLower(msg)

	switch {
	case strings.Contains(lower, "parse public key"),
		strings.Contains(lower, "illegal base64"),
		strings.Contains(lower, "base64-encoded key"),
		strings.Contains(lower, "invalid keepalive"),
		strings.Contains(lower, "invalid endpoint"),
		strings.Contains(lower, "invalid allowed ip"),
		strings.Contains(lower, "invalid allowed"),
		strings.Contains(lower, "invalid cidr"),
		strings.Contains(lower, "allowed cidr"),
		strings.Contains(lower, "parse prefix"),
		strings.Contains(lower, "parseprefix"),
		strings.Contains(lower, "netip"):
		return http.StatusBadRequest

	case strings.Contains(lower, "uq_peers_ip_address"),
		strings.Contains(lower, "duplicate key value"),
		strings.Contains(lower, "already exists"),
		strings.Contains(lower, "already assigned"):
		return http.StatusConflict
	}

	switch code {
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.NotFound:
		return http.StatusNotFound
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	default:
		return http.StatusInternalServerError
	}
}

func peerAllowedIPErrorMessage(err error) string {
	switch {
	case errors.Is(err, errPeerIPPoolMissing):
		return "Peer IP pool is not configured. Please set GOPHERGATE_PEER_IP_POOL_CIDR."

	case errors.Is(err, errPeerIPPoolInvalid):
		return "Peer IP pool is invalid. Please check GOPHERGATE_PEER_IP_POOL_CIDR."

	case errors.Is(err, errPeerIPInvalid):
		return "Invalid Allowed IP. Leave it blank to auto-assign, or enter a valid IP such as 10.8.0.25."

	case errors.Is(err, errPeerIPOutOfRange):
		return "Allowed IP is outside the configured peer IP pool."

	case errors.Is(err, errPeerIPAlreadyUsed):
		return "Allowed IP already exists. Please use a different IP or leave it blank to auto-assign one."

	case errors.Is(err, errPeerIPPoolFull):
		return "No available IP addresses remain in the configured peer IP pool."

	default:
		return "Invalid Allowed IP. Leave it blank to auto-assign, or enter a valid IP such as 10.8.0.25."
	}
}
