create table if not exists peer_traffic_snapshots (
  id uuid primary key default uuid_generate_v4 (),
  iface text not null,
  public_key text not null,
  rx_bytes bigint not null,
  tx_bytes bigint not null,
  total_bytes bigint not null,
  recorded_at timestamptz not null default now ()
);

create index if not exists idx_peer_traffic_snapshots_peer_time on peer_traffic_snapshots (public_key, recorded_at desc);

create index if not exists idx_peer_traffic_snapshots_iface_time on peer_traffic_snapshots (iface, recorded_at desc);
