package version

import (
	"runtime"
	"strconv"
)

// Info reports enough build/runtime metadata for the master to select a
// compatible release artifact and decide whether in-place upgrade is safe.
//
// It is a plain struct rather than a generated message because both agents read
// it and they no longer share one: frpps maps it into protos/pb.ClientVersion and
// frppc into the published client contract's, each in its own version_msg.go.
type Info struct {
	GitVersion  string
	BuildDate   string
	GoVersion   string
	Compiler    string
	Platform    string
	DockerImage bool
}

// ClientVersion returns this binary's build metadata.
//
// arch is the release arch token, which must equal the arch stored for this
// build's release row (see sync_releases.ParseAssetFilename). It is injected at
// build time (ReleaseArch) and, for 32-bit ARM, carries the variant (arm for
// ARMv5/v6, armv7 for ARMv7) that runtime.GOARCH cannot distinguish. Unstamped
// local builds fall back to runtime.GOARCH, which is fine because they are not
// matched against release artifacts.
func ClientVersion() Info {
	arch := ReleaseArch
	if arch == "" {
		arch = runtime.GOARCH
	}
	isDocker, _ := strconv.ParseBool(DockerImage)
	return Info{
		GitVersion:  Version,
		BuildDate:   BuildTime,
		GoVersion:   runtime.Version(),
		Compiler:    runtime.Compiler,
		Platform:    runtime.GOOS + "/" + arch,
		DockerImage: isDocker,
	}
}
