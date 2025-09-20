package config

// this file should hold config setup for dev and prod 

import (
	"os"
	"strconv"

	"go.uber.org/zap"
)

type Config struct {
	Env			string
	HTTPAddr	string
	GRPCAddr	string
	TLS			bool
}

func Load(logger *zap.SugaredLogger) (*Config, error) {
	c := &Config{
		Env:      getenv(logger, "APP_ENV", "dev"),
		HTTPAddr: getenv(logger, "HTTP_ADDR", ":8080"),
		GRPCAddr: getenv(logger, "GRPC_ADDR", "127.0.0.1:5051"),
		TLS:      getbool(logger, "GRPC_TLS_ENABLE", false),
	}
	return c, nil
}

func getenv(log *zap.SugaredLogger, key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		log.Infow("config getenv", "key", key, "value", v, "source", "env")
		return v
	}
	log.Infow("config getenv", "key", key, "value", def, "source", "default")
	return def
}

func getbool(log *zap.SugaredLogger, key string, def bool) bool {
	if v, ok := os.LookupEnv(key); ok {
		b, err := strconv.ParseBool(v)
		if err != nil {
			log.Warnw("config getbool invalid", "key", key, "raw", v, "err", err, "using", def)
			return def
		}
		log.Infow("config getbool", "key", key, "value", b, "source", "env")
		return b
	}
	log.Infow("config getbool", "key", key, "value", def, "source", "default")
	return def
}