package webApi

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"gopkg.in/mgo.v2/bson"
	"math/rand"
	"time"
	"tsEngine/tsCrypto"
	"tsEngine/tsDb"
	"tsEngine/tsOpCode"
	"tsEngine/tsRedis"
	"web-game-api/controllers/system"
	"web-game-api/core/consts"
	"web-game-api/logic/errorCode"
	"web-game-api/models/api"
	"web-game-api/models/dto"
)

//代理相关
type AgentController struct {
	system.PermissionController
}


// @Title 节点列表
// @Description 节点列表
// @Success 200 {object} admin.Node
// @Param    Keyword    query    string    true  关键字
// @Param    FilterKey    query    string    true  关键字段
// @Param    Page    query    int    true  	页码
// @Param    PageSize    query    int    true  页码数量
// @router   /list [get]
func (this *AgentController) List() {
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

	o := api.Agent{}

	items, pagination, err := o.List(req.Page, req.PageSize, req.Keyword,req.FilterKey)
	if err != nil {
		this.Error(tsOpCode.GET_PAGES_ERROR, err.Error())
	}

	for _, li := range items {
		data, err := json.Marshal(li)
		if err != nil {
			logs.Error(err)
		}
		_ = tsRedis.SetNX(fmt.Sprintf(consts.KeyWEBAPIAgentById, li.Id), string(data), 3600)
	}

	this.Success(bson.M{"Items":items,"Pagination":pagination})
}




// @Title 节点列表
// @Description 节点列表
// @Success 200 {object} admin.Node
// @Param    Mame    formData    string    true  代理名称
// @Param    Desc    formData    string    true  代理说明
// @router   / [post]
func (this *AgentController) Add() {
	var reqAddAgent dto.ReqAddAgent
	if err := this.ParseForm(&reqAddAgent); err != nil {
		this.Error(errorCode.PARAM_ERROR)
	}
	db := tsDb.NewDbBase()

	agent := api.Agent{
		Name:   reqAddAgent.Name,
		Desc:   reqAddAgent.Desc,
	}

	randomStr := fmt.Sprintf("%d%d",time.Now().UnixNano(),rand.Uint64())
	agent.AppKey = tsCrypto.GetMd5([]byte(randomStr))

	_,err := db.DbInsert(&agent)
	if err != nil {
		this.Error(errorCode.DB_OPER_ERROR)
	}

	this.Success(map[string]interface{}{})
}


// @Title 节点列表
// @Description 节点列表
// @Success 200 {object} admin.Node
// @Param    id    path    int    true  代理编号
// @router   /:id [delete]
func (this *AgentController) Delete() {
	o := api.Agent{}
	o.Id,_ = this.GetInt64(":id",0)

	db := tsDb.NewDbBase()
	_,err:=db.DbDel(&o)
	if err != nil {
		this.Error(errorCode.DB_OPER_ERROR)
	}

	tsRedis.Del(fmt.Sprintf(consts.KeyWEBAPIAgentById, o.Id))

	this.Success(map[string]interface{}{})
}


// @Title 返回某个代理的具体信息
// @Description 节点列表
// @Success 200 {object} admin.Node
// @Param    id    path    int    true  代理编号
// @router   /:id [get]
func (this *AgentController) Get() {
	o := api.Agent{}
	o.Id,_ = this.GetInt64(":id",0)

	redisKey := fmt.Sprintf(consts.KeyWEBAPIAgentById, o.Id)
	data,_ := tsRedis.Get( redisKey )
	if data != "" {

		if err := json.Unmarshal([]byte(data),&o); err != nil {
			logs.Error(err)
		}
		this.Success(o);
	}

	db := tsDb.NewDbBase()
	if err := db.DbGet(&o); err != nil {
		this.Error(errorCode.DB_OPER_ERROR)
	}

	sData,err := json.Marshal(o)
	if err != nil {
		logs.Error(err)
	}
	//更新redis值
	tsRedis.Set( redisKey,string(sData),3600 )

	this.Success(o)
}

// @Title 编辑更新某个代理信息，但是代理的appKey不会更新，只会更新名字与说明
// @Description 节点列表
// @Success 200 {object} admin.Node
// @Param    Id    formdata    int    true  代理编号
// @Param    Name    formdata    string    false  代理名字
// @Param    Desc    formdata    string    false  代理描述
// @router   / [put]
func (this *AgentController) Edit() {
	id,_ := this.GetInt64("Id",0)
	var req dto.ReqAddAgent
	if err := this.ParseForm(&req); err != nil {
		this.Error(tsOpCode.OPERATION_REQUEST_FAILED)
	}

	db := tsDb.NewDbBase()
	o := api.Agent{
		Id:     id,
		Name:   req.Name,
		Desc:   req.Desc,
	}

	if id <= 0 {
		this.Error(errorCode.PARAM_ERROR)
	}

	if o.Name == "" && o.Desc == "" {
		this.Error(errorCode.PARAM_ERROR)
		return
	}

	err := db.DbUpdate(&o,"Name","Desc")

	if err != nil {
		logs.Error(err)
		this.Error(errorCode.DB_OPER_ERROR)
	}

	sData,err := json.Marshal(o)
	if err != nil {
		logs.Error(err)
	}

	redisKey := fmt.Sprintf(consts.KeyWEBAPIAgentById, o.Id)
	tsRedis.Set(redisKey,sData,3600)
	this.Success(o)
}
