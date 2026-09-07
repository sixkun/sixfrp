package frpcclient

import (
	"context"

	"github.com/fatedier/frp/client"
	"github.com/fatedier/frp/client/proxy"
	"github.com/fatedier/frp/pkg/config/source"
	v1 "github.com/fatedier/frp/pkg/config/v1"
	"github.com/fatedier/frp/pkg/config/v1/validation"
	"github.com/fatedier/frp/pkg/policy/featuregate"
	"github.com/fatedier/frp/pkg/policy/security"
	"github.com/samber/lo"

	"cnb.cool/sixkun/sixfrp/v2/cmd/frppc/facades"
	"cnb.cool/sixkun/sixfrp/v2/utils"
)

func NewClientHandler(commonCfg *v1.ClientCommonConfig, proxyCfgs []v1.ProxyConfigurer, visitorCfgs []v1.VisitorConfigurer) (ClientHandler, error) {
	if len(commonCfg.FeatureGates) > 0 {
		if err := featuregate.SetFromMap(commonCfg.FeatureGates); err != nil {
			facades.Log().Warningf("set feature gates failed (skip): %v, gates=%+v", err, commonCfg.FeatureGates)
		}
	}

	// Drop individually-invalid proxies/visitors instead of failing the whole
	// client. A single misconfigured proxy (e.g. an http proxy with neither
	// subdomain nor custom domains) would otherwise take down every other
	// tunnel on this client via ValidateAllClientConfig's fail-fast behavior.
	proxyCfgs = lo.Filter(proxyCfgs, func(c v1.ProxyConfigurer, _ int) bool {
		if err := validation.ValidateProxyConfigurerForClient(c); err != nil {
			facades.Log().Warningf("skip invalid proxy %s: %v", c.GetBaseConfig().Name, err)
			return false
		}
		return true
	})
	visitorCfgs = lo.Filter(visitorCfgs, func(c v1.VisitorConfigurer, _ int) bool {
		if err := validation.ValidateVisitorConfigurer(c); err != nil {
			facades.Log().Warningf("skip invalid visitor %s: %v", c.GetBaseConfig().Name, err)
			return false
		}
		return true
	})

	// In frp v0.69+, ValidateAllClientConfig requires UnsafeFeatures parameter.
	// Proxies/visitors were pre-filtered above, so this now only guards the
	// common config; any remaining error is a common-config problem.
	warning, err := validation.ValidateAllClientConfig(commonCfg, proxyCfgs, visitorCfgs, &security.UnsafeFeatures{})
	if warning != nil {
		facades.Log().Warningf("validate client config warning: %+v", warning)
	}
	if err != nil {
		return nil, err
	}

	// In frp v0.69+, ServiceOptions uses ConfigSourceAggregator instead of ProxyCfgs/VisitorCfgs
	configSource := source.NewConfigSource()
	if err := configSource.ReplaceAll(proxyCfgs, visitorCfgs); err != nil {
		return nil, err
	}
	aggregator := source.NewAggregator(configSource)

	cli, err := client.NewService(client.ServiceOptions{
		Common:                 commonCfg,
		ConfigSourceAggregator: aggregator,
		UnsafeFeatures:         &security.UnsafeFeatures{},
	})
	if err != nil {
		return nil, err
	}

	return &clientImpl{
		cli:         cli,
		common:      commonCfg,
		proxyCfgs:   lo.SliceToMap(proxyCfgs, utils.TransformProxyConfigurerToMap),
		visitorCfgs: lo.SliceToMap(visitorCfgs, utils.TransformVisitorConfigurerToMap),
	}, nil
}

func (c *clientImpl) Run() {
	c.runningMu.Lock()
	if c.running {
		c.runningMu.Unlock()
		facades.Log().Warning("client already running, skip Run")
		return
	}
	c.running = true
	c.done = make(chan struct{})
	c.runningMu.Unlock()

	defer func() {
		c.runningMu.Lock()
		c.running = false
		close(c.done)
		c.runningMu.Unlock()
	}()

	if err := c.cli.Run(context.Background()); err != nil {
		facades.Log().Errorf("run frpc client error: %v", err)
	}
}

func (c *clientImpl) Stop() {
	c.cli.Close()
}

func (c *clientImpl) Update(proxyCfgs []v1.ProxyConfigurer, visitorCfgs []v1.VisitorConfigurer) {
	c.proxyCfgs = lo.SliceToMap(proxyCfgs, utils.TransformProxyConfigurerToMap)
	c.visitorCfgs = lo.SliceToMap(visitorCfgs, utils.TransformVisitorConfigurerToMap)
	c.cli.UpdateAllConfigurer(proxyCfgs, visitorCfgs)
}

func (c *clientImpl) Running() bool {
	c.runningMu.RLock()
	defer c.runningMu.RUnlock()
	return c.running
}

func (c *clientImpl) GetCommonCfg() *v1.ClientCommonConfig {
	return c.common
}

func (c *clientImpl) GetProxyStatus(name string) (*proxy.WorkingStatus, bool) {
	return c.cli.StatusExporter().GetProxyStatus(name)
}
