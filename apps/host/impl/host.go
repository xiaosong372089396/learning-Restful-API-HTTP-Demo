package impl

import (
	"context"
	"database/sql"
	"fmt"

	"gitee.com/go-course/restful-api-demo/apps/host"
	"github.com/infraboard/mcube/exception"
	"github.com/infraboard/mcube/sqlbuilder"
	"github.com/infraboard/mcube/types/ftime"
	"github.com/rs/xid"
)

func (i *impl) CreateHost(ctx context.Context, ins *host.Host) (*host.Host, error) {

	// 校验数据合法性
	if err := ins.Validate(); err != nil {
		return nil, err
	}

	// 生成UUID的一个库,
	// snow 雪花算法
	// 分布式id, app, instance, ip, mac, ......, idc(), region,
	ins.Id = xid.New().String()
	if ins.CreateAt == 0 {
		ins.CreateAt = ftime.Now().Timestamp()
	}

	// 把数据入库到resource表和host表
	// 一次需要往2个表录入数据, 我们需要2个操作，要么都成功，要么都失败, 事物的逻辑

	// 全局异常
	var (
		resStmt  *sql.Stmt
		descStmt *sql.Stmt
		err      error
	)

	// 初始化一个事务, 所有的操作都使用这个事务来进行提交
	tx, err := i.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	// 函数执行完成后, 专门判断事务是否正常
	defer func() {
		// 事务执行有异常
		if err != nil {
			err := tx.Rollback()
			i.log.Debugf("tx rollback error, %s", err)
		} else {
			err := tx.Commit()
			i.log.Debugf("tx commit error, %s", err)
		}
	}()

	// 需要判断事务执行过程当中是否有异常
	// 有异常 就会滚事务, 无异常就提交事务

	// 在这个事务里只想insert sql,  先执行Prepare，防止sql注入攻击
	resStmt, err = tx.Prepare(InsertResourceSQL)
	if err != nil {
		return nil, err
	}
	defer resStmt.Close()

	// 注意: Prepare 语句会占用mysql资源, 如果使用不关闭, 会导致 Prepare溢出, 是全局的
	_, err = resStmt.Exec(
		ins.Id, ins.Vendor, ins.Region, ins.Zone, ins.CreateAt, ins.ExpireAt, ins.Category, ins.Type, ins.InstanceId,
		ins.Name, ins.Description, ins.Status, ins.UpdateAt, ins.SyncAt, ins.SyncAccount, ins.PublicIP, ins.PrivateIP, ins.PayType, ins.ResourceHash, ins.DescribeHash,
	)
	if err != nil {
		return nil, err
	}

	// 同样的逻辑, 需要host的数据存入
	descStmt, err = tx.Prepare(InsertDescribeSQL)
	if err != nil {
		return nil, err
	}
	defer descStmt.Close()

	_, err = descStmt.Exec(
		ins.Id, ins.CPU, ins.Memory, ins.GPUAmount, ins.GPUSpec, ins.OSType, ins.OSName,
		ins.SerialNumber, ins.ImageID, ins.InternetMaxBandwidthOut,
		ins.InternetMaxBandwidthIn, ins.KeyPairName, ins.SecurityGroups,
	)
	if err != nil {
		return nil, err
	}

	return ins, nil
}

