package version

import (
	"fmt"
	"runtime"
	"time"
)

// Build-time variables set by ldflags
var (
  Version   = "0.0.2"
	GitCommit = "unknown"         // Git commit hash
	BuildTime = "unknown"         // Build timestamp
	GoVersion = runtime.Version() // Go version used
)

// Info contains complete version information
type Info struct {
	Version   string `json:"version"`
	GitCommit string `json:"git_commit"`
	BuildTime string `json:"build_time"`
	GoVersion string `json:"go_version"`
	Platform  string `json:"platform"`
	Arch      string `json:"arch"`
}

// Get returns complete version information
func Get() Info {
	return Info{
		Version:   Version,
		GitCommit: GitCommit,
		BuildTime: BuildTime,
		GoVersion: GoVersion,
		Platform:  runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
}

// GetVersion returns just the version string
func GetVersion() string {
	if Version == "" {
		return "dev"
	}
	return Version
}

// GetBuildTime returns formatted build time
func GetBuildTime() string {
	if BuildTime == "" || BuildTime == "unknown" {
		return "unknown"
	}

	// Try to parse the build time
	if t, err := time.Parse(time.RFC3339, BuildTime); err == nil {
		return t.Format("2006-01-02 15:04:05 UTC")
	}

	return BuildTime
}

// GetShortCommit returns short git commit hash
func GetShortCommit() string {
	if GitCommit == "" || GitCommit == "unknown" {
		return "unknown"
	}

	if len(GitCommit) > 8 {
		return GitCommit[:8]
	}

	return GitCommit
}

// String returns a formatted version string
func (i Info) String() string {
	return fmt.Sprintf("Guardian v%s (%s) built with %s on %s",
		i.Version,
		GetShortCommit(),
		i.GoVersion,
		GetBuildTime())
}

// IsRelease checks if this is a release build (not dev)
func IsRelease() bool {
	return Version != "" && Version != "dev" && GitCommit != "unknown"
}

// IsDevelopment checks if this is a development build
func IsDevelopment() bool {
	return !IsRelease()
}
