package mysql

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gohouse/gorose/v2"
	"log"
)

var engine *gorose.Engin
var err error

func InitMysql(hostMysql, portMysql, userMysql, pwdMysql, dbMysql string) {
	log.Printf("user %s is using database : %s\n", userMysql, dbMysql)

	DbConfig := gorose.Config{
		// Default database configuration
		Driver: "mysql",                                                                                                              // Database driver(mysql,sqlite,postgres,oracle,mssql)
		Dsn:    userMysql + ":" + pwdMysql + "@tcp(" + hostMysql + ":" + portMysql + ")/" + dbMysql + "?charset=utf8&parseTime=true", // 数据库链接
		Prefix: "",                                                                                                                   // Table prefix
		// (Connection pool) Max open connections, default value 0 means unLimit.
		SetMaxOpenConns: 300,
		// (Connection pool) Max idle connections, default value is 1.
		SetMaxIdleConns: 10,
	}

	engine, err = gorose.Open(&DbConfig)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func DB() gorose.IOrm {
	return engine.NewOrm()
}
