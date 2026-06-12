package viper

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type testConfig struct {
	Name  string `mapstructure:"name"`
	Port  int    `mapstructure:"port"`
	Debug bool   `mapstructure:"debug"`
}

func writeConfig(t *testing.T, path string, body string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
}

func waitForReload(t *testing.T, cm *ConfigManager[testConfig]) {
	t.Helper()

	select {
	case <-cm.Watch():
		return
	case err := <-cm.Errors():
		t.Fatalf("unexpected reload error: %v", err)
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for reload notification")
	}
}

func waitForError(t *testing.T, cm *ConfigManager[testConfig]) error {
	t.Helper()

	select {
	case err := <-cm.Errors():
		if err == nil {
			t.Fatal("got nil reload error")
		}
		return err
	case <-cm.Watch():
		t.Fatal("unexpected successful reload notification")
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for reload error")
	}

	return nil
}

func TestConfigManagerLoadConfig(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *testConfig
		body     string
		filetype string
		want     testConfig
		wantErr  bool
	}{
		{
			name:     "loads yaml into provided pointer",
			cfg:      &testConfig{},
			body:     "name: app\nport: 8080\ndebug: true\n",
			filetype: "yaml",
			want:     testConfig{Name: "app", Port: 8080, Debug: true},
		},
		{
			name:     "rejects invalid yaml",
			cfg:      &testConfig{},
			body:     "name: [broken\n",
			filetype: "yaml",
			wantErr:  true,
		},
		{
			name:     "rejects nil config pointer",
			cfg:      nil,
			body:     "name: app\n",
			filetype: "yaml",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.yaml")
			writeConfig(t, path, tt.body)

			cm := NewConfigManager(tt.cfg)
			defer cm.Close()

			err := cm.LoadConfig(path, tt.filetype)
			if (err != nil) != tt.wantErr {
				t.Fatalf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				if tt.cfg == nil && !errors.Is(err, ErrNilConfig) {
					t.Fatalf("LoadConfig() error = %v, want ErrNilConfig", err)
				}
				return
			}

			if got := cm.GetConfig(); got != tt.want {
				t.Fatalf("GetConfig() = %#v, want %#v", got, tt.want)
			}
			if *tt.cfg != tt.want {
				t.Fatalf("provided pointer = %#v, want %#v", *tt.cfg, tt.want)
			}
		})
	}
}

func TestConfigManagerLoadConfigMissingFile(t *testing.T) {
	cfg := testConfig{}
	cm := NewConfigManager(&cfg)
	defer cm.Close()

	err := cm.LoadConfig(filepath.Join(t.TempDir(), "missing.yaml"), "yaml")
	if err == nil {
		t.Fatal("LoadConfig() error = nil, want error")
	}
}

func TestConfigManagerReloadSuccess(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	writeConfig(t, path, "name: app\nport: 8080\ndebug: false\n")

	cfg := testConfig{}
	cm := NewConfigManager(&cfg)
	defer cm.Close()

	if err := cm.LoadConfig(path, "yaml"); err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	writeConfig(t, path, "name: worker\nport: 9090\ndebug: true\n")
	waitForReload(t, cm)

	want := testConfig{Name: "worker", Port: 9090, Debug: true}
	if got := cm.GetConfig(); got != want {
		t.Fatalf("GetConfig() = %#v, want %#v", got, want)
	}
	if cfg != want {
		t.Fatalf("provided pointer = %#v, want %#v", cfg, want)
	}
}

func TestConfigManagerReloadErrorKeepsCurrentConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	writeConfig(t, path, "name: app\nport: 8080\ndebug: false\n")

	cfg := testConfig{}
	cm := NewConfigManager(&cfg)
	defer cm.Close()

	if err := cm.LoadConfig(path, "yaml"); err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	writeConfig(t, path, "name: [broken\n")
	err := waitForError(t, cm)
	if !strings.Contains(err.Error(), "reload config") {
		t.Fatalf("reload error = %v, want context containing reload config", err)
	}

	want := testConfig{Name: "app", Port: 8080}
	if got := cm.GetConfig(); got != want {
		t.Fatalf("GetConfig() = %#v, want %#v", got, want)
	}
	if cfg != want {
		t.Fatalf("provided pointer = %#v, want %#v", cfg, want)
	}
}

func TestConfigManagerCloseIsIdempotent(t *testing.T) {
	cfg := testConfig{}
	cm := NewConfigManager(&cfg)

	cm.Close()
	cm.Close()
}

func TestConfigManagerLoadAfterClose(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	writeConfig(t, path, "name: app\n")

	cfg := testConfig{}
	cm := NewConfigManager(&cfg)
	cm.Close()

	err := cm.LoadConfig(path, "yaml")
	if !errors.Is(err, ErrAlreadyClosed) {
		t.Fatalf("LoadConfig() error = %v, want ErrAlreadyClosed", err)
	}
}
