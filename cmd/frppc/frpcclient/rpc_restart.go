package frpcclient

import (
	"fmt"
	"strings"

	"cnb.cool/sixkun/sixfrp/v2/cmd/frppc/facades"
	"cnb.cool/sixkun/sixfrp/v2/utils/selfupgrade"

	pb "cnb.cool/sixkun/sixfrp-client-proto/sixfrp/client_pb"
)

// handleRestartFrppWrapped restarts frppc's own service by launching the
// configured restart command (FRPPC_RESTART_COMMAND) detached, mirroring how
// the master restarts itself. The command owns the actual service restart, so
// this returns as soon as it has been launched — the RPC response is sent
// before the process is torn down.
func (s *Service) handleRestartFrppWrapped(_ *pb.RestartFrppRequest) (*pb.RestartFrppResponse, error) {
	command := strings.TrimSpace(facades.Config().GetString("frppc.restart_command"))
	if command == "" {
		return nil, fmt.Errorf("frppc restart command is not configured (FRPPC_RESTART_COMMAND)")
	}
	if err := selfupgrade.RunDetachedCommand(command); err != nil {
		return nil, fmt.Errorf("launch frppc restart command: %w", err)
	}
	return &pb.RestartFrppResponse{
		Status: &pb.Status{Code: pb.RespCode_RESP_CODE_SUCCESS, Message: "restart command launched"},
	}, nil
}
