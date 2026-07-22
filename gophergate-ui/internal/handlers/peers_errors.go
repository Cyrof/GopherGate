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
	errPeerIPInvalid     = errors.New("peer ip is invalid")
	errPeerIPAlreadyUsed = errors.New("peer ip is already assigned")
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
	case strings.Contains(lower, "ip pool is not configured"),
		strings.Contains(lower, "automatic ip assignment requires a configured ip pool"):
		return "Automatic IP assignment is not configured on the WireGuard agent. Enter a manual IP or configure the agent IP pool."

	case strings.Contains(lower, "no free ip addresses"),
		strings.Contains(lower, "ip pool exhausted"),
		strings.Contains(lower, "no available ip"):
		return "No available IP addresses remain in the configured pool. Enter a manual IP or expand the agent IP pool."

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
		return "Invalid Allowed IP. Leave it blank for automatic assignment, or enter a host IP such as 10.13.13.25."

	case strings.Contains(lower, "uq_peers_ip_address"),
		strings.Contains(lower, "duplicate key value"),
		strings.Contains(lower, "ip_address"),
		strings.Contains(lower, "already assigned"):
		return "Allowed IP already exists. Please use a different IP or leave it blank to auto-assign one."

	case strings.Contains(lower, "invalid keepalive"):
		return "Invalid keepalive value. Please enter a number of seconds, for example 25."

	case strings.Contains(lower, "invalid endpoint"):
		return "Invalid endpoint. Leave it blank for roaming peers, or use the format IP:port, for example 192.168.1.100:51820."

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
	case codes.FailedPrecondition:
		return "Automatic IP assignment is unavailable because the WireGuard agent IP pool is not configured."
	case codes.ResourceExhausted:
		return "No available IP addresses remain in the configured WireGuard IP pool."
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
		strings.Contains(lower, "already assigned"),
		strings.Contains(lower, "ip pool exhausted"),
		strings.Contains(lower, "no free ip addresses"):
		return http.StatusConflict
	}

	switch code {
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.AlreadyExists, codes.ResourceExhausted:
		return http.StatusConflict
	case codes.FailedPrecondition, codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.NotFound:
		return http.StatusNotFound
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	default:
		return http.StatusInternalServerError
	}
}

func peerAllowedIPErrorMessage(err error) string {
	switch {
	case errors.Is(err, errPeerIPInvalid):
		return "Invalid Allowed IP. Leave it blank to auto-assign, or enter a host IP such as 10.13.13.25."

	case errors.Is(err, errPeerIPAlreadyUsed):
		return "Allowed IP already exists. Please use a different IP or leave it blank to auto-assign one."

	default:
		return "Invalid Allowed IP. Leave it blank to auto-assign, or enter a host IP such as 10.13.13.25."
	}
}
