package version

import "time"

// Version and BuildTime are injected at build time via
//
//	-ldflags "-X cnb.cool/sixkun/sixfrp/v2/haokun/version.Version=... \
//	          -X cnb.cool/sixkun/sixfrp/v2/haokun/version.BuildTime=..."
//
// (see scripts/build/builder.ts and Dockerfile.frppc). The path must be this
// package's real import path: -X against a package outside the build graph is
// silently ignored, leaving these at their defaults with no error. They default
// to dev values for un-stamped local builds.
var (
	Version     = "dev"
	BuildTime   = ""
	ReleaseArch = ""
	DockerImage = "false"
)

// StartedAt is the process start time, used for uptime reporting.
var StartedAt = time.Now()
