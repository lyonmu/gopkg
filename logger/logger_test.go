package logger

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultConfigCreatesConsoleLoggerWithoutFileOutput(t *testing.T) {
	cfg := DefaultConfig()

	log, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if log == nil {
		t.Fatal("New() returned nil logger")
	}

	if cfg.Output.File.Enabled {
		t.Fatal("DefaultConfig() enabled file output")
	}
}

func TestNewDefault(t *testing.T) {
	log, err := NewDefault()
	if err != nil {
		t.Fatalf("NewDefault() error = %v", err)
	}
	if log == nil {
		t.Fatal("NewDefault() returned nil logger")
	}
}

func TestNewSupportsLevels(t *testing.T) {
	tests := []struct {
		name  string
		level Level
	}{
		{name: "debug", level: DebugLevel},
		{name: "info", level: InfoLevel},
		{name: "warn", level: WarnLevel},
		{name: "error", level: ErrorLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Level = tt.level

			log, err := New(cfg)
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}
			if log == nil {
				t.Fatal("New() returned nil logger")
			}
		})
	}
}

func TestNewReturnsErrorWhenLogDirectoryParentIsFile(t *testing.T) {
	parent := filepath.Join(t.TempDir(), "parent")
	if err := os.WriteFile(parent, []byte("not a directory"), 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	cfg := DefaultConfig()
	cfg.Output.Console.Enabled = false
	cfg.Output.File.Enabled = true
	cfg.Output.File.Path = filepath.Join(parent, "logs")

	_, err := New(cfg)
	if err == nil {
		t.Fatal("New() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "create log directory") {
		t.Fatalf("New() error = %q, want it to contain %q", err, "create log directory")
	}
}

func TestNewWritesFileWhenFileOutputEnabled(t *testing.T) {
	dir := t.TempDir()
	cfg := DefaultConfig()
	cfg.Module = "worker"
	cfg.Output.Console.Enabled = false
	cfg.Output.File.Enabled = true
	cfg.Output.File.Path = dir

	log, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	log.Info("file output works")
	_ = log.Sync()

	body := readLogFile(t, dir, "worker")
	if !strings.Contains(body, "file output works") {
		t.Fatalf("log file body = %q, want message", body)
	}
	if !strings.Contains(body, "[worker] INFO") {
		t.Fatalf("log file body = %q, want module level prefix", body)
	}
}

func TestNewWritesJSONFileWithModuleField(t *testing.T) {
	dir := t.TempDir()
	cfg := DefaultConfig()
	cfg.Module = "api"
	cfg.Format = "JSON"
	cfg.Output.Console.Enabled = false
	cfg.Output.File.Enabled = true
	cfg.Output.File.Path = dir

	log, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	log.Warn("json output works")
	_ = log.Sync()

	body := readLogFile(t, dir, "api")
	if !strings.Contains(body, `"module":"api"`) {
		t.Fatalf("log file body = %q, want module field", body)
	}
	if !strings.Contains(body, `"level":"WARN"`) {
		t.Fatalf("log file body = %q, want json level", body)
	}
}

func TestNewValidatesConfig(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want error
	}{
		{
			name: "invalid level",
			cfg: func() Config {
				cfg := DefaultConfig()
				cfg.Level = "trace"
				return cfg
			}(),
			want: ErrInvalidLevel,
		},
		{
			name: "invalid format",
			cfg: func() Config {
				cfg := DefaultConfig()
				cfg.Format = "text"
				return cfg
			}(),
			want: ErrInvalidFormat,
		},
		{
			name: "file enabled without path",
			cfg: func() Config {
				cfg := DefaultConfig()
				cfg.Output.File.Enabled = true
				cfg.Output.File.Path = ""
				return cfg
			}(),
			want: ErrFilePathRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.cfg)
			if err == nil {
				t.Fatal("New() error = nil, want error")
			}
			if !errors.Is(err, tt.want) {
				t.Fatalf("New() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestNewAppliesZeroValueDefaults(t *testing.T) {
	cfg := Config{}

	log, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if log == nil {
		t.Fatal("New() returned nil logger")
	}
}

func readLogFile(t *testing.T, dir string, module string) string {
	t.Helper()

	path := filepath.Join(dir, module, module+".log")
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read log file %s: %v", path, err)
	}

	return string(body)
}
