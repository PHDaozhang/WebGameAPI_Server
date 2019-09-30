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

//订单相关
type OrderController struct {
	system.PermissionController
}


// @Title 节点列表
// @Description 节点列表
// @Success 200 {object} admin.Node
// @Param    Keyword    query    string    true  关键字
// @Param    FilterKey    query    string    true  关键字段
// @Param    Page    query    int    true  	页码
// @Param    BeginTime    query    int    true  	开始时间
// @Param    EndTime    query    int    true  	结束时间
// @Param    PageSize    query    int    true  页码数量
// @router   /list [get]
func (this *OrderController) List() {
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
	if req.FilterKey == "ChannelID" || req.FilterKey == "Account" || req.FilterKey == "OrderId" {
		query[ req.FilterKey ] = req.Keyword
	}
	if req.FilterKey == "CreateTime" || req.FilterKey == "UpdateTime" {
		registerTimeMap := bson.M{}
		registerTimeMap["$gte"] = time.Unix(req.BeginTime,0)
		registerTimeMap["$lt"] = time.Unix(req.EndTime,0)

		query[ req.FilterKey ] = registerTimeMap
	}

	OrderInfoList := []api.OrderInfo{}
	err :=  mulMongo.FindPage(gameEnum.DB_NAME_PAYMENT,gameEnum.TABLE_PAY_COLLECTION_ORDER,int(req.Page - 1),int(req.PageSize), query,bson.M{},&OrderInfoList)
	if err != nil {
		this.Error(errorCode.DB_OPER_ERROR)
	}

	this.Success(OrderInfoList)
}