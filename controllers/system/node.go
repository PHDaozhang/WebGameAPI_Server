//节点管理
package system

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"tsEngine/tsDb"
	"tsEngine/tsJson"
	"tsEngine/tsOpCode"
	"tsEngine/tsRedis"
	"web-game-api/core/consts"
	"web-game-api/models/dto"
	"web-game-api/models/sys"
)

//客户端节点配置
type NodeController struct {
	PermissionController
}

// @Title 节点列表
// @Description 节点列表
// @Success 200 {object} admin.Node
// @router   /list [get]
func (this *NodeController) List() {

	o := sys.Node{}
	db := tsDb.NewDbBase()
	order := []string{"Sort", "Id"}
	list, _ := db.DbListOrder(&o, order)
	for _, li := range list {
		data, err := json.Marshal(li)
		if err != nil {
			logs.Error(err)
		}
		_ = tsRedis.SetNX(fmt.Sprintf(consts.KeyWEBAPISysNodeById, li["Id"]), string(data), 3600)
	}
	this.Success(list)
}

// @Title 添加节点
// @Description 添加节点
// @Success 200 {object} admin.Node
// @Param    Name         formData    string    true    节点名称
// @Param    Url          formData    string    true    跳转地址
// @Param    Icon         formData    string    false   图标
// @Param    Sort         formData    int       false   排序
// @Param    LangCn       formData    string    true    简体中文
// @Param    LangTw       formData    string    false   繁体中文
// @Param    LangUs       formData    string    false   英语
// @Param    Description  formData    string    false   描述
// @Param    ParentId     formData    int       false   上级
// @router   / [post]
func (this *NodeController) Add() {

	var req dto.ReqAddNode
	if err := this.ParseForm(&req); err != nil {
		this.Error(tsOpCode.OPERATION_REQUEST_FAILED)
	}
	logs.Debug(tsJson.ToString(req))

	//获取post数据
	o := sys.Node{
		Name:        req.Name,
		Url:         req.Url,
		Icon:        req.Icon,
		Sort:        req.Sort,
		LangCn:      req.LangCn,
		LangTw:      req.LangTw,
		LangUs:      req.LangUs,
		Description: req.Description,
		ParentId:    req.ParentId,
	}
	if o.Name == "" {
		this.Error(tsOpCode.OPERATION_REQUEST_FAILED)
	}

	if o.ParentId == 0 {
		o.ParentId = -1
	}

	o.ParentTree = ",-1,"

	db := tsDb.NewDbBase()
	db.Transaction()
	defer db.TransactionEnd()

	if o.ParentId > 0 {
		var oNode sys.Node
		oNode.Id = o.ParentId
		db.DbRead(&oNode)
		o.ParentTree = oNode.ParentTree
	}
	_, err := db.DbInsert(&o)
	if err != nil {
		db.SetRollback(true)
		logs.Error(err)
		this.Code = tsOpCode.OPERATION_DB_FAILED
	}

	//更新ParentTree，链接自身的Id
	o.ParentTree += fmt.Sprintf("%d,", o.Id)
	err = db.DbUpdate(&o, "ParentTree")
	if err != nil {
		db.SetRollback(true)
		logs.Error(err)
		this.Code = tsOpCode.OPERATION_DB_FAILED
	}

	this.Success(nil)
}

// @Title 获取节点详情
// @Description 获取节点详情
// @Success 200 {object} admin.Node
// @Param    id   path    int       true   id
// @router   /:id [get]
func (this *NodeController) Get() {

	//初始化
	db := tsDb.NewDbBase()
	o := sys.Node{}

	//获取get数据
	o.Id, _ = this.GetInt64(":id", 0)
	data, _ := tsRedis.Get(fmt.Sprintf(consts.KeyWEBAPISysNodeById, o.Id))
	if data != "" {
		err := json.Unmarshal([]byte(data), &o)
		if err != nil {
			logs.Error(err)
		}
		this.Success(o)
	}
	err := db.DbGet(&o)
	if err != nil {
		logs.Error(err)
		this.Error(tsOpCode.DATA_NOT_EXIST)
	}
	sData, err := json.Marshal(o)
	if err != nil {
		logs.Error(err)
	}
	_ = tsRedis.Set(fmt.Sprintf(consts.KeyWEBAPISysNodeById, o.Id), string(sData), 3600)
	this.Success(o)
}

// @Title 编辑节点
// @Description 编辑节点
// @Success 200 {object} admin.Node
// @Param    Id           formData    int       true    主键
// @Param    Name         formData    string    true    节点名称
// @Param    Url          formData    string    true    跳转地址
// @Param    Icon         formData    string    false   图标
// @Param    Sort         formData    int       false   排序
// @Param    LangCn       formData    string    true    简体中文
// @Param    LangTw       formData    string    false   繁体中文
// @Param    LangUs       formData    string    false   英语
// @Param    Description  formData    string    false   描述
// @Param    ParentId     formData    int       false   上级
// @router   / [put]
func (this *NodeController) Edit() {
	// 获取post数据
	id, _ := this.GetInt64("Id", 0)
	var req dto.ReqAddNode
	if err := this.ParseForm(&req); err != nil {
		this.Error(tsOpCode.OPERATION_REQUEST_FAILED)
	}

	//初始化
	db := tsDb.NewDbBase()
	o := sys.Node{
		Id:          id,
		Name:        req.Name,
		Url:         req.Url,
		Icon:        req.Icon,
		Sort:        req.Sort,
		LangCn:      req.LangCn,
		LangTw:      req.LangTw,
		LangUs:      req.LangUs,
		Description: req.Description,
	}

	if o.Id <= 0 || len(o.Name) == 0 || len(o.Url) == 0 {
		this.Error(tsOpCode.OPERATION_REQUEST_FAILED)
	}

	err := db.DbUpdate(&o, "Name", "Url", "Icon", "Sort", "LangCn", "LangTw", "LangUs", "Description")

	if err != nil {
		logs.Error(err)
		this.Error(tsOpCode.OPERATION_DB_FAILED)
	}
	sData, err := json.Marshal(o)
	if err != nil {
		logs.Error(err)
	}
	_ = tsRedis.Set(fmt.Sprintf(consts.KeyWEBAPISysNodeById, o.Id), string(sData), 3600)
	this.Success(o)
}

// @Title 删除节点
// @Description 删除节点
// @Success 200 {"Code":200,"Data":null}
// @Param    id     path    int    false  要删除的ID
// @router   /:id [delete]
func (this *NodeController) Del() {

	db := tsDb.NewDbBase()
	o := sys.Node{}
	o.Id, _ = this.GetInt64(":id")

	logs.Debug("delete node:", o.Id)

	_, err := db.DbDel(&o, "ParentTree__icontains", fmt.Sprintf(",%d,", o.Id))

	if err != nil {
		logs.Error(err)
		this.Error(tsOpCode.OPERATION_DB_FAILED)
	}
	_ = tsRedis.Del(fmt.Sprintf(consts.KeyWEBAPISysNodeById, o.Id))
	this.Success(nil)
}
