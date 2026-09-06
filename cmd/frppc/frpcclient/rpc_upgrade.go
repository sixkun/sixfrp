package frpcclient

import (
	"fmt"
	"runtime"
	"strings"

	"haokun-panel/cmd/frppc/facades"
	"haokun-panel/haokun/version"
	"haokun-panel/utils/selfupgrade"

	pb "cnb.cool/sixkun/sixfrp-client-proto/sixfrp/client_pb"
)

func (s *Service) handleUpgradeFrppWrapped(req *pb.UpgradeFrppRequest) (*pb.UpgradeFrppResponse, error) {
	upgradeType := selfupgrade.UpgradeType(req.GetUpgradeType())

	// The exec/respawn strategies replace the running binary in place, which is
	// not possible under Windows or a Docker image. TypeCommand delegates to an
	// external command, so it is exempt from these guards.
	if upgradeType != selfupgrade.UpgradeTypeCommand {
		if runtime.GOOS == "windows" {
			return nil, fmt.Errorf("frppc in-place upgrade is not supported on Windows")
		}
		if version.ClientVersion().DockerImage {
			return nil, fmt.Errorf("frppc in-place upgrade is not supported in Docker; update the image instead")
		}
	}

	opt := selfupgrade.Options{
		DownloadURL: req.GetDownloadUrl(),
		TargetPath:  req.GetTargetPath(),
		Backup:      req.GetBackup(),
		Type:        upgradeType,
	}
	if upgradeType == selfupgrade.UpgradeTypeCommand {
		opt.Command = strings.TrimSpace(facades.Config().GetString("frppc.upgrade_command"))
	}

	if err := selfupgrade.Upgrade(opt); err != nil {
		return nil, err
	}

	return &pb.UpgradeFrppResponse{
		Status: &pb.Status{Code: pb.RespCode_RESP_CODE_SUCCESS, Message: upgradeMessage(upgradeType)},
	}, nil
}

func upgradeMessage(t selfupgrade.UpgradeType) string {
	switch t {
	case selfupgrade.UpgradeTypeReplaceExec:
		return "upgrade installed; frppc will re-exec into the new binary shortly"
	case selfupgrade.UpgradeTypeReplaceRespawn:
		return "upgrade installed; frppc will respawn into the new binary shortly"
	case selfupgrade.UpgradeTypeCommand:
		return "upgrade command launched"
	default:
		return "upgrade installed; restart frppc to use it"
	}
}
