package mulMysql

import (
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"net/url"
)

func ConnectDbFormatConfig(driver_name,dbhost,dbport,dbuser,dbpassword,dbname,aliasName string) error {
	if dbport == "" {
		dbport = "3306"
	}

	dsn := ""
	if driver_name == "mysql" {
		dsn = dbuser + ":" + dbpassword + "@tcp(" + dbhost + ":" + dbport + ")/" + dbname + "?charset=utf8&loc=" + url.QueryEscape("Local")
	} else {
		dsn = fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable", dbuser, dbpassword, dbname, dbhost, dbport)
	}

	fmt.Println("数据库地址:", dsn)
	err := orm.RegisterDataBase(aliasName, driver_name, dsn)
	if err != nil {
		logs.Debug("数据库服务器链接失败；", err)
	} else {
		logs.Debug("数据库服务器链接成功")
	}

	return err
}