package paths

import (
	"os"
	"path/filepath"
)

const (
	AppAgent = "gophergate-wg-agent"
	AppUI    = "gophergate-ui"
)

type AppDirs struct {
	Base    string
	AppRoot string
	LogsDir string
	LogFile string
}

func BaseDir() string {
	home, _ := os.UserHomeDir()
	if home == "" {
		home = "/tmp"
	}
	xdg := os.Getenv("XDG_STATE_HOME")
	if xdg == "" {
		xdg = filepath.Join(home, ".local", "state")
	}
	return filepath.Join(xdg, "gophergate")
}

func ForApp(app string) *AppDirs {
	base := BaseDir()
	root := filepath.Join(base, app)
	return &AppDirs{
		Base:    base,
		AppRoot: root,
		LogsDir: filepath.Join(root, "logs"),
		LogFile: filepath.Join(root, "logs", app+".log"),
	}
}

func (d *AppDirs) Ensure() error {
	for _, p := range []string{d.AppRoot, d.LogsDir} {
		if err := os.MkdirAll(p, 0o755); err != nil {
			return err
		}
	}
	return nil
}
