package conf

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/infraboard/mcube/logger/zap"
)

// conf package 全部变量
// 全局配置对象
var global *Config

// 全局配置对象的访问方式
func C() *Config {
	if global == nil {
		panic("config required !")
	}
	return global
}

// 全局配置对象的设置方式
func SetGlobalConfig(conf *Config) {
	global = conf
}

// 初始化默认配置
func NewDefaultConfig() *Config {
	return &Config{
		App:   newDefaultApp(),
		MySQL: newDefaultMySQL(),
		Log:   newDefaultLog(),
	}
}

type Config struct {
	App   *app
	MySQL *mysql
	Log   *log
}

// 配置是通过对象来进行映射的
// 我们定义的是，配置对象的数据结构

func newDefaultApp() *app {
	return &app{
		Name: "restful-api",
		Host: "127.0.0.1",
		Port: "8050",
		Key:  "default app key",
	}
}

func newDefaultMySQL() *mysql {
	return &mysql{
		Host:        "192.168.1.7",
		Port:        "3306",
		UserName:    "root",
		Password:    "mysql",
		Database:    "restful_api",
		MaxOpenConn: 100,
		MaxIdleConn: 20,
		MaxLifeTime: 10 * 60 * 60,
		MaxIdleTime: 5 * 60 * 60,
	}
}

func newDefaultLog() *log {
	return &log{
		// "debug"
		Level:  zap.DebugLevel.String(),
		To:     ToStdout,
		Format: TextFormat,
	}
}

// 应用程序本身的一些配置
type app struct {
	// restful-api
	Name string
	// 127.0.0.1, 0.0.0.0,
	Host string `toml:"host"`
	// 8080, 8050
	Port string `toml:"port"`
	// 比较敏感的数据, 入库时加密后的数据, 加密的密钥就是该配置
	Key string `toml:"key"`
}

func (a *app) Addr() string {
	return fmt.Sprintf("%s:%s", a.Host, a.Port)
}

// MySQL 数据库配置
type mysql struct {
	Host        string `toml:"host"`
	Port        string `toml:"port"`
	UserName    string `toml:"username"`
	Password    string `toml:"password"`
	Database    string `toml:"database"`
	MaxOpenConn int    `toml:"max_open_conn"`
	MaxIdleConn int    `toml:"max_idle_conn"`
	// 单位是秒
	MaxLifeTime int `toml:"max_life_time"`
	MaxIdleTime int `toml:"max_idle_time"`

	lock sync.Mutex
}

// 利用MySQL配置, 构建全局MySQL单例链接
var (
	db *sql.DB
)

// getDBConn use to get db connection pool
func (m *mysql) getDBConn() (*sql.DB, error) {
	var err error
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&multiStatements=true", m.UserName, m.Password, m.Host, m.Port, m.Database)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("connect to mysql<%s> error, %s", dsn, err.Error())
	}
	db.SetMaxOpenConns(m.MaxOpenConn)
	db.SetMaxIdleConns(m.MaxIdleConn)
	db.SetConnMaxLifetime(time.Second * time.Duration(m.MaxLifeTime))
	db.SetConnMaxIdleTime(time.Second * time.Duration(m.MaxIdleTime))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping mysql<%s> error, %s", dsn, err.Error())
	}
	return db, nil
}

// GetDB todo
func (m *mysql) GetDB() (*sql.DB, error) {
	// 加载全局数据量单例
	m.lock.Lock()
	defer m.lock.Unlock()
	if db == nil {
		conn, err := m.getDBConn()
		if err != nil {
			return nil, err
		}
		db = conn
	}
	return db, nil
}

// Log todo
// os.Getenv() 方式
type log struct {
	Level   string    `toml:"level" env:"LOG_LEVEL"`
	PathDir string    `toml:"path_dir" env:"LOG_PATH_DIR"`
	Format  LogFormat `toml:"format" env:"LOG_FORMAT"`
	To      LogTo     `toml:"to" env:"LOG_TO"`
}
