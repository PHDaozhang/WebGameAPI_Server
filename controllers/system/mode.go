//管理员管理
package system

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"strings"
	"tsEngine/tsCrypto"
	"tsEngine/tsDb"
	"tsEngine/tsJson"
	"tsEngine/tsOpCode"
	"tsEngine/tsRedis"
	"tsEngine/tsString"
	"web-game-api/core/consts"
	"web-game-api/models/dto"
	"web-game-api/models/sys"

	"github.com/astaxie/beego/validation"
)

type ModeController struct {
	PermissionController
}

// @Title 模块列表
// @Description 模块列表
// @Success 200 {object} admin.Mode
// @Param    Keyword    query    string    false  搜索词
// @router   /list [get]
func (this *ModeController) List() {

	var req dto.ReqSearch
	if err := this.ParseForm(&req); err != nil {
		this.Error(tsOpCode.OPERATION_REQUEST_FAILED)
	}

	o := sys.Mode{}
	items, err := o.List(req.Keyword)

	if err != nil {
		logs.Error(err)
		this.Error(tsOpCode.OPERATION_DB_FAILED)
	}
	for _, li := range items {
		data, err := json.Marshal(li)
		if err != nil {
			logs.Error(err)
		}
		_ = tsRedis.SetNX(fmt.Sprintf(consts.KeyWEBAPISysModeById, li["Id"]), string(data), 3600)
		_ = tsRedis.SetNX(fmt.Sprintf(consts.KeyWEBAPISysModeByMD5, li["Md5"]), string(data), 3600)
	}
	this.Success(items)
}

// @Title 添加模块
// @Description 添加模块
// @Success 200 {object} admin.Mode
// @Param    Name           formData    string    true   名称
// @Param    Type           formData    int       true   类型
// @Param    Key            formData    string    true   关键词
// @Param    ParentId       formData    int       false  父ID
// @Param    NodeId         formData    int       true   节点ID
// @Param    Logs           formData    int       true
// @Param    Description    formData    string    true   描述
// @router   / [post]
func (this *ModeController) Add() {
	this.CheckRoot() //校验是否为开发者

	var req dto.ReqAddModel
	if err := this.ParseForm(&req); err != nil {
		this.Error(tsOpCode.OPERATION_REQUEST_FAILED)
	}

	o := sys.Mode{
		Name:        req.Name,
		Type:        req.Type,
		Key:         strings.ToLower(req.Key),
		ParentId:    req.ParentId,
		Logs:        req.Logs,
		Description: req.Description,
		NodeId:      req.NodeId,
	}
	if o.Type == 0 {
		o.Type = 1
	}
	if o.ParentId == 0 {
		o.ParentId = -1
	}
	o.Md5 = tsCrypto.GetMd5([]byte(o.Key))

	//****************************************************
	//数据验证
	valid := validation.Validation{}
	valid.Range(int(o.Type), 1, 3, "Type").Message("10030")
	valid.Required(o.Name, "Name").Message("10031")
	valid.MaxSize(o.Name, 200, "NameMax").Message("10032")

	if valid.HasErrors() {
		// 如果有错误信息，证明验证没通过
		// 打印错误信息
		for _, err := range valid.Errors {
			this.Error(tsString.ToInt(err.Message))
		}
	}

	db := tsDb.NewDbBase()
	db.Transaction()
	defer db.TransactionEnd()

	var pData sys.Mode
	if o.ParentId > 0 {
		pData.Id = o.ParentId
		db.DbRead(&pData)
		o.Md5 = tsCrypto.GetMd5([]byte(pData.Key + o.Key))
		o.Type = 3
	}

	_, err := db.DbInsert(&o)
	if err != nil {
		db.SetRollback(true)
		logs.Error(err)
		this.Error(tsOpCode.OPERATION_DB_FAILED)
	}

	if o.ParentId == -1 {

		var data1 sys.Mode
		data1.Name = "添加"
		data1.Type = 3
		data1.Key = "add"
		data1.Md5 = tsCrypto.GetMd5([]byte(o.Key + "add"))
		data1.Logs = 1
		data1.ParentId = o.Id

		var data2 sys.Mode
		data2.Name = "编辑"
		data2.Type = 3
		data2.Key = "edit"
		data2.Md5 = tsCrypto.GetMd5([]byte(o.Key + "edit"))
		data2.Logs = 1
		data2.ParentId = o.Id

		var data3 sys.Mode
		data3.Name = "删除"
		data3.Type = 3
		data3.Key = "del"
		data3.Md5 = tsCrypto.GetMd5([]byte(o.Key + "del"))
		data3.Logs = 1
		data3.ParentId = o.Id

		var data4 sys.Mode
		data4.Name = "列表"
		data4.Type = 3
		data4.Key = "list"
		data4.Md5 = tsCrypto.GetMd5([]byte(o.Key + "list"))
		data4.Logs = 2
		data4.ParentId = o.Id

		var data5 sys.Mode
		data5.Name = "查看"
		data5.Type = 3
		data5.Key = "get"
		data5.Md5 = tsCrypto.GetMd5([]byte(o.Key + "get"))
		data5.Logs = 2
		data5.ParentId = o.Id

		var temp []*sys.Mode
		temp = append(temp, &data1)
		temp = append(temp, &data2)
		temp = append(temp, &data3)
		temp = append(temp, &data4)
		temp = append(temp, &data5)

		err = db.DbInsertMulti(&temp, 1)
		if err != nil {
			db.SetRollback(true)
			logs.Error(err)
			this.Error(tsOpCode.OPERATION_DB_FAILED)
		}
		for _, tmp := range temp {
			data, err := json.Marshal(tmp)
			if err != nil {
				logs.Error(err)
			}
			_ = tsRedis.Set(fmt.Sprintf(consts.KeyWEBAPISysModeById, tmp.Id), string(data), 3600)
			_ = tsRedis.Set(fmt.Sprintf(consts.KeyWEBAPISysModeByMD5, tmp.Md5), string(data), 3600)
		}
	}
	data, err := json.Marshal(o)
	if err != nil {
		logs.Error(err)
	}
	_ = tsRedis.Set(fmt.Sprintf(consts.KeyWEBAPISysModeById, o.Id), string(data), 3600)
	_ = tsRedis.Set(fmt.Sprintf(consts.KeyWEBAPISysModeByMD5, o.Md5), string(data), 3600)
	this.Success(nil)
}

