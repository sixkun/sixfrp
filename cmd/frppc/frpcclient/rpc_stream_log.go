package frpcclient

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"haokun-panel/cmd/frppc/facades"
	"haokun-panel/cmd/frppc/streamlog"

	pb "cnb.cool/sixkun/sixfrp-client-proto/sixfrp/client_pb"
)

func (s *Service) handleStartStreamLogWrapped(req *pb.StartSteamLogRequest) (*pb.CommonResponse, error) {
	if err := s.startStreamLog(req.GetPkgs()); err != nil {
		return nil, err
	}
	return &pb.CommonResponse{
		Status: &pb.Status{Code: pb.RespCode_RESP_CODE_SUCCESS, Message: "ok"},
	}, nil
}

func (s *Service) handleStopStreamLogWrapped(req *pb.CommonRequest) (*pb.CommonResponse, error) {
	if err := s.stopStreamLog(); err != nil {
		return nil, err
	}
	return &pb.CommonResponse{
		Status: &pb.Status{Code: pb.RespCode_RESP_CODE_SUCCESS, Message: "ok"},
	}, nil
}

func (s *Service) startStreamLog(pkgs []string) error {
	if err := s.stopStreamLog(); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(s.Context())
	stream, err := s.masterClient.PushClientStreamLog(ctx)
	if err != nil {
		cancel()
		return err
	}

	ch := make(chan string, 256)
	done := make(chan struct{})

	s.streamLogMu.Lock()
	s.streamLogCancel = cancel
	s.streamLogCh = ch
	s.streamLogDone = done
	s.streamLogMu.Unlock()

	streamlog.Default.Attach(ch)

	go s.runStreamLogPump(stream, ch, done)

	msg := "frppc stream log started"
	if len(pkgs) > 0 {
		msg = fmt.Sprintf("frppc stream log started, pkgs=%s", strings.Join(pkgs, ","))
	}
	s.writeStreamLog(msg)
	facades.Log().Infof("start stream log requested, pkgs=%v", pkgs)
	return nil
}

func (s *Service) stopStreamLog() error {
	s.streamLogMu.Lock()
	cancel := s.streamLogCancel
	done := s.streamLogDone
	s.streamLogCancel = nil
	s.streamLogCh = nil
	s.streamLogDone = nil
	s.streamLogMu.Unlock()

	streamlog.Default.Detach()

	if cancel == nil {
		return nil
	}
	cancel()

	if done == nil {
		return nil
	}
	select {
	case <-done:
		return nil
	case <-time.After(3 * time.Second):
		return fmt.Errorf("stop stream log timeout")
	}
}

func (s *Service) writeStreamLog(msg string) {
	if strings.TrimSpace(msg) == "" {
		return
	}
	s.streamLogMu.Lock()
	ch := s.streamLogCh
	s.streamLogMu.Unlock()
	if ch == nil {
		return
	}
	select {
	case ch <- msg:
	default:
		facades.Log().Warning("stream log channel full, dropping message")
	}
}

func (s *Service) runStreamLogPump(
	stream pb.ClientMaster_PushClientStreamLogClient,
	ch <-chan string,
	done chan<- struct{},
) {
	defer close(done)
	for {
		select {
		case <-stream.Context().Done():
			_, _ = stream.CloseAndRecv()
			return
		case msg := <-ch:
			encoded := base64.StdEncoding.EncodeToString([]byte(msg))
			if err := stream.Send(&pb.PushClientStreamLogReq{
				Log: []byte(encoded),
				Base: &pb.ClientBase{
					ClientId:     s.opts.clientID,
					ClientSecret: s.opts.clientSecret,
				},
			}); err != nil {
				facades.Log().Warningf("send stream log failed: %v", err)
				_, _ = stream.CloseAndRecv()
				return
			}
		}
	}
}
