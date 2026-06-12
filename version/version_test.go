package version

import (
	"strings"
	"testing"
)

func resetVersionState() {
	Branch = ""
	Revision = ""
	GoVersion = "go1.24.0"
	GoOS = "linux"
	GoArch = "amd64"
	computedRevision = ""
	computedTags = ""
}

func TestInfo(t *testing.T) {
	tests := []struct {
		name     string
		branch   string
		revision string
		want     string
	}{
		{
			name:     "empty values",
			branch:   "",
			revision: "",
			want:     "(branch=, revision=)",
		},
		{
			name:     "with values",
			branch:   "main",
			revision: "abc123",
			want:     "(branch=main, revision=abc123)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(resetVersionState)
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
		name string
		tags string
	}{
		{
			name: "empty tags",
			tags: "",
		},
		{
			name: "with tags",
			tags: "netgo,osusergo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(resetVersionState)
			computedTags = tt.tags
			got := BuildContext()
			want := "(go=go1.24.0, platform=linux/amd64, tags=" + tt.tags + ")"
			if got != want {
				t.Errorf("BuildContext() = %q, want %q", got, want)
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
	Branch = "develop"
	Revision = "def456"
	computedRevision = "computed-rev"

	got := Print("myapp")

	// 验证关键字段都出现在输出中
	for _, want := range []string{
		"myapp",
		"develop",
		"def456", // Revision 非空，GetRevision 返回 Revision
		"go version:",
		"platform:",
		"tags:",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("Print() output missing %q; got:\n%s", want, got)
		}
	}
}

func TestSlog(t *testing.T) {
	t.Cleanup(resetVersionState)
	Branch = "main"
	Revision = "sha123"
	GoVersion = "go1.24.0"
	GoOS = "linux"
	GoArch = "amd64"

	got := Slog()

	// Slog 返回 []any，长度应为 10 (5 对 key-value)
	if len(got) != 10 {
		t.Fatalf("Slog() returned %d elements, want 10", len(got))
	}

	// 验证 key 的顺序和内容
	wantKeys := []string{
		"revision", "branch", "goversion", "goos", "goarch",
	}
	for i, want := range wantKeys {
		key := got[i*2]
		if key != want {
			t.Errorf("Slog()[%d] = %v, want key %q", i*2, key, want)
		}
	}

	// 验证 value 的顺序和内容
	wantValues := []any{
		"sha123", "main", "go1.24.0", "linux", "amd64",
	}
	for i, want := range wantValues {
		val := got[i*2+1]
		if val != want {
			t.Errorf("Slog()[%d] = %v, want value %v", i*2+1, val, want)
		}
	}
}
