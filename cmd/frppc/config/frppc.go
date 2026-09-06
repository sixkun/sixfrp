package config

import (
	"haokun-panel/cmd/frppc/facades"
)

func loadFrppc() {
	config := facades.Config()
	config.Add("frppc", map[string]any{
		"enabled":               config.Env("FRPPC_ENABLED", true),
		"id":                    config.Env("FRPPC_ID", ""),
		"secret":                config.Env("FRPPC_SECRET", ""),
		"rpc_url":               config.Env("FRPPC_RPC_URL", ""),
		"sync_interval_seconds": config.Env("FRPPC_SYNC_INTERVAL_SECONDS", 300),
		// External upgrade command for FRPP_UPGRADE_TYPE_SCRIPT. frppc runs it
		// detached with {{url}} replaced by the resolved download URL, e.g.
		//   FRPPC_UPGRADE_COMMAND="/opt/haokun-frppc/upgrade-frppc.sh {{url}}"
		"upgrade_command": config.Env("FRPPC_UPGRADE_COMMAND", ""),
		// Restart command for the EVENT_RESTART_FRPP event. frppc runs it detached
		// and the command owns restarting the frppc service, e.g.
		//   FRPPC_RESTART_COMMAND="systemctl restart haokun-frppc"
		"restart_command": config.Env("FRPPC_RESTART_COMMAND", ""),
		"features": map[string]any{
			"enable_remote_shell": config.Env("FRPPC_FEATURES_ENABLE_REMOTE_SHELL", true),
			"totp_secret":         config.Env("FRPPC_FEATURES_TOTP_SECRET", ""),
		},
	})
}
