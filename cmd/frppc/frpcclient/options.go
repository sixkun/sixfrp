package frpcclient

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"haokun-panel/cmd/frppc/facades"
)

func resolveOptionsFromConfig() options {
	cfg := facades.Config()
	s := options{}

	s.enabled = cfg.GetBool("frppc.enabled", true)
	s.clientID = strings.TrimSpace(cfg.GetString("frppc.id"))
	s.clientSecret = strings.TrimSpace(cfg.GetString("frppc.secret"))
	s.rpcURL = strings.TrimSpace(cfg.GetString("frppc.rpc_url"))

	syncIntervalSeconds := cfg.GetInt("frppc.sync_interval_seconds", 30)
	if syncIntervalSeconds <= 0 {
		syncIntervalSeconds = 30
	}
	s.syncInterval = time.Duration(syncIntervalSeconds) * time.Second

	s.enableRemoteShell = cfg.GetBool("frppc.features.enable_remote_shell", true)
	s.totpSecret = strings.TrimSpace(cfg.GetString("frppc.features.totp_secret"))

	if s.clientID == "" {
		s.clientID = firstNonEmptyEnv("FRPPC_ID", "CLIENT_ID")
	}
	if s.clientSecret == "" {
		s.clientSecret = firstNonEmptyEnv("FRPPC_SECRET", "CLIENT_SECRET")
	}
	if s.rpcURL == "" {
		s.rpcURL = firstNonEmptyEnv("FRPPC_RPC_URL", "FRP_RPC_URL")
		if s.rpcURL != "" && !strings.Contains(s.rpcURL, "://") {
			if inferredWSURL, ok := inferWebSocketRPCURL(s.rpcURL); ok {
				s.rpcURL = inferredWSURL
			} else {
				s.rpcURL = "grpc://" + s.rpcURL
			}
		}
	}
	if s.rpcURL == "" {
		s.rpcURL = defaultRPCURL
	}

	return s
}

func validateRequiredOptions(s options) error {
	if s.clientID == "" {
		return fmt.Errorf("missing client id, set frppc.id or FRPPC_ID")
	}
	if s.clientSecret == "" {
		return fmt.Errorf("missing client secret, set frppc.secret or FRPPC_SECRET")
	}
	return nil
}

func firstNonEmptyEnv(keys ...string) string {
	for _, key := range keys {
		if v := strings.TrimSpace(os.Getenv(key)); v != "" {
			return v
		}
	}
	return ""
}

func envInt(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}
