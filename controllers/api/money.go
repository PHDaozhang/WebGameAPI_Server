package api

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/httplib"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"gopkg.in/mgo.v2/bson"
	"strconv"
	"time"
	"tsEngine/tsRedis"
	"web-game-api/controllers/system"
	"web-game-api/core/cache"
	"web-game-api/core/mulMongo"
	"web-game-api/logic/common"
	"web-game-api/logic/errorCode"
	"web-game-api/logic/gameEnum"
	"web-game-api/logic/utils"
	"web-game-api/models/api"
)

//@title apiurl/money
//转账相关
type MoneyController struct {
	system.BaseController
}

// @Title 转账接口 包括上下分相关
// @Description 第三方账号登陆
// @Success 1 {object} {url:,isNew:,orderCode:}
// @Param    agent    query    int    true  代理账号
// @Param    timestamp       query    int    true   时间戳 以ms为单位
// @Param    money       query    string    true   登陆的玩家账号
// @Param    type   query    string    true   	玩家点击退出后的地址
// @Param    orderId    query    int    false  		当玩家上分值有效时，此值必须存在
// @Param    sign    query    string    true  		玩家签名
// @router   /trans [get,post]
func (this *MoneyController)Trans(){
	//检查参数是否合法
	agent,_ := this.GetInt64("agent",0);
	timestamp,_ := this.GetInt64("timestamp",0);
	account := this.GetString("account");
	money,_ := this.GetInt64("money",0);
	trans_type,_ := this.GetInt64("type",0);
	orderId := this.GetString("orderId");
	sign := this.GetString("sign");

	if agent == 0 || timestamp == 0 || account == "" || money == 0 || trans_type == 0 || orderId == "" || sign == ""{
		this.Error(errorCode.PARAM_ERROR,"参数错误")
	}

	o := orm.NewOrm()
	//查看代理是否存在
	agentOrm := new(api.Agent)
	agentOrm.Id = int64(agent)
	err := o.Read(agentOrm)
	if err != nil {
		this.Error(errorCode.AGENT_UNEXIST,"代理不存在")
	}

	var checkSignFlag = utils.CheckSign(agent,timestamp,agentOrm.AppKey,sign);
	if !checkSignFlag{
		this.Error(errorCode.SIGN_ERROR,"签名不正确")
	}

	returnMsgStruct := common.ReturnMsgStruct{}
	TransMoney(agent,timestamp,account,trans_type, orderId,sign,money,&returnMsgStruct);
	this.Code = int(returnMsgStruct.Code);
	this.Msg = returnMsgStruct.Msg;
	this.Result = returnMsgStruct.Data;

	//需要将订单日志记录下来
	this.TraceJson();
}

// @Title 查询订单
// @Description 第三方账号登陆
// @Success 1 {object} {url:,isNew:,orderCode:}
// @Param    agent    query    int    true  代理账号
// @Param    timestamp       query    int    true   时间戳 以ms为单位
// @Param    orderId    query    int    false  		当玩家上分值有效时，此值必须存在
// @Param    sign    query    string    true  		玩家签名
// @router   /queryorder [get,post]
func (this *MoneyController)QueryOrder(){
	//检查参数是否合法g
	//检查参数是否合法
	agent,_ := this.GetInt64("agent",0);
	timestamp,_ := this.GetInt64("timestamp",0);
	orderId := this.GetString("orderId");
	sign := this.GetString("sign");

	if agent == 0 || timestamp == 0 ||  orderId == "" || sign == ""{
		this.Code =  errorCode.PARAM_ERROR;
		this.Result = map[string]interface{}{}
		this.TraceJson();
		return;
	}

	o := orm.NewOrm()
	//查看代理是否存在
	agentOrm := new(api.Agent)
	agentOrm.Id = agent
	err := o.Read(agentOrm)
	if err != nil {
		this.Msg = "代理不存在"
		this.Code = errorCode.AGENT_UNEXIST
		this.Result = map[string]interface{}{}
		this.TraceJson()
		return
	}

	var checkSignFlag = utils.CheckSign(agent,timestamp,agentOrm.AppKey,sign);
	if !checkSignFlag{
		this.Msg = "签名不正确"
		this.Code = errorCode.SIGN_ERROR
		this.Result = map[string]interface{}{}
		this.TraceJson()
		return
	}

	//查看数据库中是否有此订单信息
	orderInfo := api.OrderInfo{}
	err =  mulMongo.FindOne(gameEnum.DB_NAME_PAYMENT,gameEnum.TABLE_PAY_COLLECTION_ORDER,bson.M{"OrderID": orderId},bson.M{},&orderInfo)
	if err != nil {
		this.Msg = "查询订单不存在"
		this.Code = errorCode.ORDER_UN_EXIST
		this.Result = map[string]interface{}{}
		this.TraceJson()
		return
	}

	//				1) status:
	//					a) 1：未处理
	//					b) 2：正在处理
	//					c) 3：处理成功
	//					d) 4：处理失败
	resultMap := map[string]interface{}{}
	this.Result  = resultMap
	if orderInfo.Process {
		resultMap["status"] = 3
		this.Msg = "订单已处理"
		this.Code = errorCode.SUCC
		this.Result = map[string]interface{}{}
		this.TraceJson()
		return
	}

	//查看订单是否存在
	orderRedisKey := gameEnum.REDIS_ORDER + orderId
	orderValue,err := tsRedis.Get(orderRedisKey)		//订单的时间戳在5分钟以内是有效的		值 0：未处理 1：正在处理  2：已处理 3:处理失败
	if err != nil {
		this.Msg = "查询redis键值不存在"
		this.Code = errorCode.INNER_ERROR
		this.Result = map[string]interface{}{}
		this.TraceJson()
		return
	}

	orderStatus,err := strconv.Atoi(orderValue)
	if err != nil {
		this.Msg = "状态转换错误"
		this.Code = errorCode.INNER_ERROR
		this.Result = map[string]interface{}{}
		this.TraceJson()
		return
	}

	if  orderStatus == gameEnum.ORDER_REDIS_STATUS_UN_HANDLER {
		this.Msg = "订单未处理"
		this.Code = errorCode.SUCC
		resultMap["status"] = 1
		this.TraceJson()
		return
	} else if orderStatus == gameEnum.ORDER_REDIS_STATUS_HANDING {
		this.Msg = "订单正在处理"
		this.Code = errorCode.SUCC
		resultMap["status"] = 2
		this.TraceJson()
		return
	} else if orderStatus == gameEnum.ORDER_REDIS_STATUS_HANDLER_ERR {
		this.Msg = "处理订单失败"
		this.Code = errorCode.SUCC
		resultMap["status"] = 4
		this.TraceJson()
		return
	}

	this.TraceJson();
}

