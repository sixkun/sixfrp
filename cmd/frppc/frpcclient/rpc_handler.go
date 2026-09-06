package frpcclient

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"

	"haokun-panel/cmd/frppc/facades"
	"haokun-panel/haokun/defs"

	pb "cnb.cool/sixkun/sixfrp-client-proto/sixfrp/client_pb"
)

func (s *Service) ConnectAndServeRPC(ctx context.Context) error {
	stream, err := s.masterClient.ServerSend(ctx)
	if err != nil {
		return fmt.Errorf("connect rpc stream: %w", err)
	}

	if err := s.registerToMaster(stream); err != nil {
		return fmt.Errorf("register to master: %w", err)
	}

	return s.consumeServerStream(ctx, stream)
}

func (s *Service) registerToMaster(stream pb.ClientMaster_ServerSendClient) error {
	if err := stream.Send(&pb.ClientMessage{
		Event:     pb.ClientEvent_CLIENT_EVENT_REGISTER,
		ClientId:  s.opts.clientID,
		SessionId: uuid.New().String(),
		Secret:    s.opts.clientSecret,
	}); err != nil {
		return err
	}

	resp, err := stream.Recv()
	if err != nil {
		return err
	}

	if resp.GetEvent() == pb.ClientEvent_CLIENT_EVENT_REGISTER {
		facades.Log().Infof("client registered to master successfully, id=%s", s.opts.clientID)
		s.emitLifecycleEvent(defs.SystemEventTypeFrppcStarted, defs.SystemEventLevelInfo)
		return nil
	}
	if resp.GetEvent() == pb.ClientEvent_CLIENT_EVENT_ERROR {
		return fmt.Errorf("register rejected: %s", strings.TrimSpace(string(resp.GetData())))
	}
	return fmt.Errorf("unexpected register response event: %s", resp.GetEvent().String())
}

func (s *Service) consumeServerStream(ctx context.Context, stream pb.ClientMaster_ServerSendClient) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		req, err := stream.Recv()
		if err != nil {
			return err
		}
		if req == nil {
			continue
		}

		resp := s.handleServerMessage(req)
		if resp == nil {
			continue
		}
		resp.ClientId = s.opts.clientID
		resp.SessionId = req.GetSessionId()
		if err := stream.Send(resp); err != nil {
			return err
		}
	}
}

func (s *Service) handleServerMessage(req *pb.ServerMessage) (resp *pb.ClientMessage) {
	defer func() {
		if r := recover(); r != nil {
			facades.Log().Errorf("panic in handleServerMessage: %v", r)
			resp = rpcErrorMessage(fmt.Errorf("internal panic: %v", r))
		}
	}()

	switch req.GetEvent() {
	case pb.ClientEvent_CLIENT_EVENT_UPDATE_FRPC:
		return wrapServerMsg(s, req, (*Service).handleUpdateFrpcWrapped)
	case pb.ClientEvent_CLIENT_EVENT_REMOVE_FRPC:
		return wrapServerMsg(s, req, (*Service).handleRemoveFrpcWrapped)
	case pb.ClientEvent_CLIENT_EVENT_START_FRPC:
		return wrapServerMsg(s, req, (*Service).handleStartFRPCWrapped)
	case pb.ClientEvent_CLIENT_EVENT_STOP_FRPC:
		return wrapServerMsg(s, req, (*Service).handleStopFRPCWrapped)
	case pb.ClientEvent_CLIENT_EVENT_GET_PROXY_INFO:
		return wrapServerMsg(s, req, (*Service).handleGetProxyConfigWrapped)
	case pb.ClientEvent_CLIENT_EVENT_START_STREAM_LOG:
		return wrapServerMsg(s, req, (*Service).handleStartStreamLogWrapped)
	case pb.ClientEvent_CLIENT_EVENT_STOP_STREAM_LOG:
		return wrapServerMsg(s, req, (*Service).handleStopStreamLogWrapped)
	case pb.ClientEvent_CLIENT_EVENT_START_PTY_CONNECT:
		return wrapServerMsg(s, req, (*Service).handleStartPTYConnectWrapped)
	case pb.ClientEvent_CLIENT_EVENT_PTY_TOTP_BIND:
		return wrapServerMsg(s, req, (*Service).handlePTYTOTPBindWrapped)
	case pb.ClientEvent_CLIENT_EVENT_PTY_TOTP_VERIFY:
		return wrapServerMsg(s, req, (*Service).handlePTYTOTPVerifyWrapped)
	case pb.ClientEvent_CLIENT_EVENT_UPGRADE_AGENT:
		return wrapServerMsg(s, req, (*Service).handleUpgradeFrppWrapped)
	case pb.ClientEvent_CLIENT_EVENT_RESTART_AGENT:
		return wrapServerMsg(s, req, (*Service).handleRestartFrppWrapped)
	case pb.ClientEvent_CLIENT_EVENT_PING:
		versionRaw, _ := proto.Marshal(clientVersion())
		return &pb.ClientMessage{Event: pb.ClientEvent_CLIENT_EVENT_PONG, Data: versionRaw}
	default:
		return rpcErrorMessage(fmt.Errorf("unsupported event: %s", req.GetEvent().String()))
	}
}
