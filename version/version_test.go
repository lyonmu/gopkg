package version

import (
	"runtime/debug"
	"strings"
	"testing"
)

func restoreVersionState(t *testing.T) {
	t.Helper()

	originalBranch := Branch
	originalCommit := Commit
	originalGoVersion := GoVersion
	originalGoOS := GoOS
	originalGoArch := GoArch
	originalComputedRevision := computedRevision
	originalComputedTags := computedTags

	t.Cleanup(func() {
		Branch = originalBranch
		Commit = originalCommit
		GoVersion = originalGoVersion
		GoOS = originalGoOS
		GoArch = originalGoArch
		computedRevision = originalComputedRevision
		computedTags = originalComputedTags
	})
}

func TestInfo(t *testing.T) {
	tests := []struct {
		name     string
		branch   string
		commit string
		want     string
	}{
		{
			name:     "empty values",
			branch:   "",
			commit: "",
			want:     "(branch=, commit=)",
		},
		{
			name:     "with values",
			branch:   "main",
			commit: "abc123",
			want:     "(branch=main, commit=abc123)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			restoreVersionState(t)
			Branch = tt.branch
			Commit = tt.commit
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
			restoreVersionState(t)
			GoVersion = "go1.24.0"
			GoOS = "linux"
			GoArch = "amd64"
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
		commit string
		computed string
		want     string
	}{
		{
			name:     "Commit set",
			commit: "v1.0.0",
			computed: "abc123",
			want:     "v1.0.0",
		},
		{
			name:     "Commit empty, use computed",
			commit: "",
			computed: "abc123",
			want:     "abc123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			restoreVersionState(t)
			Commit = tt.commit
			computedRevision = tt.computed
			got := GetCommit()
			if got != tt.want {
				t.Errorf("GetCommit() = %q, want %q", got, tt.want)
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
			restoreVersionState(t)
			computedTags = tt.computed
			got := GetTags()
			if got != tt.want {
				t.Errorf("GetTags() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestComputeRevisionFrom(t *testing.T) {
	tests := []struct {
		name         string
		buildInfo    *debug.BuildInfo
		ok           bool
		wantRevision string
		wantTags     string
	}{
		{
			name:         "unavailable build info",
			ok:           false,
			wantRevision: "unknown",
			wantTags:     "unknown",
		},
		{
			name: "revision and tags",
			buildInfo: &debug.BuildInfo{
				Settings: []debug.BuildSetting{
					{Key: "vcs.revision", Value: "abc123"},
					{Key: "-tags", Value: "netgo,osusergo"},
				},
			},
			ok:           true,
			wantRevision: "abc123",
			wantTags:     "netgo,osusergo",
		},
		{
			name: "modified revision without tags",
			buildInfo: &debug.BuildInfo{
				Settings: []debug.BuildSetting{
					{Key: "vcs.revision", Value: "def456"},
					{Key: "vcs.modified", Value: "true"},
				},
			},
			ok:           true,
			wantRevision: "def456-modified",
			wantTags:     "unknown",
		},
		{
			name:         "nil build info reported available",
			ok:           true,
			wantRevision: "unknown",
			wantTags:     "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRevision, gotTags := computeRevisionFrom(tt.buildInfo, tt.ok)
			if gotRevision != tt.wantRevision {
				t.Errorf("computeRevisionFrom() revision = %q, want %q", gotRevision, tt.wantRevision)
			}
			if gotTags != tt.wantTags {
				t.Errorf("computeRevisionFrom() tags = %q, want %q", gotTags, tt.wantTags)
			}
		})
	}
}

func TestPrint(t *testing.T) {
	restoreVersionState(t)
	Branch = "develop"
	Commit = "def456"
	computedRevision = "computed-rev"

	got := Print("myapp")

	// 验证关键字段都出现在输出中
	for _, want := range []string{
		"myapp",
		"develop",
		"def456", // Commit 非空，GetCommit 返回 Commit
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
	restoreVersionState(t)
	Branch = "main"
	Commit = "sha123"
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
