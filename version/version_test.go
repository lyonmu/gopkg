package version

import (
	"strings"
	"testing"
)

func resetVersionState() {
	Version = ""
	Branch = ""
	Revision = ""
	BuildUser = ""
	BuildDate = ""
	GoVersion = "go1.24.0"
	GoOS = "linux"
	GoArch = "amd64"
	computedRevision = ""
	computedTags = ""
}

func TestInfo(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		branch   string
		revision string
		want     string
	}{
		{
			name:     "empty values",
			version:  "",
			branch:   "",
			revision: "",
			want:     "(version=, branch=, revision=)",
		},
		{
			name:     "with values",
			version:  "1.0.0",
			branch:   "main",
			revision: "abc123",
			want:     "(version=1.0.0, branch=main, revision=abc123)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(resetVersionState)
			Version = tt.version
			Branch = tt.branch
			Revision = tt.revision
			computedRevision = "" // ensure empty revision case is truly empty
			got := Info()
			if got != tt.want {
				t.Errorf("Info() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildContext(t *testing.T) {
	tests := []struct {
		name      string
		buildUser string
		buildDate string
	}{
		{
			name:      "empty values",
			buildUser: "",
			buildDate: "",
		},
		{
			name:      "with values",
			buildUser: "testuser",
			buildDate: "2026-05-05",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(resetVersionState)
			BuildUser = tt.buildUser
			BuildDate = tt.buildDate
			got := BuildContext()
			// 仅验证格式，不验证具体 runtime 值
			if !strings.HasPrefix(got, "(go=") || !strings.HasSuffix(got, ")") {
				t.Errorf("BuildContext() unexpected format: %q", got)
			}
			if tt.buildUser != "" {
				if !strings.Contains(got, "testuser") {
					t.Errorf("BuildContext() missing build user; got: %q", got)
				}
				if !strings.Contains(got, "2026-05-05") {
					t.Errorf("BuildContext() missing build date; got: %q", got)
				}
			}
		})
	}
}

func TestGetRevision(t *testing.T) {
	tests := []struct {
		name     string
		revision string
		computed string
		want     string
	}{
		{
			name:     "Revision set",
			revision: "v1.0.0",
			computed: "abc123",
			want:     "v1.0.0",
		},
		{
			name:     "Revision empty, use computed",
			revision: "",
			computed: "abc123",
			want:     "abc123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(resetVersionState)
			Revision = tt.revision
			computedRevision = tt.computed
			got := GetRevision()
			if got != tt.want {
				t.Errorf("GetRevision() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetTags(t *testing.T) {
	tests := []struct {
		name     string
		computed string
		want     string
	}{
		{
			name:     "with tags",
			computed: "netgo,osusergo",
			want:     "netgo,osusergo",
		},
		{
			name:     "empty tags",
			computed: "",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(resetVersionState)
			computedTags = tt.computed
			got := GetTags()
			if got != tt.want {
				t.Errorf("GetTags() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPrint(t *testing.T) {
	t.Cleanup(resetVersionState)
	// 设置已知值
	Version = "2.0.0"
	Branch = "develop"
	Revision = "def456"
	BuildUser = "builder"
	BuildDate = "2026-01-01"
	computedRevision = "computed-rev"

	got := Print("myapp")

	// 验证关键字段都出现在输出中
	for _, want := range []string{
		"myapp",
		"2.0.0",
		"develop",
		"def456", // Revision 非空，GetRevision 返回 Revision
		"builder",
		"2026-01-01",
		"go version:",
		"platform:",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("Print() output missing %q; got:\n%s", want, got)
		}
	}
}

func TestSlog(t *testing.T) {
	t.Cleanup(resetVersionState)
	Version = "1.2.3"
	Branch = "main"
	Revision = "sha123"
	BuildUser = "user"
	BuildDate = "today"
	GoVersion = "go1.24.0"
	GoOS = "linux"
	GoArch = "amd64"

	got := Slog()

	// Slog 返回 []any，长度应为 16 (8 对 key-value)
	if len(got) != 16 {
		t.Fatalf("Slog() returned %d elements, want 16", len(got))
	}

	// 验证 key 的顺序和内容
	wantKeys := []string{
		"version", "revision", "branch", "builduser",
		"builddate", "goversion", "goos", "goarch",
	}
	for i, want := range wantKeys {
		key := got[i*2]
		if key != want {
			t.Errorf("Slog()[%d] = %v, want key %q", i*2, key, want)
		}
	}

	// 验证 value 的顺序和内容
	wantValues := []any{
		"1.2.3", "sha123", "main", "user",
		"today", "go1.24.0", "linux", "amd64",
	}
	for i, want := range wantValues {
		val := got[i*2+1]
		if val != want {
			t.Errorf("Slog()[%d] = %v, want value %v", i*2+1, val, want)
		}
	}
}
