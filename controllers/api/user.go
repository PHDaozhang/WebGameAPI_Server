package api

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/httplib"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/astaxie/beego/validation"
	"gopkg.in/mgo.v2/bson"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
	"tsEngine/tsCrypto"
	"web-game-api/controllers/system"
	"web-game-api/core/mulMongo"
	"web-game-api/logic/common"
	"web-game-api/logic/errorCode"
	"web-game-api/logic/gameEnum"
	"web-game-api/logic/manager"
	"web-game-api/logic/services/WorldService"
	"web-game-api/logic/utils"
	"web-game-api/models/api"
	"web-game-api/models/dto"
)

type UserController struct{
	system.BaseController
}

type ReqLoginData struct {
	Agent			int64 `valid:"Required" form:"agent"`
	Timestamp		int64 `valid:"Required" form:"timestamp"`
	Account			string `valid:"Required" form:"account"`
	HomeUrl			string `valid:"Required" form:"homeUrl" `
	Currency		int64 `form:"currency"`
	OrderId			string `form:"orderId"`
	Sign			string `valid:"Required" form:"sign"`
	Ip				string `valid:"Required" form:"ip"`
	GameId			int64 `form:"gameId"`
}

// @Title 玩家登陆
// @Description 第三方账号登陆
// @Success 1 {object} {url:,isNew:,orderCode:}
// @Param    agent    query    int    true  代理账号
// @Param    timestamp       query    int    true   时间戳 以ms为单位
// @Param    account       query    string    true   登陆的玩家账号
// @Param    homeUrl   query    string    true   	玩家点击退出后的地址
// @Param    currency  query    int    false  		玩家登陆上分值
// @Param    orderId    query    int    false  		当玩家上分值有效时，此值必须存在
// @Param    sign    query    string    true  		玩家签名
// @Param    ip    query    string    false  		登陆地址
// @Param    gameId    query    int    false  		具体登陆的游戏id
// @router   /login [get]
func (this *UserController)Login()  {
	var reqDto ReqLoginData
	if err := this.ParseForm(&reqDto); err != nil {
		logs.Error("login",err)
		this.Error(errorCode.PARAM_ERROR,"参数错误")
	}

	channelId := strconv.FormatInt(reqDto.Agent,10)

	valid := validation.Validation{}
	b,err := valid.Valid(&reqDto)
	if err != nil {
		logs.Error("login",err)
		this.Error(errorCode.PARAM_ERROR,"参数错误")
	}

	if !b {
		logs.Error("login",err)
		this.Error(errorCode.PARAM_ERROR,"参数错误")
	}

	gameId,err := this.GetInt("gameId",0)
	if err != nil {
		this.Error(errorCode.PARAM_ERROR,"gameId必须传数字或不传")
	}

	if reqDto.Currency > 0 {
		if reqDto.OrderId == "" {
			this.Error(errorCode.PARAM_ERROR,"参数错误")
		}

		//当上分值不为空时，则检查订单号是否舍法
		checkFlag :=	utils.CheckOrderId(reqDto.OrderId,reqDto.Agent,reqDto.Account)
		if !checkFlag {
			this.Error(errorCode.ORDER_REG_ERROR,"订单格式不合法")
		}
	}

	o := orm.NewOrm()
	//查看代理是否存在
	agentOrm := new(api.Agent)
	agentOrm.Id =  reqDto.Agent
	err = o.Read(agentOrm)
	if err != nil {
		this.Error(errorCode.AGENT_UNEXIST,"代理不存在")
	}

	//查看签名是否正确
	var checkSignFlag = utils.CheckSign(reqDto.Agent,reqDto.Timestamp,agentOrm.AppKey,reqDto.Sign);
	if !checkSignFlag {
		this.Error(errorCode.SIGN_ERROR,"签名不正确")
	}

	//不存在的渠道不能进入
	agentInfo := dto.AgentInfo{}
	err = mulMongo.FindOne(gameEnum.DB_NAME_CONFIG,gameEnum.TABLE_COLLECTION_AGENT_INFO,bson.M{"ChannelId":channelId},bson.M{},&agentInfo)
	if err != nil || (err != nil && err.Error() == "not found") {
		this.Error(errorCode.AGENT_UNEXIST,"代理不存在" + channelId)
	}

	//查询是否有此玩家
	accountDbInfo := new(api.AccountDBInfo)
	accountDbInfoAccountName :=  utils.GameAccount(reqDto.Agent,reqDto.Account)
	isNew := false

	err = mulMongo.FindOne(gameEnum.DB_NAME_ACCOUNT,gameEnum.ACCOUNT_COLLECTION_TABLE,&bson.M{"Account":accountDbInfoAccountName},&bson.M{},&accountDbInfo)
	if err == nil {
		//如果有此玩家，则更新玩家的 randKey与lastTime
		accountDbInfo.RandKey =  rand.Int31()
		accountDbInfo.LastTime = reqDto.Timestamp / 1000
		err =  mulMongo.Update(gameEnum.DB_NAME_ACCOUNT,gameEnum.ACCOUNT_COLLECTION_TABLE,bson.M{"Account":accountDbInfoAccountName},&accountDbInfo)
		if err != nil {
			if err != nil {
				this.Error(errorCode.DB_OPER_ERROR,"操作更新数据库失败")
			}
		}
	} else {
		//检查玩家账号是否合法，不能包含,<>:/\?
		reg := regexp.MustCompile(`[,<:/\?\\>]`)
		if reg.MatchString(reqDto.Account) {
			this.Error(errorCode.ACCOUNT_NAME_ILLIGLE,"账号名字含有非法字符")
		}

		//如果没有此玩家，则创建一个玩家，并将玩家账号插入到ACCOUNTTABLE中去
		accountDbInfo.Id_ = bson.NewObjectId()
		accountDbInfo.Account = accountDbInfoAccountName;
		accountDbInfo.LastTime = reqDto.Timestamp
		accountDbInfo.LastIp = reqDto.Ip
		accountDbInfo.ChannelId = strconv.FormatInt(reqDto.Agent,10)					//渠道号
		accountDbInfo.ClientChannelId = strconv.FormatInt(reqDto.Agent,10)			//包渠道号
		accountDbInfo.RegisterIp = reqDto.Ip
		accountDbInfo.Level = 1
		accountDbInfo.RandKey =  rand.Int31()
		accountDbInfo.RegisterTime = time.Now()
		accountDbInfo.Platform = "web"
		accountDbInfo.DevicePlatform = "web"

		err = mulMongo.Insert(gameEnum.DB_NAME_ACCOUNT,gameEnum.ACCOUNT_COLLECTION_TABLE, &accountDbInfo)

		if err != nil {
			this.Error(errorCode.DB_OPER_ERROR,"操作插入数据库失败")
		}

		isNew = true
	}

	serverInfo,ok := WorldService.GetInstance().AllocServer(accountDbInfoAccountName)
	if !ok {
		this.Error(errorCode.GATE_SERVER_ERROR)
	}

	returnMsgStruct :=common.ReturnMsgStruct{}
	if reqDto.Currency > 0 {
		TransMoney(reqDto.Agent,reqDto.Timestamp,reqDto.Account,gameEnum.TRANS_OTHER_TO_US,reqDto.OrderId,reqDto.Sign,reqDto.Currency,&returnMsgStruct)
	}

	urlConfig := manager.GetInstance().GetJsonUrlDiffer("")

	this.Code = errorCode.SUCC
	this.Msg = "ok"

	token := strconv.FormatInt(int64(accountDbInfo.RandKey),10) + ":" + strconv.FormatInt(accountDbInfo.LastTime,10)
	token = tsCrypto.GetMd5([]byte(token))
	//为与服务器的校验一致，需要转换为大写
	token = strings.ToUpper(token)
	game_url := beego.AppConfig.String("game_client_url")
	game_url += "?info=" + token
	game_url += "&acc=" + accountDbInfo.Account
	game_url += "&gameip=" + serverInfo.GateIp
	game_url += "&homeUrl=" + reqDto.HomeUrl
	game_url += "&gameId="+ strconv.Itoa(gameId)
	game_url += "&urlConfig=" +  urlConfig

	resultMap := map[string]interface{}{}
	resultMap["url"] = game_url
	resultMap["isNew"] = isNew
	resultMap["orderCode"] = returnMsgStruct.Code

	this.APISuccess(resultMap)
}

