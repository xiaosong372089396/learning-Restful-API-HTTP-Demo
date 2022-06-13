package protocol

import (
	"context"
	"fmt"
	"net/http"
	"time"

	hostAPI "xiaosong372089396/learning-Restful-API-HTTP-Demo/apps/host/http"
	"xiaosong372089396/learning-Restful-API-HTTP-Demo/conf"

	"github.com/infraboard/mcube/logger"
	"github.com/infraboard/mcube/logger/zap"
	"github.com/julienschmidt/httprouter"
)

func NewHTTPService() *HTTPService {
	r := httprouter.New()
	return &HTTPService{
		r: r,
		l: zap.L().Named("HTTP Server"),
		server: &http.Server{
			// http server监听地址
			Addr: conf.C().App.Addr(),
			// http handler/router
			Handler: r,
			// 读取Header 超时设置
			ReadHeaderTimeout: 60 * time.Second,
			// 链接， client --> Server
			// 请求 1g， 60s 读取完, 就超时
			ReadTimeout: 60 * time.Second,
			// 服务端详情数据超时
			// resp 1g, 60s读取不完
			WriteTimeout: 60 * time.Second,
			// http tcp 复用
			IdleTimeout: 60 * time.Second,
			// header 大小控制
			MaxHeaderBytes: 1 << 20, // 1M
		},
	}
}

// HTTPService http服务
type HTTPService struct {
	// router, root router, 路由, method+path --> handler
	r *httprouter.Router
	// 日志
	l logger.Logger
	// 配置
	c *conf.Config
	// 服务的实例对象, http服务器
	server *http.Server
}

// 启用HTTP 服务
func (s *HTTPService) Start() error {
	// 装置子服务路由
	// host http api 服务模块, 初始化
	hostAPI.API.Init()
	hostAPI.API.Registry(s.r)

	// 启动 HTTP服务
	s.l.Infof("HTTP服务启动成功, 监听地址: %s", s.server.Addr)
	if err := s.server.ListenAndServe(); err != nil {
		if err == http.ErrServerClosed {
			s.l.Info("service is stopped")
		}
		return fmt.Errorf("start service error, %s", err.Error())
	}
	return nil
}

// 关闭http 服务
func (s *HTTPService) Stop() error {
	s.l.Info("start graceful shutdown")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	// 优雅关闭HTTP服务
	if err := s.server.Shutdown(ctx); err != nil {
		s.l.Errorf("graceful shutdown timeout, force exit")
	}
	return nil
}