/**
具体转账使用，在登陆或是转换接口中都可以使用
 */
func TransMoney(agent int64,timestamp int64,account string,trans_type int64,orderId string,sign string,money int64, returnMsg *common.ReturnMsgStruct ) {

	curTime := time.Now()
	curTimeUnix := curTime.Unix();
	timeGap := curTimeUnix * 1000 - timestamp
	timeGap /= 1000
	if timeGap > 300 {
		returnMsg.Code = errorCode.ORDER_TIME_EXPIRE
		return
	}

	gameAccountName := utils.GameAccount(agent,account)

	//查看订单是否存在
	orderRedisKey := gameEnum.REDIS_ORDER + orderId
	logs.Trace("orderId:",orderId)
	orderRedisValue := cache.GetRedis().SetNX(orderRedisKey,gameEnum.ORDER_REDIS_STATUS_UN_HANDLER,cache.DAY_1)				//订单的时间戳在5分钟以内是有效的		值 0：未处理 1：正在处理  2：已处理 3：失败
	if !orderRedisValue.Val() {
		returnMsg.Code = errorCode.ORDER_REPEAT_ERROR
		returnMsg.Msg = "订单重复"

		//如果未处理的订单，则向游戏服务器再次发起请求处理订单的需求
		cur_redis_cmd := cache.GetRedis().Get(orderRedisKey)
		orderStatus,err := strconv.Atoi(cur_redis_cmd.Val())
		if err == nil {
			if orderStatus == gameEnum.ORDER_REDIS_STATUS_HANDING {
				returnMsg.Msg = "正在处理"
			} else if orderStatus == gameEnum.ORDER_REDIS_STATUS_HANDED {
				returnMsg.Msg = "订单已处理"
			} else if orderStatus == gameEnum.ORDER_REDIS_STATUS_HANDLER_ERR {
				returnMsg.Msg = "订单处理失败"
			} else if orderStatus == gameEnum.ORDER_REDIS_STATUS_UN_HANDLER {
				//如果是未处理，则通知游戏服务器处理
				requestGameServer(gameAccountName,orderId,trans_type,returnMsg)
			}
		}
		return
	}


	OrderInfoList := []api.OrderInfo{}
	orderSelect := map[string]interface{}{}
	orderSelect["Account"] = gameAccountName
	orderSelect["Process"] = false					//未处理
	//orderSelect["State"] = 0						//初始化状态，此里可能需要与服务器		//支付状态  默认0：  1：  2：支付成功		[3：下分失败] 未与服务器配对
	dbErr := mulMongo.FindAll(gameEnum.DB_NAME_PAYMENT,gameEnum.TABLE_PAY_COLLECTION_ORDER,orderSelect,bson.M{},&OrderInfoList)
	if dbErr != nil {
		returnMsg.Code = errorCode.DB_OPER_ERROR
		returnMsg.Msg = "查询数据库失败"
		return
	}

	if len(OrderInfoList) >= 1 {
		returnMsg.Code = errorCode.ACCOUNT_ORDER_UN_OPER
		returnMsg.Msg = "有未处理的上下分订单"
		return
	}

	//查看当前是否有还有未处理的订单存在，如果有未处理的订单存在，则不能进行新的订单处理 end
	orderInfo := api.OrderInfo{}
	orderInfo.Id = bson.NewObjectId()
	orderInfo.OrderID = orderId
	orderInfo.Account = gameAccountName
	orderInfo.ChannelID = strconv.FormatInt(agent,10)
	orderInfo.RMB = money / 100
	orderInfo.RealRMB = orderInfo.RMB
	orderInfo.PayCode = money
	orderInfo.Custom = 0;
	orderInfo.PayPlatform = "api"
	orderInfo.CreateTime  = time.Now()
	orderInfo.Process = false

	if trans_type == gameEnum.TRANS_US_TO_OTHERS {			//如果是下分接口 则需要检查当前账号是否存在
		gameAccount := utils.GameAccount(agent,account)
		accountResult := api.AccountDBInfo{}
		err := mulMongo.FindOne(gameEnum.DB_NAME_ACCOUNT,gameEnum.ACCOUNT_COLLECTION_TABLE,bson.M{"Account":gameAccount},bson.M{},accountResult)
		if err != nil {
			returnMsg.Code = errorCode.ACCOUNT_UN_EXIST
			returnMsg.Msg = "此账号不存在"
			return
		}
	}

	if trans_type == gameEnum.TRANS_OTHER_TO_US {
		orderInfo.State = 2 //已支付
		orderInfo.Type = 1
	} else {
		orderInfo.State = 0				//未支付
		orderInfo.Type = 2
	}

	err := mulMongo.Insert(gameEnum.DB_NAME_PAYMENT,gameEnum.TABLE_PAY_COLLECTION_ORDER,&orderInfo)

	if err != nil {
		//插入失败
		returnMsg.Code = errorCode.DB_OPER_ERROR
		returnMsg.Msg = "插入上分订单进数据库失败"
		return
	}

	//更新缓存状态为正在处理
	cache.GetRedis().Set(orderRedisKey,gameEnum.ORDER_REDIS_STATUS_HANDING,0)

	requestGameServer(gameAccountName,orderId,trans_type,returnMsg)

	if returnMsg.Code != errorCode.SUCC {		//如果处理失败，则将状态还原为未处理状态
		cache.GetRedis().Set(orderRedisKey,gameEnum.ORDER_REDIS_STATUS_UN_HANDLER,0)
	}

	//将订单日志插入或是更新入订单日志表中
	o := orm.NewOrm()
	order_log := api.Order_log{}
	order_log.Order = orderInfo.OrderID
	order_log.Timestamp = orderInfo.CreateTime
	order_log.Agent = agent
	order_log.Account = account
	order_log.Money = money
	order_log.Type = orderInfo.Type
	order_log.Process = 0					//0:未处理
	order_log.AcTime = curTime
	if  orderInfo.Type == gameEnum.TRANS_OTHER_TO_US {
		order_log.Status = 2
	} else {
		order_log.Status = 0
	}

	id,err :=o.Insert(&order_log)
	if err == nil {
		logs.Trace("insert id:",id)
	}
}


