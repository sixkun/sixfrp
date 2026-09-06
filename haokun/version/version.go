package version

import "time"

// Version and BuildTime are injected at build time via
//
//	-ldflags "-X haokun-panel/haokun/version.Version=... -X haokun-panel/haokun/version.BuildTime=..."
//
// (see scripts/build/builder.ts). They default to dev values for un-stamped
// local builds.
var (
	Version     = "dev"
	BuildTime   = ""
	ReleaseArch = ""
	DockerImage = "false"
)

// StartedAt is the process start time, used for uptime reporting.
var StartedAt = time.Now()
