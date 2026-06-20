package viper

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
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

func waitForConfig(t *testing.T, cm *ConfigManager[testConfig], want testConfig) {
	t.Helper()

	deadline := time.NewTimer(3 * time.Second)
	defer deadline.Stop()

	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()

	for {
		if got := cm.GetConfig(); got == want {
			return
		}

		select {
		case err := <-cm.Errors():
			t.Fatalf("unexpected reload error: %v", err)
		case <-ticker.C:
		case <-deadline.C:
			t.Fatalf("GetConfig() = %#v, want %#v", cm.GetConfig(), want)
		}
	}
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

func TestConfigManagerSequentialLoadConfigOnlyReloadsCurrentFile(t *testing.T) {
	tests := []struct {
		name     string
		filetype string
	}{
		{
			name:     "yaml",
			filetype: "yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			firstPath := filepath.Join(dir, "first.yaml")
			secondPath := filepath.Join(dir, "second.yaml")
			writeConfig(t, firstPath, "name: first\nport: 1001\n")
			writeConfig(t, secondPath, "name: second\nport: 2001\n")

			cfg := testConfig{}
			cm := NewConfigManager(&cfg)
			defer cm.Close()

			if err := cm.LoadConfig(firstPath, tt.filetype); err != nil {
				t.Fatalf("LoadConfig(first) error = %v", err)
			}
			firstSession := cm.session
			if err := cm.LoadConfig(secondPath, tt.filetype); err != nil {
				t.Fatalf("LoadConfig(second) error = %v", err)
			}

			writeConfig(t, firstPath, "name: stale\nport: 1002\n")
			cm.reload(firstSession)
			if got := cm.GetConfig(); got != (testConfig{Name: "second", Port: 2001}) {
				t.Fatalf("stale session changed config to %#v", got)
			}
			select {
			case <-cm.Watch():
				t.Fatal("stale session emitted reload notification")
			case err := <-cm.Errors():
				t.Fatalf("stale session emitted reload error: %v", err)
			default:
			}

			writeConfig(t, secondPath, "name: current\nport: 2002\n")
			waitForConfig(t, cm, testConfig{Name: "current", Port: 2002})

			if cfg != (testConfig{Name: "current", Port: 2002}) {
				t.Fatalf("provided pointer = %#v, want current config", cfg)
			}
		})
	}
}

func TestConfigManagerFailedLoadConfigPreservesCurrentSession(t *testing.T) {
	tests := []struct {
		name        string
		invalidBody string
	}{
		{
			name:        "invalid replacement",
			invalidBody: "name: [broken\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			currentPath := filepath.Join(dir, "current.yaml")
			invalidPath := filepath.Join(dir, "invalid.yaml")
			writeConfig(t, currentPath, "name: current\nport: 8080\n")
			writeConfig(t, invalidPath, tt.invalidBody)

			cfg := testConfig{}
			cm := NewConfigManager(&cfg)
			defer cm.Close()
			if err := cm.LoadConfig(currentPath, "yaml"); err != nil {
				t.Fatalf("LoadConfig(current) error = %v", err)
			}
			currentSession := cm.session

			if err := cm.LoadConfig(invalidPath, "yaml"); err == nil {
				t.Fatal("LoadConfig(invalid) error = nil, want error")
			}
			if cm.session != currentSession {
				t.Fatal("failed LoadConfig replaced current session")
			}
			if got := cm.GetConfig(); got != (testConfig{Name: "current", Port: 8080}) {
				t.Fatalf("GetConfig() = %#v, want current config", got)
			}

			writeConfig(t, currentPath, "name: reloaded\nport: 9090\n")
			waitForConfig(t, cm, testConfig{Name: "reloaded", Port: 9090})
		})
	}
}

