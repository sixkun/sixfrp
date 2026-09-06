package frpcclient

import (
	"sync"

	"github.com/pquerna/otp/totp"
)

var (
	totpBound   = make(map[string]bool)
	totpBoundMu sync.RWMutex
)

func IsTOTPBound(clientID string) bool {
	totpBoundMu.RLock()
	defer totpBoundMu.RUnlock()
	return totpBound[clientID]
}

func MarkTOTPBound(clientID string) {
	totpBoundMu.Lock()
	defer totpBoundMu.Unlock()
	totpBound[clientID] = true
}

func ValidateTOTP(secret, code string) bool {
	return totp.Validate(code, secret)
}
