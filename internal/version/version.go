package version

import (
	"fmt"
	"runtime"
)

const (
	// programName is the name of this program.
	programName string = "katok"
)

var (
	// Git variables imported at build stage.
	gitTag, gitCommit, gitBranch string
)

type VersionInfo struct {
	GitTag    string
	GitCommit string
	GitBranch string
	OS        string
	Platform  string
}

func Get() *VersionInfo {
	// These variables typically come from -ldflags settings and in
	// their absence fallback to the constants above
	return &VersionInfo{
		GitTag:    gitTag,
		GitCommit: gitCommit,
		GitBranch: gitBranch,
		OS:        fmt.Sprintf("%s", runtime.GOOS),
		Platform:  fmt.Sprintf("%s", runtime.GOARCH),
	}
}
