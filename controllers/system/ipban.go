//Ip屏蔽管理
package system

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"gopkg.in/mgo.v2/bson"
	"tsEngine/tsDb"
	"tsEngine/tsJson"
	"tsEngine/tsOpCode"
	"tsEngine/tsRedis"
	"tsEngine/tsString"
	"tsEngine/tsTime"
	"web-game-api/core/consts"
	"web-game-api/models/sys"

	"github.com/astaxie/beego/validation"
)

type IpbanController struct {
	PermissionController
}

// @Title 屏蔽IP列表
// @Description 屏蔽IP列表
// @Success 200 {object} admin.Ipban
// @Param    Keyword    query    string    false  搜索关键词
// @Param    Page       query    int       true   页码
// @Param    PageSize   query    int       true   单页查询数据量
// @router   /list [get]
func (this *IpbanController) List() {

	var req = struct {
		Keyword  string
		Page     int64
		PageSize int64
	}{}
	if err := this.ParseForm(&req); err != nil {
		this.Error(tsOpCode.OPERATION_REQUEST_FAILED)
	}

	o := sys.Ipban{}
	items, pagination, err := o.List(req.Page, req.PageSize, req.Keyword)

	if err != nil {
		logs.Error(err)
		this.Error(tsOpCode.OPERATION_DB_FAILED)
	}
	for _, li := range items {
		data, err := json.Marshal(li)
		if err != nil {
			logs.Error(err)
		}
		_ = tsRedis.Set(fmt.Sprintf(consts.KeyWEBAPISysIpbanById, li["Id"]), string(data), 3600)
		_ = tsRedis.Set(fmt.Sprintf(consts.KeyWEBAPISysIpbanByIp, li["Ip"]), string(data), 3600)
	}
	this.Success(bson.M{"Items": items, "Pagination": pagination})
}

// @Title 添加要屏蔽的IP
// @Description 添加要屏蔽的IP
// @Success 200 {object} admin.Ipban
// @Param    Keyword       formData    string    false  搜索关键词
// @Param    Start         formData    int       true   屏蔽开始时间 yyyy-MM-dd HH:mm:ss  时间戳（秒）
// @Param    End           formData    int       true   屏蔽结束时间 yyyy-MM-dd HH:mm:ss  时间戳（秒）
// @Param    Description   formData    int       true   单页查询数据量
// @router   / [post]
func (this *IpbanController) Add() {

	var req = struct {
		Ip          string
		Start       uint64
		End         uint64
		Description string
	}{}
	if err := this.ParseForm(&req); err != nil {
		this.Error(tsOpCode.OPERATION_REQUEST_FAILED)
	}
	logs.Debug(tsJson.ToString(req))

	o := sys.Ipban{}
	o.Ip = req.Ip
	//o.Start = tsTime.StringToSe(req.Start, 1)
	//o.End = tsTime.StringToSe(req.End, 1)
	o.Start = req.Start
	o.End = req.End
	o.Description = req.Description
	o.AdminId = this.AdminId
	o.CreateTime = tsTime.CurrSe()
	o.UpdateTime = o.CreateTime
	logs.Debug(tsJson.ToString(o))

	//数据验证
	valid := validation.Validation{}
	valid.Required(o.Ip, "Ip").Message("10026")

	if valid.HasErrors() {
		// 如果有错误信息，证明验证没通过
		// 打印错误信息
		for _, err := range valid.Errors {
			logs.Debug("HasErrors:", err)
			this.Error(tsString.ToInt(err.Message))
		}
	}

	db := tsDb.NewDbBase()
	id, err := db.DbInsert(&o)
	if err != nil {
		logs.Error(err)
		this.Error(tsOpCode.OPERATION_DB_FAILED)
	}
	data, err := json.Marshal(o)
	if err != nil {
		logs.Error(err)
	}
	_ = tsRedis.Set(fmt.Sprintf(consts.KeyWEBAPISysIpbanById, id), string(data), 3600)
	_ = tsRedis.Set(fmt.Sprintf(consts.KeyWEBAPISysIpbanByIp, o.Ip), string(data), 3600)
	this.Success(o)
}

