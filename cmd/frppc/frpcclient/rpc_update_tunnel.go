package frpcclient

import (
	"reflect"

	"haokun-panel/cmd/frppc/facades"
	"haokun-panel/utils"

	pb "cnb.cool/sixkun/sixfrp-client-proto/sixfrp/client_pb"
)

func (s *Service) handleUpdateFrpcWrapped(req *pb.UpdateFRPShadowClientRequest) (*pb.UpdateFRPShadowClientResponse, error) {
	facades.Log().Infof("update frpc, clientID=%s serverID=%s", req.GetClientId(), req.GetServerId())

	commonCfg, proxyCfgs, visitorCfgs, err := utils.LoadClientConfig(req.GetConfig(), false)
	if err != nil {
		facades.Log().Warningf("load client config failed: %v", err)
		return &pb.UpdateFRPShadowClientResponse{
			Status: &pb.Status{Code: pb.RespCode_RESP_CODE_INVALID, Message: err.Error()},
		}, err
	}

	clientID := req.GetClientId()
	serverID := req.GetServerId()

	existing := s.controller.Get(clientID, serverID)
	if existing == nil {
		if err := s.recreateHandler(clientID, serverID, commonCfg, proxyCfgs, visitorCfgs); err != nil {
			return &pb.UpdateFRPShadowClientResponse{
				Status: &pb.Status{Code: pb.RespCode_RESP_CODE_INVALID, Message: err.Error()},
			}, nil
		}
		facades.Log().Infof("add new client, clientID=%s serverID=%s", clientID, serverID)
		return &pb.UpdateFRPShadowClientResponse{
			Status: &pb.Status{Code: pb.RespCode_RESP_CODE_SUCCESS, Message: "ok"},
		}, nil
	}

	if reflect.DeepEqual(existing.GetCommonCfg(), commonCfg) {
		facades.Log().Infof("client %s common config unchanged, updating proxies", clientID)
		existing.Update(proxyCfgs, visitorCfgs)
		return &pb.UpdateFRPShadowClientResponse{
			Status: &pb.Status{Code: pb.RespCode_RESP_CODE_SUCCESS, Message: "ok"},
		}, nil
	}

	if err := s.recreateHandler(clientID, serverID, commonCfg, proxyCfgs, visitorCfgs); err != nil {
		return &pb.UpdateFRPShadowClientResponse{
			Status: &pb.Status{Code: pb.RespCode_RESP_CODE_INVALID, Message: err.Error()},
		}, nil
	}
	facades.Log().Infof("update client, clientID=%s serverID=%s success, running", clientID, serverID)
	return &pb.UpdateFRPShadowClientResponse{
		Status: &pb.Status{Code: pb.RespCode_RESP_CODE_SUCCESS, Message: "ok"},
	}, nil
}
