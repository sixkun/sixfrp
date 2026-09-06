package frpcclient

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	pty "github.com/aymanbagabas/go-pty"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"

	"haokun-panel/cmd/frppc/facades"

	pb "cnb.cool/sixkun/sixfrp-client-proto/sixfrp/client_pb"
)

// defaultShell picks the interactive shell to attach to the pseudo-terminal.
// On Windows the pty is backed by ConPTY (Windows 10 1809+); on unix it is a
// classic /dev/ptmx pair. Each platform has no notion of the other's shell, so
// the default is chosen per-OS rather than assuming $SHELL exists.
func defaultShell() string {
	if runtime.GOOS == "windows" {
		// powershell.exe ships on every Windows 10+ target; fall back to
		// cmd.exe (via COMSPEC) only if it is somehow unavailable.
		if _, err := exec.LookPath("powershell.exe"); err == nil {
			return "powershell.exe"
		}
		if comspec := strings.TrimSpace(os.Getenv("COMSPEC")); comspec != "" {
			return comspec
		}
		return "cmd.exe"
	}
	if sh := strings.TrimSpace(os.Getenv("SHELL")); sh != "" {
		return sh
	}
	return "/bin/sh"
}

func (s *Service) handleStartPTYConnectWrapped(req *pb.CommonRequest) (*pb.CommonResponse, error) {
	if !s.opts.enableRemoteShell {
		return nil, fmt.Errorf("remote shell is disabled, set FRPPC_FEATURES_ENABLE_REMOTE_SHELL=true to enable")
	}

	if s.opts.totpSecret != "" {
		return s.respondTOTPGated()
	}

	return s.openPTYAndRespond()
}

func (s *Service) respondTOTPGated() (*pb.CommonResponse, error) {
	resp := &pb.PTYConnectResponse{}
	if !IsTOTPBound(s.opts.clientID) {
		totpURI := fmt.Sprintf("otpauth://totp/sixfrp:%s?secret=%s&issuer=sixfrp", s.opts.clientID, s.opts.totpSecret)
		resp.PtyConnectStatus = pb.PTYConnectStatus_PTY_CONNECT_STATUS_TOTP_BIND_REQUIRED
		resp.TotpUri = &totpURI
	} else {
		resp.PtyConnectStatus = pb.PTYConnectStatus_PTY_CONNECT_STATUS_TOTP_VERIFY_REQUIRED
	}
	return marshalPTYConnectResponse(resp)
}

func (s *Service) openPTYAndRespond() (*pb.CommonResponse, error) {
	sessionID, err := s.startPTYConnectSession()
	if err != nil {
		return nil, err
	}
	resp := &pb.PTYConnectResponse{
		PtyConnectStatus: pb.PTYConnectStatus_PTY_CONNECT_STATUS_READY,
		SessionId:        &sessionID,
	}
	return marshalPTYConnectResponse(resp)
}

func marshalPTYConnectResponse(resp *pb.PTYConnectResponse) (*pb.CommonResponse, error) {
	raw, err := proto.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("marshal PTYConnectResponse failed: %w", err)
	}
	data := string(raw)
	return &pb.CommonResponse{
		Status: &pb.Status{Code: pb.RespCode_RESP_CODE_SUCCESS, Message: "ok"},
		Data:   &data,
	}, nil
}

func (s *Service) startPTYConnectSession() (string, error) {
	conn, err := s.masterClient.PTYConnect(s.Context())
	if err != nil {
		return "", err
	}

	sessionID := uuid.New().String()
	if err := conn.Send(&pb.PTYClientMessage{
		SessionId: sessionID,
		Base: &pb.ClientBase{
			ClientId:     s.opts.clientID,
			ClientSecret: s.opts.clientSecret,
		},
	}); err != nil {
		return "", err
	}

	ack, err := conn.Recv()
	if err != nil {
		return "", err
	}
	if string(ack.GetData()) != "ok" {
		return "", fmt.Errorf("pty connect ack error")
	}

	go s.bridgePTY(conn, sessionID)
	return sessionID, nil
}

func (s *Service) bridgePTY(conn pb.ClientMaster_PTYConnectClient, sessionID string) {
	shell := defaultShell()

	ptmx, err := pty.New()
	if err != nil {
		facades.Log().Warningf("open pty failed: %v", err)
		_ = conn.Send(&pb.PTYClientMessage{Data: []byte("failed to start tty: " + err.Error()), SessionId: sessionID})
		_ = conn.CloseSend()
		return
	}

	cmd := ptmx.Command(shell)
	if err := cmd.Start(); err != nil {
		facades.Log().Warningf("start pty shell failed: %v", err)
		_ = conn.Send(&pb.PTYClientMessage{Data: []byte("failed to start tty: " + err.Error()), SessionId: sessionID})
		_ = ptmx.Close()
		_ = conn.CloseSend()
		return
	}

	defer func() {
		_ = ptmx.Close()
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		_ = cmd.Wait()
		_ = conn.CloseSend()
	}()

	errCh := make(chan error, 2)
	var once sync.Once
	done := func(err error) {
		once.Do(func() { errCh <- err })
	}

	// tty -> master
	go func() {
		buf := make([]byte, 4096)
		for {
			n, readErr := ptmx.Read(buf)
			if n > 0 {
				if sendErr := conn.Send(&pb.PTYClientMessage{
					Data:      append([]byte(nil), buf[:n]...),
					SessionId: sessionID,
				}); sendErr != nil {
					done(sendErr)
					return
				}
			}
			if readErr != nil {
				done(readErr)
				return
			}
		}
	}()

	// master -> tty
	go func() {
		for {
			msg, recvErr := conn.Recv()
			if recvErr != nil {
				done(recvErr)
				return
			}
			if msg.GetDone() {
				done(io.EOF)
				return
			}
			if msg.Height != nil && msg.Width != nil {
				if err := ptmx.Resize(int(*msg.Width), int(*msg.Height)); err != nil {
					facades.Log().Warningf("resize pty failed: %v", err)
				}
				continue
			}
			if len(msg.GetData()) == 0 {
				continue
			}
			if _, err := ptmx.Write(msg.GetData()); err != nil {
				done(err)
				return
			}
		}
	}()

	bridgeErr := <-errCh
	if bridgeErr != nil && bridgeErr != io.EOF {
		facades.Log().Warningf("pty bridge closed: %v", bridgeErr)
	}
}
