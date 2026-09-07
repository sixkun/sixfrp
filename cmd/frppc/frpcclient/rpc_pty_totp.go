package frpcclient

import (
	"fmt"

	"cnb.cool/sixkun/sixfrp/v2/cmd/frppc/facades"

	pb "cnb.cool/sixkun/sixfrp-client-proto/sixfrp/client_pb"
)

func (s *Service) handlePTYTOTPBindWrapped(req *pb.PTYTOTPRequest) (*pb.CommonResponse, error) {
	if s.opts.totpSecret == "" {
		return nil, fmt.Errorf("TOTP is not configured on this client")
	}
	if !ValidateTOTP(s.opts.totpSecret, req.GetTotpCode()) {
		facades.Log().Warning("invalid TOTP code during bind attempt")
		return nil, fmt.Errorf("invalid TOTP code")
	}

	MarkTOTPBound(s.opts.clientID)
	facades.Log().Infof("TOTP bound for client %s", s.opts.clientID)

	return s.openPTYAndRespond()
}

func (s *Service) handlePTYTOTPVerifyWrapped(req *pb.PTYTOTPRequest) (*pb.CommonResponse, error) {
	if s.opts.totpSecret == "" {
		return nil, fmt.Errorf("TOTP is not configured on this client")
	}
	if !ValidateTOTP(s.opts.totpSecret, req.GetTotpCode()) {
		facades.Log().Warning("invalid TOTP code during verify attempt")
		return nil, fmt.Errorf("invalid TOTP code")
	}

	facades.Log().Infof("TOTP verified for client %s", s.opts.clientID)
	return s.openPTYAndRespond()
}
