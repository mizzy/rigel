package version

import (
	"fmt"
	"runtime"
)

var (
	Version   = "0.1.0"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

func String() string {
	return fmt.Sprintf("rigel version %s (commit: %s, built: %s, %s/%s)",
		Version,
		GitCommit,
		BuildDate,
		runtime.GOOS,
		runtime.GOARCH,
	)
}

func Short() string {
	return Version
}
