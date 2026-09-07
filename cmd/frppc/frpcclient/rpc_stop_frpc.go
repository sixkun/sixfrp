package frpcclient

import (
	"cnb.cool/sixkun/sixfrp/v2/cmd/frppc/facades"

	pb "cnb.cool/sixkun/sixfrp-client-proto/sixfrp/client_pb"
)

func (s *Service) handleStopFRPCWrapped(req *pb.StopFRPShadowClientRequest) (*pb.StopFRPShadowClientResponse, error) {
	facades.Log().Infof("stop frpc requested, req=%+v", req)

	s.controller.StopAll()
	s.controller.DeleteAll()

	return &pb.StopFRPShadowClientResponse{
		Status: &pb.Status{Code: pb.RespCode_RESP_CODE_SUCCESS, Message: "ok"},
	}, nil
}
