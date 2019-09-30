package api

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/httplib"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/astaxie/beego/validation"
	"gopkg.in/mgo.v2/bson"
	"strconv"
	"web-game-api/controllers/system"
	"web-game-api/core/mulMongo"
	"web-game-api/logic/errorCode"
	"web-game-api/logic/gameEnum"
	"web-game-api/models/api"
)

type HistoryController struct {
	system.BaseController
}

type reqSearchDto struct {

}

///admin/operationgamedetail/gamedetail 获取游戏列表
// @Title 游戏详情列表
// @Description 游戏详情列表
// @Param    Agent      		query    	 int    true  代理id
// @Param    Account      		query    	 string    false  游戏ID
// @Param    BeginTime    		query   	 int       false  开始时间
// @Param    EndTime      		query    	 int       false  结束时间
// @Param    GameId      		query    	 string    false  游戏ID
// @Param    Page         		query    	 string    true   页码"
// @Param    PageSize     		query    	 string    true   单页数据量
// @Success 200 {object} beego.M
// @router /gamedetail [get]
func (this *HistoryController)GameDetail(){
	req := struct {
		Agent 			int		`valid:"Required"`
		Account			string
		BeginTime		int
		EndTime			int
		GameId			int
		Page			int	`valid:"Required"`
		PageSize		int	`valid:"Required"`
	}{}

	if err := this.ParseForm(&req); err != nil {
		this.Error(errorCode.PARAM_ERROR)
	}

	valid := validation.Validation{}
	b,err := valid.Valid(&req)
	if err != nil || !b {
		logs.Error("login",err)
		this.Error(errorCode.PARAM_ERROR,"参数错误")
	}

	resultDto := struct {
		Account				string
		PlayerId			int
	}{}
	if req.Account != "" {
		//根据account查询玩家的playerId
		query := bson.M{"Account":req.Account}
		selector := bson.M{}
		err = mulMongo.FindOne(gameEnum.DB_NAME_PLAYER,gameEnum.TABKE_COLLECTION_PLAYERINFO,&query,selector,&resultDto)
		if err != nil {
			this.Error(errorCode.DB_OPER_ERROR,"查询mongo数据库玩家信息报错")
		}
	}

	//查看代理是否存在
	o := orm.NewOrm()
	agentOrm := new(api.Agent)
	agentOrm.Id = int64(req.Agent)
	err = o.Read(agentOrm)
	if err != nil {
		this.Msg = "代理不存在"
		this.Code = errorCode.AGENT_UNEXIST
		this.Result = map[string]interface{}{}
		this.TraceJson()
		return
	}

	url := beego.AppConfig.String("game_cloud_url")
	url += "/api/operationgamedetail/gamedetail"					//这里可能会不一样，需要与云平台对比下
	url	+= "?"
	url += "Page=" + strconv.Itoa(req.Page)  + "&PageSize=" + strconv.Itoa(req.PageSize)
	if resultDto.PlayerId > 0 {
		url += "&PlayerId=" + strconv.Itoa(resultDto.PlayerId)
	}
	if req.GameId > 0 {
		url +=  "&GameId=" + strconv.Itoa(req.GameId)
	}
	if req.BeginTime > 0 {
		url +=  "&BeginTime=" + strconv.Itoa(req.BeginTime)
	}
	if req.BeginTime > 0 {
		url +=  "&EndTime=" + strconv.Itoa(req.EndTime)
	}
	if req.Agent > 0 {
		url +=  "&Agent=" + strconv.Itoa(req.Agent)
	}

	httpReq := httplib.Get(url)
	responseStr,err := httpReq.String()

	if err != nil {
		this.Code = errorCode.INNER_ERROR
		this.Msg = "向云平台服务器请求数据错误"
		this.Result = map[string]interface{}{}
		return
	}

	this.Success(responseStr)
}

// @Title 游戏详情 - 查询
// @Description 游戏详情 - 查询
// @Param    GameId      		query    	 string    false  游戏ID
// @Param    DetailId      		query    	 string    true  关联游戏数据ID
// @Param    SelfDetailId      	query    	 string    false  关联游戏用户数据ID
// @Success 200 {object} beego.M
func (this *HistoryController)Onegamedetail(){
	//req := struct {
	//	GameId 				string
	//	DetailId			string
	//	SelfDetailId		string
	//}{}
}