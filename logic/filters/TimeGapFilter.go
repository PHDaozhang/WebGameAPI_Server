package filters

import (
	"github.com/astaxie/beego/context"
	"github.com/astaxie/beego/logs"
	"strconv"
	"time"
	"web-game-api/core/cache"
	"web-game-api/logic/gameEnum"
)

//每个接口多长时间才能请求一次，间隔单位为s
var accountTimeGapMap map[string]int = map[string]int{}			//account账号限制
var agentTimeGapMap map[string]int = map[string]int{}			//agent账号限制

func init(){
	accountTimeGapMap["/apiurl/test/trans"] = 5
}

/*
每个接口的请求时间
 */
func TimeGapFitler(ctx *context.Context) {
	urlPath := ctx.Request.URL.Path

	gapTime,ok := accountTimeGapMap[urlPath]
	if ok {
		//如果是账号限制
		setTimeGapFilterByKey(ctx,"account",gapTime)
		return
	}

	gapTime,ok = agentTimeGapMap[urlPath]
	if ok {
		setTimeGapFilterByKey(ctx,"agent",gapTime)
		return
	}


	logs.Trace(gapTime)
}

func setTimeGapFilterByKey(ctx *context.Context,paramKey string,gapTime int) bool {
	urlPath := ctx.Request.URL.Path

	paramValue := GetString(paramKey,ctx);
	if  paramValue == "" {
		return true
	}

	cur_time :=  time.Now()
	cur_tm_s := cur_time.Unix()
	mapRedisKey :=  gameEnum.REDIS_FILTER_TIME_GAP + paramKey + ":" +  urlPath
	//redisKey := gameEnum.REDIS_FILTER_TIME_GAP + urlPath + ":" + account
	cmd := cache.GetRedis().Exists(mapRedisKey)
	if cmd.Val() != 1 {
		next_zero_tm := int(cur_tm_s) - cur_time.Hour() * 3600 - cur_time.Minute() * 60 - cur_time.Second() + 86400
		//如果不存在
		cache.GetRedis().HMSet(mapRedisKey,map[string]interface{}{})
		nextDayTime := time.Unix( int64(next_zero_tm),0)
		cache.GetRedis().ExpireAt(mapRedisKey,nextDayTime)			//设置过期时间为次日的0:0:0
	}

	if cmd := cache.GetRedis().HExists(mapRedisKey,paramValue);!cmd.Val() {
		//设置下次再次可以接收到信息的时间值
		cache.GetRedis().HMSet(mapRedisKey,map[string]interface{}{paramValue:int(cur_tm_s) + gapTime})
	} else {
		sliceCmd := cache.GetRedis().HMGet(mapRedisKey,paramValue)
		expireTmInt := sliceCmd.Val()[0]

		expireTmStr := expireTmInt.(string)
		expireTm,err := strconv.Atoi(expireTmStr)
		if err == nil {
			if  cur_tm_s <  int64(expireTm) {
				//表示不能访问，在冷却时间内，不能访问
				ctx.Redirect(302,"/apiurl/error/cooldown")
				return false
			} else {
				//能访问，则更新过冷却时间
				cache.GetRedis().HMSet(mapRedisKey,map[string]interface{}{paramValue:int(cur_tm_s) + gapTime})
			}
		}
	}

	return true
}

func GetString(key string,ctx *context.Context) string {
	if v := ctx.Input.Query(key); v != "" {
		return v
	}

	return ""
}