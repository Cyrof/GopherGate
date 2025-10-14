create extension if not exists "uuid-ossp";

create table if not exists peers(
	id uuid primary key default uuid_generate_v4(),
	name text not null,
	public_key text not null,
	ip_address inet not null,
	allowed_ips inet[] default '{}',
	endpoint text,
	persistent_keepalive smallint,
	last_handshake timestamptz,
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now()
);

-- add uniques for wg agent
alter table peers
	add constraint uq_peers_public_key unique (public_key);

alter table peers
	add constraint uq_peers_ip_address unique (ip_address);

create index if not exists idx_peers_name on peers(name);

