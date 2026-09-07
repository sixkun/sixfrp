package frpcclient

import (
	"fmt"

	"github.com/samber/lo"

	"cnb.cool/sixkun/sixfrp/v2/cmd/frppc/facades"

	pb "cnb.cool/sixkun/sixfrp-client-proto/sixfrp/client_pb"
)

func (s *Service) handleGetProxyConfigWrapped(req *pb.GetProxyConfigRequest) (*pb.GetProxyConfigResponse, error) {
	clientID := req.GetClientId()
	serverID := req.GetServerId()
	proxyName := req.GetName()

	cli := s.controller.Get(clientID, serverID)
	if cli == nil {
		facades.Log().Errorf("get proxy info: client not found, clientID=%s serverID=%s", clientID, serverID)
		return nil, fmt.Errorf("cannot get client")
	}

	workingStatus, ok := cli.GetProxyStatus(proxyName)
	if !ok {
		facades.Log().Errorf("get proxy info: proxy status missing, clientID=%s serverID=%s proxy=%s", clientID, serverID, proxyName)
		return nil, fmt.Errorf("cannot get proxy status")
	}

	return &pb.GetProxyConfigResponse{
		Status: &pb.Status{Code: pb.RespCode_RESP_CODE_SUCCESS, Message: "success"},
		WorkingStatus: &pb.ProxyWorkingStatus{
			Name:       lo.ToPtr(workingStatus.Name),
			Type:       lo.ToPtr(workingStatus.Type),
			Status:     lo.ToPtr(workingStatus.Phase),
			Err:        lo.ToPtr(workingStatus.Err),
			RemoteAddr: lo.ToPtr(workingStatus.RemoteAddr),
		},
	}, nil
}
