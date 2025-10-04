package dbx

import (
	"time"
)

type Config struct {
	App string

	DSN          string
	Host         string
	Port         int
	User         string
	DBName       string
	Password     string
	PasswordFile string
	SSLMode      string

	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
	HealthCheckFreq time.Duration
	ConnectTimeout  time.Duration
}

func Default(app string) Config {
	return Config{
		App:    app,
		Host:   "127.0.0.1",
		Port:   5432,
		User:   "gg_admin",
		DBName: "gophergate",
		// currently tls is disabled for dev
		SSLMode:         "disabled",
		MaxConns:        4,
		MinConns:        0,
		MaxConnLifetime: 30 * time.Minute,
		MaxConnIdleTime: 5 * time.Minute,
		HealthCheckFreq: 30 * time.Second,
		ConnectTimeout:  5 * time.Second,
	}
}

func pickStr(v, d string) string {
	if v != "" {
		return v
	}
	return d
}

func pickI32(v, d int32) int32 {
	if v != 0 {
		return v
	}
	return d
}

func pickDur(v, d time.Duration) time.Duration {
	if v != 0 {
		return v
	}
	return d
}
