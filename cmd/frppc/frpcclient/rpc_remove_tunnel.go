package frpcclient

import (
	"os"
	"time"

	"haokun-panel/cmd/frppc/facades"

	pb "cnb.cool/sixkun/sixfrp-client-proto/sixfrp/client_pb"
)

func (s *Service) handleRemoveFrpcWrapped(req *pb.RemoveFRPShadowClientRequest) (*pb.RemoveFRPShadowClientResponse, error) {
	facades.Log().Infof("remove frpc requested, will exit in 10s, req=%+v", req)

	go func() {
		time.Sleep(10 * time.Second)
		os.Exit(0)
	}()

	return &pb.RemoveFRPShadowClientResponse{
		Status: &pb.Status{Code: pb.RespCode_RESP_CODE_SUCCESS, Message: "ok"},
	}, nil
}
