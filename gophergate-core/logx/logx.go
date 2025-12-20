package logx

import (
	"errors"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/Cyrof/GopherGate/gophergate-core/envx"
	"github.com/Cyrof/GopherGate/gophergate-core/paths"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Mode int

const (
	Auto Mode = iota // dev -> console, else file
	Console
	File
)

type Config struct {
	App        string
	Mode       Mode
	Level      zapcore.Level
	FileRotate Rotate
}

type Rotate struct {
	MaxSizeMB  int  // 5
	MaxBackups int  // 3
	MaxAgeDays int  // 28
	Compress   bool // true
}

func Default(app string) Config {
	return Config{
		App:        app,
		Mode:       Auto,
		Level:      zap.InfoLevel,
		FileRotate: Rotate{MaxSizeMB: 5, MaxBackups: 3, MaxAgeDays: 28, Compress: true},
	}
}

func isKubernetes() bool {
	return os.Getenv("K3S_SERVICE_HOST") != ""
}

func Init(cfg Config) (*zap.SugaredLogger, func()) {

	// decide mode
	mode := cfg.Mode
	if mode == Auto {
		if envx.IsDev() || isKubernetes() {
			mode = Console
		} else {
			mode = File
		}
	}

	encCfg := zapcore.EncoderConfig{
		TimeKey:      "ts",
		LevelKey:     "level",
		CallerKey:    "caller",
		MessageKey:   "msg",
		EncodeLevel:  zapcore.CapitalLevelEncoder,
		EncodeTime:   zapcore.TimeEncoderOfLayout(time.RFC3339),
		EncodeCaller: zapcore.ShortCallerEncoder,
	}

	var core zapcore.Core

	switch mode {
	case Console:
		dev := zap.NewDevelopmentEncoderConfig()
		dev.EncodeLevel = zapcore.CapitalColorLevelEncoder
		dev.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
		core = zapcore.NewCore(zapcore.NewConsoleEncoder(dev), zapcore.AddSync(os.Stdout), cfg.Level)
	case File:
		d := paths.ForApp(cfg.App)
		if err := d.Ensure(); err != nil {
			// fallback to stdout if file path not writable
			core = zapcore.NewCore(zapcore.NewJSONEncoder(encCfg), zapcore.AddSync(os.Stdout), cfg.Level)
			break
		}
		w := zapcore.AddSync(&lumberjack.Logger{
			Filename:   d.LogFile,
			MaxSize:    cfg.FileRotate.MaxSizeMB,
			MaxBackups: cfg.FileRotate.MaxBackups,
			MaxAge:     cfg.FileRotate.MaxAgeDays,
			Compress:   cfg.FileRotate.Compress,
		})
		core = zapcore.NewCore(zapcore.NewConsoleEncoder(encCfg), w, cfg.Level)
	}

	z := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	flush := func() {
		if err := z.Sync(); err != nil && !ignorableSyncErr(err) {
			_, _ = os.Stderr.WriteString("logger sync error: " + err.Error() + "\n")
		}
	}
	return z.Sugar(), flush
}

func ignorableSyncErr(err error) bool {
	if err == nil {
		return true
	}

	// There is a common error where "invalid argument" is return when syncing stdout/stderr
	if runtime.GOOS == "windows" && strings.Contains(strings.ToLower(err.Error()), "invalid argument") {
		return true
	}
	// some writer may wrap EOF/closed pipe
	if errors.Is(err, os.ErrClosed) {
		return true
	}

	return false
}
