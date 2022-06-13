package impl

import (
	"database/sql"

	"xiaosong372089396/learning-Restful-API-HTTP-Demo/conf"

	"github.com/infraboard/mcube/logger"
	"github.com/infraboard/mcube/logger/zap"
)

var Service *impl = &impl{}

type impl struct {
	// 可以更换成你们熟悉的 Logrus, 标准库log, zap
	// mcube Log模块是包装的 zap实现
	log logger.Logger

	// 依赖数据库
	db *sql.DB
}

func (i *impl) Init() error {
	i.log = zap.L().Named("Host")

	db, err := conf.C().MySQL.GetDB()
	if err != nil {
		return err
	}

	i.db = db
	return nil
}
