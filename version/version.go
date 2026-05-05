package version

import (
	"bytes"
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
	"text/template"
)

// Build information. Populated at build-time.
var (
	Revision  string
	Branch    string
	GoVersion = runtime.Version()
	GoOS      = runtime.GOOS
	GoArch    = runtime.GOARCH

	computedRevision string
	computedTags     string
)

// printTmpl is the pre-compiled template used by Print.
var printTmpl = template.Must(template.New("version").Parse(`{{.program}}, (branch: {{.branch}}, revision: {{.revision}})
go version:	{{.goVersion}}
platform:	{{.platform}}
tags:	{{.tags}}
`))

// Print returns version information formatted with the given program name.
func Print(program string) string {
	m := map[string]string{
		"program":   program,
		"revision":  GetRevision(),
		"branch":    Branch,
		"goVersion": GoVersion,
		"platform":  GoOS + "/" + GoArch,
		"tags":      GetTags(),
	}

	var buf bytes.Buffer
	if err := printTmpl.Execute(&buf, m); err != nil {
		panic(err)
	}
	return strings.TrimSpace(buf.String())
}

// Info returns version, branch and revision information.
func Info() string {
	return fmt.Sprintf("(branch=%s, revision=%s)", Branch, GetRevision())
}

// BuildContext returns goVersion, platform, buildUser and buildDate information.
func BuildContext() string {
	return fmt.Sprintf("(go=%s, platform=%s, tags=%s)", GoVersion, GoOS+"/"+GoArch, GetTags())
}

// Slog returns a slice of key-value pairs for use with structured logging.
//
// Example:
//
//	logger.Info("Starting server", version.Slog()...)
func Slog() []any {
	return []any{
		"revision", GetRevision(),
		"branch", Branch,
		"goversion", GoVersion,
		"goos", GoOS,
		"goarch", GoArch,
	}
}

// GetRevision returns the revision string, preferring the injected Revision
// variable over the auto-detected value from debug.ReadBuildInfo().
func GetRevision() string {
	if Revision != "" {
		return Revision
	}
	return computedRevision
}

// GetTags returns the build tags used during compilation.
func GetTags() string {
	return computedTags
}

func init() {
	computedRevision, computedTags = computeRevision()
}

func computeRevision() (string, string) {
	var (
		rev      = "unknown"
		tags     = "unknown"
		modified bool
	)

	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return rev, tags
	}
	for _, v := range buildInfo.Settings {
		if v.Key == "vcs.revision" {
			rev = v.Value
		}
		if v.Key == "vcs.modified" {
			if v.Value == "true" {
				modified = true
			}
		}
		if v.Key == "-tags" {
			tags = v.Value
		}
	}
	if modified {
		return rev + "-modified", tags
	}
	return rev, tags
}