func (i *impl) QueryHost(ctx context.Context, req *host.QueryHostRequest) (*host.Set, error) {
	query := sqlbuilder.NewQuery(queryHostSQL).Order("create_at").Desc().Limit(int64(req.Offset()), uint(req.PageSize))

	// 用户输入了关键字
	// Prepare 占位符是?  '%kws%' 是一个整体, 是一个值
	if req.Keywords != "" {
		query.Where("r.name LIKE ?", "%"+req.Keywords+"%")
	}

	sqlStr, args := query.BuildQuery()
	i.log.Debugf("sql: %s, args: %v", sqlStr, args)

	//  Prepare
	stmt, err := i.db.Prepare(sqlStr)
	if err != nil {
		return nil, fmt.Errorf("prepare query host sql error, %s", err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(args...)
	if err != nil {
		return nil, fmt.Errorf("stmt query error, %s", err)
	}

	// 初始化需要返回的对象
	set := host.NewSet()

	// 迭代查询表里的数据
	for rows.Next() {
		ins := host.NewDefaultHost()
		if err := rows.Scan(
			&ins.Id, &ins.Vendor, &ins.Region, &ins.Zone, &ins.CreateAt, &ins.ExpireAt,
			&ins.Category, &ins.Type, &ins.InstanceId, &ins.Name,
			&ins.Description, &ins.Status, &ins.UpdateAt, &ins.SyncAt, &ins.SyncAccount,
			&ins.PublicIP, &ins.PrivateIP, &ins.PayType, &ins.ResourceHash, &ins.DescribeHash,
			&ins.Id, &ins.CPU,
			&ins.Memory, &ins.GPUAmount, &ins.GPUSpec, &ins.OSType, &ins.OSName,
			&ins.SerialNumber, &ins.ImageID, &ins.InternetMaxBandwidthOut, &ins.InternetMaxBandwidthIn,
			&ins.KeyPairName, &ins.SecurityGroups,
		); err != nil {
			return nil, err
		}
		set.Add(ins)

	}

	// Count 获取总数据量
	// build 一个count语句
	countStr, countArgs := query.BuildCount()
	countStmt, err := i.db.Prepare(countStr)
	if err != nil {
		return nil, fmt.Errorf("prepare count stmt error, %s", err)
	}
	defer countStmt.Close()

	if err := countStmt.QueryRow(countArgs...).Scan(&set.Total); err != nil {
		return nil, fmt.Errorf("query count stmt, %s", err)
	}

	return set, nil
}

func (i *impl) DesribeHost(ctx context.Context, req *host.DesribeHostRequest) (*host.Host, error) {
	query := sqlbuilder.NewQuery(queryHostSQL).Where("r.id = ?", req.Id)

	sqlStr, args := query.BuildQuery()
	i.log.Debugf("sql: %s, args: %v", sqlStr, args)

	//  Prepare
	stmt, err := i.db.Prepare(sqlStr)
	if err != nil {
		return nil, fmt.Errorf("prepare query host sql error, %s", err)
	}
	defer stmt.Close()

	ins := host.NewDefaultHost()
	err = stmt.QueryRow(args...).Scan(
		&ins.Id, &ins.Vendor, &ins.Region, &ins.Zone, &ins.CreateAt, &ins.ExpireAt,
		&ins.Category, &ins.Type, &ins.InstanceId, &ins.Name,
		&ins.Description, &ins.Status, &ins.UpdateAt, &ins.SyncAt, &ins.SyncAccount,
		&ins.PublicIP, &ins.PrivateIP, &ins.PayType, &ins.ResourceHash, &ins.DescribeHash,
		&ins.Id, &ins.CPU,
		&ins.Memory, &ins.GPUAmount, &ins.GPUSpec, &ins.OSType, &ins.OSName,
		&ins.SerialNumber, &ins.ImageID, &ins.InternetMaxBandwidthOut, &ins.InternetMaxBandwidthIn,
		&ins.KeyPairName, &ins.SecurityGroups,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, exception.NewNotFound("host %s not found", req.Id)
		}
		return nil, fmt.Errorf("stmt query error, %s", err)
	}
	return ins, nil
}

// 自己模仿 Insert 使用事务一次完成2个SQL操作
func (i *impl) UpdateHost(ctx context.Context, req *host.UpdateHostRequest) (*host.Host, error) {

	// 重新查询出来
	ins, err := i.DesribeHost(ctx, host.NewDescribeHostRequestWithID(req.Id))
	if err != nil {
		return nil, err
	}

	// 对象更新(PATCH/PUT)
	switch req.UpdateMode {
	case host.PUT:
		// 对象更新(全量更新)
		ins.Update(req.Resource, req.Describe)
	case host.PATCH:
		// 对象打补丁(部分更新)
		err := ins.Patch(req.Resource, req.Describe)
		if err != nil {
			return nil, err
		}
	}

	// 校验更新后的数据是否合法
	if err := ins.Validate(); err != nil {
		return nil, err
	}

	stmt, err := i.db.Prepare(updateResourceSQL)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	// DML
	// vendor=?,region=?,zone=?,expire_at=?,name=?,descruotion=? WHERE id = ?
	_, err = stmt.Exec(ins.Vendor, ins.Region, ins.Zone, ins.ExpireAt, ins.Name, ins.Description, ins.Id)
	if err != nil {
		return nil, err
	}

	return ins, nil
}

// 自己模仿 Insert 使用事务一次完成2个SQL操作
func (i *impl) DeleteHost(ctx context.Context, req *host.DeleteHostRequest) (*host.Host, error) {

	// 全局异常
	var (
		resStmt  *sql.Stmt
		descStmt *sql.Stmt
		err      error
	)

	// 重新查询出来
	ins, err := i.DesribeHost(ctx, host.NewDescribeHostRequestWithID(req.Id))
	if err != nil {
		return nil, err
	}

	// 初始化一个事务, 所有的操作都使用这个事务来进行提交
	tx, err := i.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	// 函数执行完成后, 专门判断事务是否正常
	defer func() {
		// 事务执行有异常
		if err != nil {
			err := tx.Rollback()
			i.log.Debugf("tx rollback error, %s", err)
		} else {
			err := tx.Commit()
			i.log.Debugf("tx commit error, %s", err)
		}
	}()

	resStmt, err = tx.Prepare(deleteResourceSQL)
	if err != nil {
		return nil, err
	}
	defer resStmt.Close()
	_, err = resStmt.Exec(req.Id)
	if err != nil {
		return nil, err
	}

	descStmt, err = tx.Prepare(deleteHostSQL)
	if err != nil {
		return nil, err
	}
	defer descStmt.Close()
	_, err = descStmt.Exec(req.Id)
	if err != nil {
		return nil, err
	}

	return ins, nil
}
