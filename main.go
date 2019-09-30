package main

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/astaxie/beego/plugins/cors"
	"os"
	"time"
	"tsEngine/tsDb"
	"tsEngine/tsRand"
	"web-game-api/conf"
	"web-game-api/core/cache"
	"web-game-api/core/mulMongo"
	"web-game-api/logic/filters"
	_ "web-game-api/routers"
)

func init()  {

}

func main() {
	logs.SetLogger("file", `{"filename":"./logs/logs.log", "perm":"0775", "maxDays":15}`)
	logs.SetLogFuncCall(true)
	logs.SetLogFuncCallDepth(3)
	//beego.ErrorController(&base.ErrorController{})

	err :=  tsDb.ConnectDbFormatConfig("mysql", conf.DBHost, conf.DBPort, conf.DBUser, conf.DBPassword, conf.DBName)
	if err != nil {
		logs.Error("连接数据库失败,default")
		os.Exit(1)
	}

	//err = mulMysql.ConnectDbFormatConfig("mysql",conf.HISTORY_DBHost,conf.HISTORY_DBPort,conf.HISTORY_DBUser,conf.HISTORY_DBPassword,conf.HISTORY_DBName,conf.HISTORY_ALIAS_NAME)
	//if err != nil {
	//	logs.Error("连接数据库失败：history")
	//	os.Exit(1)
	//}

	orm.RunSyncdb("default",false,true)

	beego.BConfig.WebConfig.DirectoryIndex = true
	beego.BConfig.WebConfig.StaticDir ["/swagger"] = "swagger"


	mulMongo.InitMongo("Mongodb_")						//AccountDB
	mulMongo.InitMongo("Mongodb_PLAYER_")					//PLayerDB_DWC
	mulMongo.InitMongo("Mongodb_Pay_")					//PaymentDB
	mulMongo.InitMongo("Mongodb_CONFIG_")					//ConfigDB

	beego.InsertFilter("*",beego.BeforeExec,filters.TimeGapFitler)

	beego.InsertFilter("*", beego.BeforeRouter, cors.Allow(&cors.Options{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"Origin", "Authorization", "Access-Control-Allow-Origin","token"},
		ExposeHeaders:    []string{"Content-Length", "Access-Control-Allow-Origin"},
		AllowCredentials: true,
	}))

	cache.InitRedis();

	tsRand.Seed(time.Now().UnixNano())

	beego.Run()
}

