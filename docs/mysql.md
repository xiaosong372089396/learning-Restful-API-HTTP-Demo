# MYSQL 相关

# 容器安装

```
docker run --name mysqlserver -e MYSQL_ROOT_PASSWORD=123456 -p 3306:3306 -d mysql
```

使用sql语句来添加一行记录
```
INSERT INTO `resource` (id, vendor,region,zone,create_at,expire_at,category,type,instance_id,`name`,description,`status`,update_at,sync_at,sync_accout,public_ip,private_ip,pay_type,describe_bash,resource_bash) VALUES ('0001', 0, 'hangzhou', 'a', 1110, 1110, 'cat', 't', 'ins-01', 'host01', 'sql执行', 'running', 1100, 1100, 'xxxx', '127.0.0.1', '127.0.0.1', 'p01', 'xxx', 'xxx');

```