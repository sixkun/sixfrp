package frpcclient

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	v1 "github.com/fatedier/frp/pkg/config/v1"
	"github.com/samber/lo"

	"haokun-panel/cmd/frppc/facades"
	"haokun-panel/utils"

	pb "cnb.cool/sixkun/sixfrp-client-proto/sixfrp/client_pb"
)

func (s *Service) PullConfig(ctx context.Context) error {
	return s.pullConfigForID(ctx, s.opts.clientID)
}

func (s *Service) pullConfigForID(ctx context.Context, clientID string) error {
	if !s.ShouldRun() {
		return nil
	}

	facades.Log().Infof("start to pull client config, clientID=%s", clientID)
	resp, err := s.masterClient.PullClientConfig(ctx, &pb.PullClientConfigReq{
		Base: &pb.ClientBase{
			ClientId:     clientID,
			ClientSecret: s.opts.clientSecret,
		},
	})
	if err != nil {
		facades.Log().Warningf("pull client config failed, clientID=%s err=%v", clientID, err)
		return err
	}

	cli := resp.GetClient()
	if cli == nil {
		facades.Log().Infof("empty pull config response, clientID=%s wait for server init", clientID)
		return nil
	}

	if cli.GetStopped() {
		facades.Log().Infof("client %s is stopped by master, stopping local frpc", clientID)
		s.controller.StopByClient(clientID)
		return nil
	}

	if len(cli.GetOriginClientId()) == 0 {
		currentClientIDs := s.controller.List()
		if idsToRemove, _ := lo.Difference(cli.GetClientIds(), currentClientIDs); len(idsToRemove) > 0 {
			facades.Log().Infof("client %s has %d expired child clients, removing: %+v", clientID, len(idsToRemove), idsToRemove)
			for _, id := range idsToRemove {
				s.controller.StopByClient(id)
				s.controller.DeleteByClient(id)
			}
		}
	}

	// shadow client: pull each child config
	if len(cli.GetClientIds()) > 0 {
		for _, id := range cli.GetClientIds() {
			if id == clientID {
				facades.Log().Infof("client %s is shadow client, skipping self", clientID)
				continue
			}
			if err := s.pullConfigForID(ctx, id); err != nil {
				facades.Log().Warningf("pull child client config failed, id=%s err=%v", id, err)
			}
		}
		return nil
	}

	if len(cli.GetConfig()) == 0 {
		facades.Log().Infof("client %s config is empty, wait for server init", clientID)
		return nil
	}

	// Save config to storage in debug mode
	if facades.Config().GetBool("app.debug") {
		if err := s.saveConfigToStorage(clientID, cli.GetConfig()); err != nil {
			facades.Log().Warningf("failed to save config to storage, clientID=%s err=%v", clientID, err)
		}
	}

	commonCfg, proxyCfgs, visitorCfgs, err := utils.LoadClientConfig([]byte(cli.GetConfig()), true)
	if err != nil {
		facades.Log().Warningf("cannot load client config, clientID=%s err=%v", clientID, err)
		return err
	}

	serverID := cli.GetServerId()

	existing := s.controller.Get(clientID, serverID)
	if existing == nil {
		facades.Log().Infof("client %s for server %s not exists, creating", clientID, serverID)
		return s.recreateHandler(clientID, serverID, commonCfg, proxyCfgs, visitorCfgs)
	}

	if !reflect.DeepEqual(existing.GetCommonCfg(), commonCfg) {
		facades.Log().Infof("client %s for server %s common config changed, recreating", clientID, serverID)
		return s.recreateHandler(clientID, serverID, commonCfg, proxyCfgs, visitorCfgs)
	}

	if !existing.Running() {
		facades.Log().Infof("client %s for server %s not running, recreating", clientID, serverID)
		return s.recreateHandler(clientID, serverID, commonCfg, proxyCfgs, visitorCfgs)
	}

	facades.Log().Infof("client %s for server %s common config unchanged, updating proxies", clientID, serverID)
	existing.Update(proxyCfgs, visitorCfgs)
	return nil
}

func (s *Service) recreateHandler(clientID, serverID string, commonCfg *v1.ClientCommonConfig, proxyCfgs []v1.ProxyConfigurer, visitorCfgs []v1.VisitorConfigurer) error {
	if existing := s.controller.Get(clientID, serverID); existing != nil {
		existing.Stop()
		s.controller.Delete(clientID, serverID)
	}
	h, err := NewClientHandler(commonCfg, proxyCfgs, visitorCfgs)
	if err != nil {
		facades.Log().Errorf("create client handler failed, clientID=%s serverID=%s err=%v", clientID, serverID, err)
		return err
	}
	s.controller.Add(clientID, serverID, h)
	s.controller.Run(clientID, serverID)
	return nil
}

// saveConfigToStorage saves the pulled config to storage directory in debug mode
func (s *Service) saveConfigToStorage(clientID, config string) error {
	storageDir := "storage/frppc-configs"
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		return fmt.Errorf("create storage directory failed: %w", err)
	}

	filename := filepath.Join(storageDir, fmt.Sprintf("%s_config.json", clientID))

	if err := os.WriteFile(filename, []byte(config), 0644); err != nil {
		return fmt.Errorf("write config file failed: %w", err)
	}

	facades.Log().Infof("saved config to %s", filename)
	return nil
}
