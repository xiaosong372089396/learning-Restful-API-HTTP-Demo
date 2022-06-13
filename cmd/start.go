package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"xiaosong372089396/learning-Restful-API-HTTP-Demo/apps"
	"xiaosong372089396/learning-Restful-API-HTTP-Demo/apps/host/impl"
	"xiaosong372089396/learning-Restful-API-HTTP-Demo/conf"
	"xiaosong372089396/learning-Restful-API-HTTP-Demo/protocol"

	"github.com/infraboard/mcube/logger"
	"github.com/infraboard/mcube/logger/zap"
	"github.com/spf13/cobra"
)

var (
	configType string
	confFile   string
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Demo后端API服务",
	Long:  `Demo后端API服务`,
	RunE: func(c *cobra.Command, args []string) error {
		// 加载全局配置
		if err := loadGlobalConfig(configType); err != nil {
			return err
		}

		// 初始化日志
		if err := loadGlobalLogger(); err != nil {
			return err
		}

		// 初始化服务层 Ioc 初始化
		if err := impl.Service.Init(); err != nil {
			return err
		}
		// 把服务实例注册给IOC层
		apps.Host = impl.Service

		// 启动服务后, 需要处理的事件
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP, syscall.SIGQUIT)

		// 启动服务
		svr := NewService(conf.C())

		// 等待程序退出
		go svr.waitSign(ch)

		// 启动服务
		return svr.Start()
	},
}

func NewService(conf *conf.Config) *Service {
	return &Service{
		conf: conf,
		http: protocol.NewHTTPService(),
		log:  zap.L().Named("service"),
	}
}

// service
// 服务的整体配置
// 服务可能启动很多模块; http, grpc, contable
type Service struct {
	conf *conf.Config
	http *protocol.HTTPService
	log  logger.Logger
}

func (s *Service) Start() error {
	return s.http.Start()
}

// 当发现用户收到终止掉程序的时候, 要完成处理
func (s *Service) waitSign(sign chan os.Signal) {
	for sg := range sign {
		switch v := sg.(type) {
		// 预留扩展, 啥时候 需要单独处理时, 在补充
		// 没做信号取反
		// reload
		// term
		// quick
		default:
			// 资源管理
			s.log.Infof("receive signal '%v', start graceful shudown", v.String())
			if err := s.http.Stop(); err != nil {
				s.log.Errorf("graceful shudown err: %s, force exit", err)
			}
			s.log.Infof("service stop complete")
			return
		}

	}
}

// config 为全局变量, 只需要load 即可全局可用户
func loadGlobalConfig(configType string) error {
	// 配置加载
	switch configType {
	case "file":
		err := conf.LoadConfigFromToml(confFile)
		if err != nil {
			return err
		}
	case "env":
		err := conf.LoadConfigFromEnv()
		if err != nil {
			return err
		}
	case "etcd":
		return errors.New("not implemented")
	default:
		return errors.New("unknown config type")
	}
	return nil
}

// log 为全局变量, 只需要load 即可全局可用户, 依赖全局配置先初始化
func loadGlobalLogger() error {
	var (
		logInitMsg string
		level      zap.Level
	)

	// 获取出日志配置对象
	lc := conf.C().Log

	// debug, info, xxx
	// 解析配置的日志级别是否正确
	lv, err := zap.NewLevel(lc.Level)
	if err != nil {
		// 解析失败, 默认使用info级别
		logInitMsg = fmt.Sprintf("%s, use default level INFO", err)
		level = zap.InfoLevel
	} else {
		// 解析成功，直接只用用户配置的日志级别
		level = lv
		logInitMsg = fmt.Sprintf("log level: %s", lv)
	}

	// 初始化了日志的默认配置
	zapConfig := zap.DefaultConfig()
	zapConfig.Level = level
	zapConfig.Files.RotateOnStartup = false
	switch lc.To {
	case conf.ToStdout:
		zapConfig.ToStderr = true
		zapConfig.ToFiles = false
	case conf.ToFile:
		zapConfig.Files.Name = "restful-api.log"
		zapConfig.Files.Path = lc.PathDir
	}
	switch lc.Format {
	case conf.JSONFormat:
		zapConfig.JSON = true
	}

	// 初始化全局Logger的配置
	if err := zap.Configure(zapConfig); err != nil {
		return err
	}

	// 全局Logger初始化后, 就可以正常使用
	zap.L().Named("INIT").Info(logInitMsg)
	return nil
}

func init() {
	RootCmd.AddCommand(startCmd)
	startCmd.PersistentFlags().StringVarP(&configType, "config_type", "t", "file", "the restful-api demo config type")
	startCmd.PersistentFlags().StringVarP(&confFile, "config_file", "f", "etc/restful-api.toml", "the restful-api config file path")
}