func TestConfigManagerConcurrentLoadConfig(t *testing.T) {
	tests := []struct {
		name  string
		loads int
	}{
		{
			name:  "many valid files",
			loads: 16,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			paths := make([]string, tt.loads)
			wants := make(map[testConfig]struct{}, tt.loads)
			for i := range paths {
				paths[i] = filepath.Join(dir, fmt.Sprintf("config-%02d.yaml", i))
				want := testConfig{Name: fmt.Sprintf("config-%02d", i), Port: 8000 + i}
				wants[want] = struct{}{}
				writeConfig(t, paths[i], fmt.Sprintf("name: %s\nport: %d\n", want.Name, want.Port))
			}

			cfg := testConfig{}
			cm := NewConfigManager(&cfg)
			defer cm.Close()

			start := make(chan struct{})
			errs := make(chan error, tt.loads)
			var wg sync.WaitGroup
			for _, path := range paths {
				wg.Add(1)
				go func() {
					defer wg.Done()
					<-start
					errs <- cm.LoadConfig(path, "yaml")
				}()
			}

			close(start)
			wg.Wait()
			close(errs)

			for err := range errs {
				if err != nil {
					t.Fatalf("LoadConfig() error = %v", err)
				}
			}
			if _, ok := wants[cm.GetConfig()]; !ok {
				t.Fatalf("GetConfig() = %#v, want one complete loaded config", cm.GetConfig())
			}
			if cfg != cm.GetConfig() {
				t.Fatalf("provided pointer = %#v, GetConfig() = %#v", cfg, cm.GetConfig())
			}
		})
	}
}

func TestConfigManagerLoadConfigAndCloseRepeated(t *testing.T) {
	tests := []struct {
		name       string
		iterations int
	}{
		{
			name:       "load races with close",
			iterations: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.yaml")
			writeConfig(t, path, "name: app\nport: 8080\n")

			for i := 0; i < tt.iterations; i++ {
				cfg := testConfig{}
				cm := NewConfigManager(&cfg)
				start := make(chan struct{})
				loadErr := make(chan error, 1)
				var wg sync.WaitGroup
				wg.Add(2)

				go func() {
					defer wg.Done()
					<-start
					loadErr <- cm.LoadConfig(path, "yaml")
				}()
				go func() {
					defer wg.Done()
					<-start
					cm.Close()
				}()

				close(start)
				wg.Wait()

				err := <-loadErr
				if err != nil && !errors.Is(err, ErrAlreadyClosed) {
					t.Fatalf("iteration %d: LoadConfig() error = %v, want nil or ErrAlreadyClosed", i, err)
				}
				for attempt := 0; attempt < 3; attempt++ {
					if err := cm.LoadConfig(path, "yaml"); !errors.Is(err, ErrAlreadyClosed) {
						t.Fatalf("iteration %d attempt %d: LoadConfig() error = %v, want ErrAlreadyClosed", i, attempt, err)
					}
				}
			}
		})
	}
}

func TestConfigManagerGetConfigDuringReload(t *testing.T) {
	tests := []struct {
		name    string
		readers int
	}{
		{
			name:    "concurrent readers",
			readers: 8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.yaml")
			initial := testConfig{Name: "initial", Port: 8080}
			updated := testConfig{Name: "updated", Port: 9090, Debug: true}
			writeConfig(t, path, "name: initial\nport: 8080\n")

			cfg := testConfig{}
			cm := NewConfigManager(&cfg)
			defer cm.Close()
			if err := cm.LoadConfig(path, "yaml"); err != nil {
				t.Fatalf("LoadConfig() error = %v", err)
			}

			stop := make(chan struct{})
			badRead := make(chan testConfig, 1)
			var wg sync.WaitGroup
			for i := 0; i < tt.readers; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for {
						select {
						case <-stop:
							return
						default:
							got := cm.GetConfig()
							if got != initial && got != updated {
								select {
								case badRead <- got:
								default:
								}
								return
							}
						}
					}
				}()
			}

			writeConfig(t, path, "name: updated\nport: 9090\ndebug: true\n")
			waitForConfig(t, cm, updated)
			close(stop)
			wg.Wait()

			select {
			case got := <-badRead:
				t.Fatalf("GetConfig() observed partial config %#v", got)
			default:
			}
		})
	}
}
