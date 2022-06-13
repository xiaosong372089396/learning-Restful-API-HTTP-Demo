package conf

import (
	"github.com/BurntSushi/toml"
	"github.com/caarlos0/env/v6"
)

// 从文件中加载
func LoadConfigFromToml(path string) error {
	// new 配置对象
	cfg := NewDefaultConfig()
	// 解析配置文件, 赋值给对象cfg
	_, err := toml.DecodeFile(path, cfg)
	if err != nil {
		return err
	}

	SetGlobalConfig(cfg)

	return nil
}

// 从环境变量中加载
func LoadConfigFromEnv() error {
	// new 配置对象
	cfg := NewDefaultConfig()
	// 解析配置文件, 赋值给对象cfg
	if err := env.Parse(cfg); err != nil {
		return err
	}
	SetGlobalConfig(cfg)
	return nil
}