// @Title 踢玩家下线
// @Description 踢玩家下线
// @Success 200 {object} admin.Admin
// @Param    agent    query    int64    true  代理账号
// @Param    timestamp       query    int64    true   时间戳
// @Param    sign  query    string    true  签名
// @Param    account    query    string    true  被踢的账号
// @router   /kick [get]
func (this *UserController)Kick() {
	agent,_ := this.GetInt64("agent",0);
	timestamp,_ := this.GetInt64("timestamp",0);
	sign := this.GetString("sign");
	account := this.GetString("account")

	if agent == 0 || timestamp == 0 || sign == "" || account == ""{
		this.Code =  errorCode.PARAM_ERROR;
		this.Msg = "参数错误"
		this.Result = map[string]interface{}{}
		this.TraceJson();
		return;
	}

	o := orm.NewOrm()
	//查看代理是否存在
	agentOrm := new(api.Agent)
	agentOrm.Id = int64(agent)
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

	gameAccountName := utils.GameAccount(agent,account)

	url := beego.AppConfig.String("game_pre_url")
	url += "/KickUser"
	url	+= "?"
	url += "account=" + gameAccountName
	req := httplib.Get(url)
	responseStr,err := req.String()

	if err != nil {
		this.Code = errorCode.INNER_ERROR
		this.Msg = "向游戏服务器请求数据错误"
		this.Result = map[string]interface{}{}
		return
	}

	var responseBody common.ReturnMsgStruct = common.ReturnMsgStruct{}

	err = json.Unmarshal([]byte(responseStr),&responseBody)
	if err != nil {
		logs.Error(responseStr)
		this.Code = errorCode.INNER_ERROR
		this.Msg = "服务器过来数据格式有误"
		this.Result = map[string]interface{}{}
		return
	}

	this.Code = int(responseBody.Code)
	this.Msg = responseBody.Msg
	this.Result = responseBody.Data

	this.TraceJson();
}

