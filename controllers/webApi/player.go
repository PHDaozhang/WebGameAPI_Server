package webApi

import (
	"gopkg.in/mgo.v2/bson"
	"time"
	"tsEngine/tsOpCode"
	"web-game-api/controllers/system"
	"web-game-api/core/mulMongo"
	"web-game-api/logic/errorCode"
	"web-game-api/logic/gameEnum"
	"web-game-api/models/api"
	"web-game-api/models/dto"
)

type PlayerController struct {
	system.PermissionController
}


// @Title 人物列表
// @Description 节点列表
// @Success 200 {object} admin.Node
// @Param    Keyword    query    string    true  关键字
// @Param    FilterKey    query    string    true  关键字段
// @Param    Page    query    int    true  	页码
// @Param    PageSize    query    int    true  页码数量
// @router   /list [get]
func (this *PlayerController) List() {
	var req dto.ReqSearch
	if err := this.ParseForm(&req); err != nil {
		this.Error(tsOpCode.OPERATION_REQUEST_FAILED)
	}

	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}
	if req.PageSize < 1 {
		req.PageSize = 1
	}

	query := bson.M{}
	query["Account"] = bson.RegEx{Pattern:"^api_\\d+_",Options:"ig"}
	if req.FilterKey == "Agent" {
		query["ChannelId"] = req.Keyword
	} else if req.FilterKey == "Account" {
		query["Account"] = bson.RegEx{Pattern:req.Keyword,Options:"ig"}
	}

	if req.BeginTime != 0 && req.EndTime != 0 {
		registerTimeMap := bson.M{}
		registerTimeMap["$gte"] = time.Unix(req.BeginTime,0)
		registerTimeMap["$lt"] = time.Unix(req.EndTime,0)
		query["RegisterTime"] = registerTimeMap
	}


	PlayerInfoList := []api.AccountDBInfo{}
	err :=  mulMongo.FindPage(gameEnum.DB_NAME_ACCOUNT,gameEnum.ACCOUNT_COLLECTION_TABLE,int(req.Page-1),int(req.PageSize),  query,bson.M{},&PlayerInfoList)  //mulMongo.FindAllByCondition(gameEnum.DB_NAME_PAYMENT,gameEnum.TABLE_PAY_COLLECTION_ORDER,&bson.M{},&bson.M{},"",0,100,&OrderInfoList)
	if err != nil {
		this.Error(errorCode.DB_OPER_ERROR)
	}

	this.Success(PlayerInfoList)
}