/**
向游戏服务器请求上下分
gameAccountName:上下分的相关账号
orderId：订单编号
trans_type：上下分方向 1：api->game 2:game->api
returnMsg:接收的信息
 */
func requestGameServer(gameAccountName string,orderId string,trans_type int64, returnMsg *common.ReturnMsgStruct ) {
	orderRedisKey := gameEnum.REDIS_ORDER + orderId

	url := beego.AppConfig.String("game_pre_url")
	url += "/OrderHandler"
	url	+= "?"
	url += "account=" + gameAccountName
	url += "&orderId=" + orderId
	url += "&type=" + strconv.FormatInt(trans_type,10)
	req := httplib.Get(url)
	responseStr,err := req.String()

	if err != nil {
		returnMsg.Code = errorCode.INNER_ERROR
		returnMsg.Msg = "向游戏服务器请求数据错误"
		returnMsg.Data = map[string]interface{}{}
		return
	}


	err = json.Unmarshal([]byte(responseStr),returnMsg)
	if err != nil {
		logs.Error(responseStr)
		returnMsg.Code = errorCode.INNER_ERROR
		returnMsg.Msg = "游戏服务器过来数据格式有误"
		returnMsg.Data = map[string]interface{}{}
		return
	}

	//更新redis缓存状态
	if returnMsg.Code != errorCode.SUCC {
		cache.GetRedis().Set(orderRedisKey,gameEnum.ORDER_REDIS_STATUS_HANDLER_ERR,0)
		return
	} else {
		cache.GetRedis().Set(orderRedisKey,gameEnum.ORDER_REDIS_STATUS_HANDED,0)
	}
}