// @Title 模块编辑
// @Description 模块编辑
// @Success 200 {object} admin.Mode
// @Param    id   path    int       true   id
// @router   /:id [get]
func (this *ModeController) Get() {
	this.CheckRoot() //校验是否为开发者

	//初始化
	db := tsDb.NewDbBase()
	o := sys.Mode{}

	//获取get数据
	o.Id, _ = this.GetInt64(":id", 0)
	data, _ := tsRedis.Get(fmt.Sprintf(consts.KeyWEBAPISysModeById, o.Id))
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
		this.Error(tsOpCode.NO_PERMISSION_UPDATE_CHILD)
	}
	sData, err := json.Marshal(o)
	if err != nil {
		logs.Error(err)
	}
	_ = tsRedis.Set(fmt.Sprintf(consts.KeyWEBAPISysModeById, o.Id), string(sData), 3600)
	_ = tsRedis.Set(fmt.Sprintf(consts.KeyWEBAPISysModeByMD5, o.Md5), string(sData), 3600)
	this.Success(o)
}

// @Title 模块编辑
// @Description 模块编辑
// @Success 200 {object} admin.Mode
// @Param    Id             formData    int       true   id
// @Param    Name           formData    string    true   名称
// @Param    Type           formData    string    true   类型
// @Param    Key            formData    string    true   主键
// @Param    ParentId       formData    string    true   上级
// @Param    Logs           formData    int       false
// @Param    Description    formData    string    true
// @Param    Sort           formData    string    true
// @Param    NodeId         formData    int64     true
// @router   / [put]
func (this *ModeController) Edit() {

	id, _ := this.GetInt64("Id", 0)
	var req dto.ReqAddModel
	if err := this.ParseForm(&req); err != nil {
		this.Error(tsOpCode.OPERATION_REQUEST_FAILED)
	}
	logs.Debug(tsJson.ToString(req))

	db := tsDb.NewDbBase()
	o := sys.Mode{
		Id: id,
	}

	//获取post数据
	err := db.DbGet(&o)
	if err != nil {
		logs.Error(err)
		this.Error(tsOpCode.DATA_NOT_EXIST)
	}

	o.Name = req.Name
	o.Type = req.Type
	o.Key = strings.ToLower(req.Key)
	o.ParentId = req.ParentId
	o.Logs = req.Logs
	o.Description = req.Description
	o.Sort = req.Sort
	o.NodeId = req.NodeId

	if o.ParentId == 0 {
		o.ParentId = -1
	}
	o.Md5 = tsCrypto.GetMd5([]byte(o.Key))

	//****************************************************
	//数据验证
	valid := validation.Validation{}
	valid.Range(int(o.Type), 1, 3, "Type").Message("10030")
	valid.Required(o.Name, "Name").Message("10031")
	valid.MaxSize(o.Name, 200, "NameMax").Message("10032")

	if valid.HasErrors() {
		// 如果有错误信息，证明验证没通过
		// 打印错误信息
		for _, err := range valid.Errors {
			this.Error(tsString.ToInt(err.Message))
		}

	}

	var pData sys.Mode
	if o.ParentId > 0 {
		pData.Id = o.ParentId
		db.DbRead(&pData)
		o.Md5 = tsCrypto.GetMd5([]byte(pData.Key + o.Key))
		o.Type = 3
	}

	err = db.DbUpdate(&o, "Name", "Type", "Key", "Md5", "Logs", "Description", "Sort", "NodeId")
	if err != nil {
		logs.Error(err)
		this.Error(tsOpCode.OPERATION_DB_FAILED)
	}
	sData, err := json.Marshal(o)
	if err != nil {
		logs.Error(err)
	}
	_ = tsRedis.Set(fmt.Sprintf(consts.KeyWEBAPISysModeById, o.Id), string(sData), 3600)
	_ = tsRedis.Set(fmt.Sprintf(consts.KeyWEBAPISysModeByMD5, o.Md5), string(sData), 3600)
	this.Success(nil)
}

