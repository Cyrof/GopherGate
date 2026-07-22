package ippool

import "testing"

func TestFromEnvDisabledWithoutCIDR(t *testing.T) {
	t.Setenv(EnvCIDR, "")
	pool, err := FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if pool != nil {
		t.Fatalf("pool=%v want nil", pool)
	}
}

func TestFromEnvLoadsRangeAndReservedAddresses(t *testing.T) {
	t.Setenv(EnvCIDR, "10.20.0.0/24")
	t.Setenv(EnvStart, "10.20.0.10")
	t.Setenv(EnvEnd, "10.20.0.20")
	t.Setenv(EnvReserved, "10.20.0.11, 10.20.0.12")

	pool, err := FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if got, want := pool.Capacity(), uint64(9); got != want {
		t.Fatalf("capacity=%d want=%d", got, want)
	}
	if got, want := len(pool.Reserved()), 2; got != want {
		t.Fatalf("reserved=%d want %d", got, want)
	}
}