// @Title 玩家登陆
// @Description 管理员列表
// @Success 200 {object} admin.Admin
// @Param    agent    query    string    true  代理账号
// @Param    timestamp       query    string    true   时间戳
// @Param    sign   query    string    true   签名
// @Param    account  query    string    true  被踢玩家账号
// @router   /userinfo [get]
func (this *UserController)UserInfo() {
	agent,_ := this.GetInt64("agent",0);
	timestamp,_ := this.GetInt64("timestamp",0);
	sign := this.GetString("sign");
	account := this.GetString("account")

	if agent == 0 || timestamp == 0 || sign == "" || account == ""{
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


	gameAccountName := utils.GameAccount(agent,account)

	url := beego.AppConfig.String("game_pre_url")
	url += "/UserInfo"
	url	+= "?"
	url += "account=" + gameAccountName
	req := httplib.Get(url)
	responseStr,err := req.String()

	if err != nil {
		this.Error(errorCode.INNER_ERROR,"向游戏服务器请求数据错误")
	}

	var responseBody common.ReturnMsgStruct = common.ReturnMsgStruct{}

	err = json.Unmarshal([]byte(responseStr),&responseBody)
	if err != nil {
		this.Error(errorCode.INNER_ERROR,"服务器过来数据格式有误")
	}

	resultMap := map[string]interface{}{}
	resultMap["freeMoney"] = responseBody.Data["freeMoney"]

	this.Code = int(responseBody.Code)
	this.Msg = responseBody.Msg
	this.Result = resultMap

	//当前玩家状态，还是需要问下服务器数据库字段
	this.TraceJson();
}