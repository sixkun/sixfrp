package frpcclient

import (
	"cnb.cool/sixkun/sixfrp/v2/haokun/version"

	pb "cnb.cool/sixkun/sixfrp-client-proto/sixfrp/client_pb"
)

// clientVersion maps the pb-free version.Info into the generated message this
// agent puts on the wire.
func clientVersion() *pb.ClientVersion {
	v := version.ClientVersion()
	return &pb.ClientVersion{
		GitVersion:  v.GitVersion,
		BuildDate:   v.BuildDate,
		GoVersion:   v.GoVersion,
		Compiler:    v.Compiler,
		Platform:    v.Platform,
		DockerImage: v.DockerImage,
	}
}
