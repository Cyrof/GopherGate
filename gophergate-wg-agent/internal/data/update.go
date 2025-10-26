package data

import (
	"context"
	"fmt"
	"strings"
)

func (r *Repository) UpdateByPublicKey(ctx context.Context, pubkey string, in UpdatePeerDBInput) (UpdatePeerDBResult, error) {
	var (
		setParts []string
		args     []any
	)
	add := func(expr string, v any) {
		setParts = append(setParts, expr)
		args = append(args, v)
	}

	// allowed IPs
	var changeAllowed bool
	switch {
	case len(in.ReplaceAllowed) > 0:
		add(`allowed_ips = $`+fmt.Sprint(len(args)+1)+`::inet[]`, in.ReplaceAllowed)
		changeAllowed = true

	case len(in.AppendAllowed) > 0:
		add(`allowed_ips = (
			select array(
				select distinct x from (
				select unnest(coalesce(allowed_ips, '{}::inet[]')) as x
					union all
					select unnest($`+fmt.Sprint(len(args)+1)+`::inet[]) as x
				) t
			)
		)`, in.AppendAllowed)
		changeAllowed = true
	}

	// endpoint
	changeEndpoint := false
	if in.SetEndpoint {
		changeEndpoint = true
		if in.Endpoint == nil || strings.TrimSpace(*in.Endpoint) == "" {
			setParts = append(setParts, `endpoint = NULL`)
		} else {
			add(`endpoint = $`+fmt.Sprint(len(args)+1), *in.Endpoint)
		}
	}

	// keepalive
	changeKeepalive := false
	if in.SetKeepalive {
		changeKeepalive = true
		if in.Keepalive == nil || *in.Keepalive <= 0 {
			setParts = append(setParts, `persistent_keepalive = NULL`)
		} else {
			add(`persistent_keepalive = $`+fmt.Sprint(len(args)+1), *in.Keepalive)
		}
	}

	// nothing to do?
	if len(setParts) == 0 {
		return UpdatePeerDBResult{}, nil
	}

	// always bump updated_at
	setParts = append(setParts, `updated_at = NOW()`)

	q := `update peers set ` + strings.Join(setParts, ", ") + ` where public_key = $` + fmt.Sprint(len(args)+1) + `;`
	args = append(args, pubkey)

	ct, err := r.db.Exec(ctx, q, args...)
	if err != nil {
		return UpdatePeerDBResult{}, err
	}
	if ct.RowsAffected() == 0 {
		return UpdatePeerDBResult{}, fmt.Errorf("no row matched public key")
	}

	return UpdatePeerDBResult{
		ChangeAllowed:   changeAllowed,
		ChangeEndpoint:  changeEndpoint,
		ChangeKeepalive: changeKeepalive,
	}, nil
}
