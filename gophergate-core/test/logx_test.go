package gophergatecore_test

import (
	"bufio"
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Cyrof/GopherGate/gophergate-core/logx"
	"github.com/stretchr/testify/require"
)

func captureStdout(t *testing.T) (restore func(), read func() string) {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)

	os.Stdout = w
	return func() {
			_ = w.Close()
			os.Stdout = orig
		}, func() string {
			_ = w.Close()
			var buf bytes.Buffer
			sc := bufio.NewScanner(r)
			for sc.Scan() {
				buf.Write(sc.Bytes())
				buf.WriteByte('\n')
			}
			_ = r.Close()
			return buf.String()
		}
}

func TestLogx_Console_WritesToStdout(t *testing.T) {
	restore, read := captureStdout(t)
	defer restore()

	cfg := logx.Default("gg-test")
	cfg.Mode = logx.Console
	cfg.Level = 0

	log, flush := logx.Init(cfg)
	defer flush()

	log.Infow("hello-console", "k", 1)

	out := read()
	require.Contains(t, out, "hello-console")
}

func TestLogx_Auto_RespectsDevEnv_UsesConsole(t *testing.T) {
	t.Setenv("GOPHERGATE_ENV", "dev")
	restore, read := captureStdout(t)
	defer restore()

	cfg := logx.Default("gg-test-auto-dev")
	cfg.Mode = logx.Auto
	log, flush := logx.Init(cfg)
	defer flush()

	log.Infow("auto-dev-console")

	out := read()
	require.Contains(t, out, "auto-dev-console")
}

func TestLogx_File_WritesRotatingLogUnderXDG(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_STATE_HOME", filepath.Join(tmp, "state"))
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, "config"))
	t.Setenv("GOPHERGATE_ENV", "prod")

	cfg := logx.Default("gg-test-file")
	cfg.Mode = logx.Auto

	log, flush := logx.Init(cfg)
	defer flush()

	const msg = "hello-file-sink"
	log.Infow(msg)

	var logFiles []string
	_ = filepath.WalkDir(tmp, func(path string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(strings.ToLower(d.Name()), ".log") {
			logFiles = append(logFiles, path)
		}
		return nil
	})

	require.NotEmpty(t, logFiles, "expected at least one .log file under XDG state dir")

	foundMsg := false
	for _, f := range logFiles {
		b, _ := os.ReadFile(f)
		if bytes.Contains(b, []byte(msg)) {
			foundMsg = true
			break
		}
	}
	require.True(t, foundMsg, "expected a .log file to contain the log message")
}
