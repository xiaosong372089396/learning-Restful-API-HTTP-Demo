package http

import (
	"xiaosong372089396/learning-Restful-API-HTTP-Demo/apps"
	"xiaosong372089396/learning-Restful-API-HTTP-Demo/apps/host"

	"github.com/infraboard/mcube/logger"
	"github.com/infraboard/mcube/logger/zap"
	"github.com/julienschmidt/httprouter"
)

// Host 模块的 HTTP API 服务实例
var API = handler{}

type handler struct {
	host host.Service
	log  logger.Logger
}

// 初始化的时候 依赖外部Host Service的实例对象  svr host.Service
func (h *handler) Init() {
	h.log = zap.L().Named("HOST API")

	if apps.Host == nil {
		panic("dependence host service is nil")
	}
	h.host = apps.Host
}

// 把Handler 实现的方法 注册给主路由
func (h *handler) Registry(r *httprouter.Router) {
	r.POST("/hosts", h.CreateHost)
	r.GET("/hosts", h.QueryHost)
	// 路径匹配，路径参数/hosts/110001
	r.GET("/hosts/:id", h.DescribeHost)
	r.PUT("/hosts/:id", h.UpdateHost)
	r.PATCH("/hosts/:id", h.PatchHost)
	r.DELETE("/hosts/:id", h.DeleteHost)
}