// @Title 删除模块
// @Description 删除模块
// @Success 200 {object} admin.Mode
// @Param    id          path    int    true   管理员ID
// @router   /:id [delete]
func (this *ModeController) Del() {
	this.CheckRoot() //校验是否为开发者

	db := tsDb.NewDbBase()
	db.Transaction()
	defer db.TransactionEnd()

	o := sys.Mode{}
	o.Id, _ = this.GetInt64(":id")
	if o.Id <= 0 {
		this.Error(tsOpCode.OPERATION_REQUEST_FAILED)
	}

	err := db.DbRead(&o)
	if err != nil {
		this.Error(tsOpCode.OPERATION_REQUEST_FAILED)
	}

	//先删除所有子节点
	list, err := db.DbList(&o, "ParentId", o.Id)
	for _, li := range list {
		_ = tsRedis.Del(fmt.Sprintf(consts.KeyWEBAPISysModeById, li["Id"]))
		_ = tsRedis.Del(fmt.Sprintf(consts.KeyWEBAPISysModeByMD5, li["Md5"]))
	}
	_, err = db.DbDel(&o, "ParentId", o.Id)
	if err != nil {
		db.SetRollback(true)
		logs.Error(err)
		this.Error(tsOpCode.OPERATION_DB_FAILED)
	}
	//再删除主节点
	id := o.Id
	_, err = db.DbDel(&o)
	if err != nil {
		db.SetRollback(true)
		logs.Error(err)
		this.Error(tsOpCode.OPERATION_DB_FAILED)
	}
	_ = tsRedis.Del(fmt.Sprintf(consts.KeyWEBAPISysModeById, id))
	_ = tsRedis.Del(fmt.Sprintf(consts.KeyWEBAPISysModeByMD5, o.Md5))
	this.Success(nil)
}
