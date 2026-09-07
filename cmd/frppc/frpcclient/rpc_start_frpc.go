package frpcclient

import (
	"cnb.cool/sixkun/sixfrp/v2/cmd/frppc/facades"

	pb "cnb.cool/sixkun/sixfrp-client-proto/sixfrp/client_pb"
)

func (s *Service) handleStartFRPCWrapped(req *pb.StartFRPShadowClientRequest) (*pb.StartFRPShadowClientResponse, error) {
	facades.Log().Infof("start frpc requested, clientID=%s", req.GetClientId())

	clientID := req.GetClientId()
	if clientID == "" {
		clientID = s.opts.clientID
	}

	if err := s.pullConfigForID(s.Context(), clientID); err != nil {
		return &pb.StartFRPShadowClientResponse{
			Status: &pb.Status{Code: pb.RespCode_RESP_CODE_INVALID, Message: err.Error()},
		}, err
	}

	return &pb.StartFRPShadowClientResponse{
		Status: &pb.Status{Code: pb.RespCode_RESP_CODE_SUCCESS, Message: "ok"},
	}, nil
}
