package main

import (
	"github.com/taolx0/secKill/pkg/bootstrap"
	conf "github.com/taolx0/secKill/pkg/config"
	"github.com/taolx0/secKill/pkg/mysql"
	"github.com/taolx0/secKill/sk-admin/setup"
)

func main() {
	mysql.InitMysql(conf.MysqlConfig.Host, conf.MysqlConfig.Port, conf.MysqlConfig.User, conf.MysqlConfig.Pwd, conf.MysqlConfig.Db) // conf.MysqlConfig.Db
	//setup.InitEtcd()
	setup.InitZk()
	setup.InitServer(bootstrap.HttpConfig.Host, bootstrap.HttpConfig.Port)

}