// @Title 获取屏蔽的IP详情
// @Description 获取屏蔽的IP详情
// @Success 200 {object} admin.Ipban
// @Param    id   path    int       true   id
// @router   /:id [get]
func (this *IpbanController) Get() {
	//初始化
	db := tsDb.NewDbBase()
	o := sys.Ipban{}

	//获取get数据
	o.Id, _ = this.GetInt64(":id", 0)
	data, _ := tsRedis.Get(fmt.Sprintf(consts.KeyWEBAPISysIpbanById, o.Id))
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
	_ = tsRedis.Set(fmt.Sprintf(consts.KeyWEBAPISysIpbanById, o.Id), string(sData), 3600)
	_ = tsRedis.Set(fmt.Sprintf(consts.KeyWEBAPISysIpbanByIp, o.Ip), string(sData), 3600)
	this.Success(o)
}

// @Title 编辑要屏蔽的IP
// @Description 编辑要屏蔽的IP
// @Success 200 {object} admin.Ipban
// @Param    Id            query    int          false
// @Param    Ip            query    string       true
// @Param    Description   query    int          true
// @Param    Start         query    int          true
// @Param    End           query    int          true
// @router   / [put]
func (this *IpbanController) Edit() {
	db := tsDb.NewDbBase()

	//获取post数据
	var req = struct {
		Id          int64
		Ip          string
		Start       uint64
		End         uint64
		Description string
	}{}
	if err := this.ParseForm(&req); err != nil {
		this.Error(tsOpCode.OPERATION_REQUEST_FAILED)
	}
	logs.Debug(tsJson.ToString(req))

	o := sys.Ipban{
		Id:          req.Id,
		Ip:          req.Ip,
		Description: req.Description,
		Start:       req.Start,
		End:         req.End,
		//Start:       tsTime.StringToSe(req.Start, 1),
		//End:         tsTime.StringToSe(req.End, 1),
	}
	logs.Debug(tsJson.ToString(o))

	//****************************************************
	//数据验证
	valid := validation.Validation{}

	valid.Required(o.Ip, "Ip").Message("10026")

	if valid.HasErrors() {
		// 如果有错误信息，证明验证没通过
		// 打印错误信息
		for _, err := range valid.Errors {
			logs.Debug("HasErrors:", err)
			this.Error(tsString.ToInt(err.Message))
		}
	}

	//****************************************************
	err := db.DbUpdate(&o, "IP", "Description", "Start", "End")

	if err != nil {
		logs.Error(err)
		this.Error(tsOpCode.OPERATION_REQUEST_FAILED)
	}
	data, err := json.Marshal(o)
	if err != nil {
		logs.Error(err)
	}
	_ = tsRedis.Set(fmt.Sprintf(consts.KeyWEBAPISysIpbanById, o.Id), string(data), 3600)
	_ = tsRedis.Set(fmt.Sprintf(consts.KeyWEBAPISysIpbanByIp, o.Ip), string(data), 3600)
	this.Success(o)
}

// @Title 删除要屏蔽的IP
// @Description 删除要屏蔽的IP
// @Success 200 {"Code":200,"Data":null}
// @Param    id     path    int    false  要删除的ID
// @router   /:id [delete]
func (this *IpbanController) Del() {

	o := sys.Ipban{}
	id, _ := this.GetInt64(":id")
	o.Id = id
	db := tsDb.NewDbBase()
	err := db.DbRead(&o)
	if err != nil {
		this.Error(tsOpCode.DATA_NOT_EXIST)
	}

	_, err = db.DbDel(&o)
	if err != nil {
		logs.Error(err)
		this.Error(tsOpCode.OPERATION_DB_FAILED)
	}
	_ = tsRedis.Del(fmt.Sprintf(consts.KeyWEBAPISysIpbanById, id))
	_ = tsRedis.Del(fmt.Sprintf(consts.KeyWEBAPISysIpbanByIp, o.Ip))
	this.Success(nil)
}
