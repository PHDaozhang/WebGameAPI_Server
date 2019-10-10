package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
	"web-game-api/controllers/system"
	"web-game-api/core/cache"
	"web-game-api/logic/common"
	"web-game-api/logic/errorCode"
	"web-game-api/logic/gameEnum"
	"web-game-api/logic/services/WorldService"
	"web-game-api/logic/utils"

	"github.com/astaxie/beego/httplib"
	"github.com/astaxie/beego/logs"
)

type TestController struct {
	system.BaseController
}

func (this *TestController) TestJson() {
	//responseStr := "{\"a\":1}"

	testMap := map[string]interface{}{}
	testMap["a"] = 1
	testMap["b"] = 2

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	encoder.Encode(testMap)

	str := buf.String()
	fmt.Println(str)

	outputJsonData := map[string]interface{}{}
	err := json.Unmarshal([]byte(str), &outputJsonData)

	if err != nil {
		fmt.Println("go this...", err)
	}

	orderRedisKey := "testRedis11"

	orderRedisValue := cache.GetRedis().SetNX(orderRedisKey, gameEnum.ORDER_REDIS_STATUS_UN_HANDLER, cache.DAY_1)

	logs.Trace("val:", orderRedisValue.Val())

	this.TraceJson()
}

// @Title 测试登陆
// @Description 测试登陆
// @Success 200 {object} admin.Admin
// @Param    agent    query    int    false  代理账号
// @Param    currency  query    int    false  		玩家登陆上分值
// @Param    account    query    int    false  		玩家的账号id
// @router   /login [get]
func (this *TestController) TestLogin() {
	agent, _ := this.GetInt("agent", 1)
	account := this.GetString("account", "test1")
	currency, _ := this.GetInt("currency", 0)
	appKey := "testKey"
	cur_tm_ms := time.Now().Unix() * 1000
	url := "http://127.0.0.1:8080"
	url += "/apiurl/user/login"
	url += "?"
	url += "account=" + account
	url += "&agent=" + strconv.Itoa(agent)
	url += "&timestamp=" + strconv.FormatInt(cur_tm_ms, 10)
	url += "&ip=127.0.0.1"
	url += "&homeUrl=http://www.baidu.com"
	url += "&gameId=1"
	if currency > 0 {
		url += "&currency=" + strconv.Itoa(currency)
		url += "&orderId=" + utils.GenOrderId(agent, account)
	}
	url += "&sign=" + utils.GenSign(agent, cur_tm_ms, appKey)

	logs.Trace(url)

	req := httplib.Get(url)

	responseStr, err := req.String()

	if err != nil {
		this.Code = errorCode.INNER_ERROR
		this.Msg = "向游戏服务器请求数据错误"
		return
	}

	responseBody := common.ReturnMsgStruct{}
	err = json.Unmarshal([]byte(responseStr), &responseBody)

	if err != nil {
		logs.Error(responseStr)
		this.Code = errorCode.INNER_ERROR
		this.Msg = "服务器过来数据格式有误"
		return
	}

	this.Code = responseBody.Code
	this.Msg = responseBody.Msg
	if responseBody.Data == nil {
		this.Result = ""
	} else {
		this.Result = responseBody.Data
	}

	this.TraceJson()
}

// @Title 测试转账
// @Description 测试登陆
// @Success 1 {object} admin.Admin
// @Param    agent    query    int    false  代理账号
// @Param    money  query    int    false  		玩家登陆上分值
// @Param    account    query    int    false  		玩家的账号id
// @Param    type    query    int    false  		玩家的账号id
// @router   /trans [get]
func (this *TestController) Trans() {
	appKey := "testKey"

	agent, _ := this.GetInt("agent", 1)
	account := this.GetString("account", "test1")
	money, _ := this.GetInt("money", 1000)
	ac_type, _ := this.GetInt("type", 1)
	timestamp := time.Now().Unix() * 1000
	randomOrderId := utils.GenOrderId(agent, account)
	orderId := this.GetString("orderId", randomOrderId)
	sign := utils.GenSign(agent, timestamp, appKey)

	url := "http://127.0.0.1:8080"
	url += "/apiurl/money/trans"
	url += "?"
	url += "agent=" + strconv.Itoa(agent)
	url += "&account=" + account
	url += "&timestamp=" + strconv.FormatInt(timestamp, 10)
	url += "&money=" + strconv.Itoa(money)
	url += "&type=" + strconv.Itoa(ac_type)
	url += "&orderId=" + orderId
	url += "&sign=" + sign

	logs.Trace(url)

	req := httplib.Get(url)

	responseStr, err := req.String()

	if err != nil {
		this.Code = errorCode.INNER_ERROR
		this.Msg = "向游戏服务器请求数据错误"
		return
	}

	responseBody := common.ReturnMsgStruct{}
	err = json.Unmarshal([]byte(responseStr), &responseBody)

	if err != nil {
		logs.Error(responseStr)
		this.Code = errorCode.INNER_ERROR
		this.Msg = "服务器过来数据格式有误"
		return
	}

	this.Code = responseBody.Code
	this.Msg = responseBody.Msg
	this.Result = responseBody.Data

	this.TraceJson()
}

// @Title 测试查询订单
// @Description 测试查询订单
// @Success 1 {object} admin.Admin
// @Param    agent    query    int    false  代理账号
// @Param    orderId  query    int    false  		订单编号
// @router   /trans [get]
func (this *TestController) QueryOrder() {
	appKey := "testKey"

	agent, _ := this.GetInt("agent", 1)
	timestamp := time.Now().Unix() * 1000
	orderId := this.GetString("orderId")
	sign := utils.GenSign(agent, timestamp, appKey)

	url := "http://127.0.0.1:8080"
	url += "/apiurl/money/queryorder"
	url += "?"
	url += "agent=" + strconv.Itoa(agent)
	url += "&timestamp=" + strconv.FormatInt(timestamp, 10)
	url += "&orderId=" + orderId
	url += "&sign=" + sign

	logs.Trace(url)

	req := httplib.Get(url)

	responseStr, err := req.String()

	if err != nil {
		this.Code = errorCode.INNER_ERROR
		this.Msg = "向游戏服务器请求数据错误"
		return
	}

	responseBody := common.ReturnMsgStruct{}
	err = json.Unmarshal([]byte(responseStr), &responseBody)

	if err != nil {
		logs.Error(responseStr)
		this.Code = errorCode.INNER_ERROR
		this.Msg = "服务器过来数据格式有误"
		return
	}

	this.Code = responseBody.Code
	this.Msg = responseBody.Msg
	this.Result = responseBody.Data

	this.TraceJson()
}

// @Title 查询下当前玩家在某个服务器上面，返回其所在服务器上面的http服务器地址与端口
// @Description 测试查询订单
// @Success 1 {object} admin.Admin
// @Param    Account    query    string    false  玩家账号
// @router   /selectWorld [get]
func (this *TestController) SelectWorld() {
	account := this.GetString("Account","auto_0bd993d70bf94888926028f97f6f6d0c")

	path := WorldService.GetWorldHttpPath(account)
	this.Success(path)
}