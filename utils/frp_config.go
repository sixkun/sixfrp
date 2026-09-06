package utils

import (
	v1 "github.com/fatedier/frp/pkg/config/v1"
)

func TransformProxyConfigurerToMap(origin v1.ProxyConfigurer) (key string, r v1.ProxyConfigurer) {
	key = origin.GetBaseConfig().Name
	r = origin
	return
}

func TransformVisitorConfigurerToMap(origin v1.VisitorConfigurer) (key string, r v1.VisitorConfigurer) {
	key = origin.GetBaseConfig().Name
	r = origin
	return
}
