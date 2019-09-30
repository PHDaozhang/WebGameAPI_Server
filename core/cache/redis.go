package cache

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/go-redis/redis"
	"tsEngine/tsRedis"
)

var client *redis.Client

const(
	MINUTE_1 = 60 * 1000;
	MINUTE_5 = MINUTE_1 * 5;
	MINUTE_10 = MINUTE_1 * 10;
	MINUTE_15 = MINUTE_1 * 15;
	MINUTE_30 = MINUTE_1 * 30;
	HOUR_1 = MINUTE_1 * 60;
	HOUR_2 = HOUR_1 * 2;
	HOUR_3 = HOUR_1 * 3;
	HOUR_5 = HOUR_1 * 5;
	HOUR_12 = HOUR_1 * 12;
	DAY_1 = HOUR_1 * 24;
	DAY_2 = DAY_1 * 2;
	DAY_3 = DAY_1 * 3;
	DAY_5 = DAY_1 * 5;
	DAY_15 = DAY_1 * 15;
	WEEK_1 = DAY_1 * 7;
	WEEK_2 = WEEK_1 * 2;
	MONTH_1 = DAY_1 * 30;
	MONTH_2 = MONTH_1 * 2;
	MONTH_3 = MONTH_1 * 3;
)



func InitRedis() bool{
	fmt.Println("init the redis")
	redis_host_port := beego.AppConfig.String("redis_host_port")

	redis_pwd := beego.AppConfig.String("redis_pwd")
	client = redis.NewClient(&redis.Options{
		Addr:redis_host_port,
		Password:redis_pwd,
	})

	if err := tsRedis.InitPool(redis_host_port, 1, redis_pwd, 10); err != nil {
		logs.Error("连接redis失败:", err)
		return false
	}

	_,err := client.Ping().Result()

	if err == nil {
		logs.Trace("连接redis成功")
	} else {
		logs.Error("连接redis失败2")
		return false
	}

	return true
}

func DoNothing(){

}

func Test(){
	client.Set("czh","czhddata",0)
	data := client.Get("czh")

	fmt.Println("data:",data.Val())
}

func GetRedis() *redis.Client{
	return client
}