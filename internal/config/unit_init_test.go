//go:build scopesentry_unit

package config

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
)

// Run the real package init in a fresh process: t.Setenv is too late for init.
func TestUnitConfigBootstrap(t *testing.T) {
	if os.Getenv("SCOPESENTRY_UNIT_CONFIG_CHILD") == "1" {
		want := Config{System: SystemConfig{Timezone: "UTC"}, Log: LogConfig{Level: "info"}}
		if !reflect.DeepEqual(GlobalConfig, want) {
			t.Fatal("unit config must ignore environment/files and use safe defaults")
		}
		return
	}

	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(self)
	if err != nil {
		t.Fatal(err)
	}
	for _, poisonedFile := range []bool{false, true} {
		name := "fresh"
		if poisonedFile {
			name = "poisoned config"
		}
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			cwd, bin := filepath.Join(root, "cwd"), filepath.Join(root, "bin")
			for _, dir := range []string{cwd, bin} {
				if err := os.Mkdir(dir, 0700); err != nil {
					t.Fatal(err)
				}
			}
			executable := filepath.Join(bin, "config.test")
			if err := os.WriteFile(executable, data, 0700); err != nil {
				t.Fatal(err)
			}
			const poison = "system:\n  timezone: poisoned-file\n  plugin_key: poisoned-file\n"
			configPath := filepath.Join(bin, "config.yaml")
			if poisonedFile {
				if err := os.WriteFile(configPath, []byte(poison), 0600); err != nil {
					t.Fatal(err)
				}
			}
			cmd := exec.Command(executable, "-test.run=^TestUnitConfigBootstrap$", "-test.v")
			cmd.Dir = cwd
			cmd.Env = []string{
				"SCOPESENTRY_UNIT_CONFIG_CHILD=1", "HOME=" + root, "PATH=" + os.Getenv("PATH"), "TZ=UTC",
				"TIMEZONE=poisoned-env", "MONGODB_IP=192.0.2.1", "MONGODB_PORT=27099",
				"MONGODB_DATABASE=poisoned-env", "MONGODB_USER=poisoned-env", "MONGODB_PASSWORD=poisoned-env",
				"REDIS_IP=192.0.2.2", "REDIS_PORT=6399", "REDIS_PASSWORD=poisoned-env",
			}
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Errorf("fresh process: %v\n%s", err, output)
			}
			entries, err := os.ReadDir(cwd)
			if err != nil || len(entries) != 0 {
				t.Errorf("config init wrote into fresh cwd (including PLUGINKEY): entries=%v err=%v", entries, err)
			}
			entries, err = os.ReadDir(bin)
			wantEntries := 1
			if poisonedFile {
				wantEntries = 2
				got, err := os.ReadFile(configPath)
				if err != nil || string(got) != poison {
					t.Errorf("config init modified existing config: err=%v", err)
				}
			}
			if err != nil || len(entries) != wantEntries {
				t.Errorf("config init wrote beside executable: entries=%v err=%v", entries, err)
			}
		})
	}
}
