//角色管理
package system

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"gopkg.in/mgo.v2/bson"
	"tsEngine/tsDb"
	"tsEngine/tsOpCode"
	"tsEngine/tsRedis"
	"tsEngine/tsString"
	"tsEngine/tsTime"
	"web-game-api/conf"
	"web-game-api/core/consts"
	"web-game-api/models/dto"
	"web-game-api/models/sys"
)

type RoleController struct {
	PermissionController
}

// @Title 角色列表
// @Description 角色列表
// @Success 200 {"Code":200,"Data":{"Items": items, "Pagination": pagination}
// @Param    Keyword    formData    string    true   关键字
// @Param    Page    	formData    int64     true   当前页
// @Param    PageSize   formData    int64     true   页数
// @router /list [get]
func (this *RoleController) List() {

	var req dto.ReqSearch
	if err := this.ParseForm(&req); err != nil {
		this.Error(tsOpCode.OPERATION_REQUEST_FAILED)
	}

	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}
	if req.Page < 1 {
		req.Page = 1
	}

	o := sys.Role{}

	items, pagination, err := o.List(this.AdminId, req.Page, req.PageSize, req.Keyword)
	if err != nil {
		this.Error(tsOpCode.GET_PAGES_ERROR, err.Error())
	}
	for _, li := range items {
		data, err := json.Marshal(li)
		if err != nil {
			logs.Error(err)
		}
		_ = tsRedis.SetNX(fmt.Sprintf(consts.KeyWEBAPISysRoleById, li.Id), string(data), 3600)
	}
	this.Success(bson.M{"Items": items, "Pagination": pagination})
}

// @Title 角色信息
// @Description 角色信息
// @Success 200 {"Code":200,"Data":{"Role": oRole, "Mode": list}}
// @Param    id    path    int    true   角色ID
// @router /:id [get]
func (this *RoleController) Get() {
	roleId, _ := this.GetInt64(":id", 0)

	//初始化对象
	var oRole sys.Role
	oRole.Id = roleId
	db := tsDb.NewDbBase()
	oRole.Id, _ = this.GetInt64(":id", 0)
	data, _ := tsRedis.Get(fmt.Sprintf(consts.KeyWEBAPISysRoleById, oRole.Id))
	if data != "" {
		err := json.Unmarshal([]byte(data), &oRole)
		if err != nil {
			logs.Error(err)
		}
		this.Success(oRole)
	}
	if err := db.DbGet(&oRole); err != nil {
		this.Error(tsOpCode.OPERATION_DB_FAILED)
	}
	var oMode sys.Mode
	//通过偏移量获取数据
	list, err := db.DbList(&oMode)
	if err != nil {
		this.Error(tsOpCode.OPERATION_DB_FAILED)
	}
	sData, err := json.Marshal(oRole)
	if err != nil {
		logs.Error(err)
	}
	_ = tsRedis.Set(fmt.Sprintf(consts.KeyWEBAPISysRoleById, oRole.Id), string(sData), 3600)
	this.Success(bson.M{"Role": oRole, "Mode": list})
}

// @Title 新建角色
// @Description 新建角色
// @Success 200 {"Code":200,"Data":"success"}
// @Param    Name    		formData    string    true   名称
// @Param    Permission     formData    string    true   权限
// @Param    Description    formData    string    true   描述
// @router / [post]
//
func (this *RoleController) Add() {
	req := struct {
		Name        string
		Permission  string
		Node        string
		Description string
	}{}
	if err := this.ParseForm(&req); err != nil {
		this.Error(tsOpCode.OPERATION_REQUEST_FAILED)
	}
	db := tsDb.NewDbBase()
	adminUid := conf.SystemAdminId
	o := sys.Role{}
	pass, err := o.VerifyRolePermission(this.AdminId, req.Permission)
	if err != nil {
		logs.Error("[Role][Add]VerifyRolePermission DBError: ", err)
		this.Error(tsOpCode.OPERATION_DB_FAILED)
	}
	if !pass {
		logs.Error("[Role][Add]VerifyRolePermission Not Pass")
		this.Error(tsOpCode.DONT_COVER_YOUR_PROMISS)
	}
	o.AdminId = this.RealAdminId
	o.CreateAdminId = this.AdminId
	o.Type = 2
	if this.RealAdminId == adminUid {
		o.Type = 1
	}
	o.Name = req.Name
	o.Permission = req.Permission
	o.Description = req.Description
	o.Node = req.Node
	o.CreateTime = tsTime.CurrSe()

	_, err = db.DbInsert(&o)
	if err != nil {
		logs.Error("[Role][Add]Create Role DBError: ", err)
		this.Error(tsOpCode.OPERATION_DB_FAILED)
	}
	data, err := json.Marshal(o)
	if err != nil {
		logs.Error(err)
	}
	admin := sys.Admin{}
	err = admin.AppendCreateRole(this.AdminId, tsString.FromInt64(o.Id))
	if err != nil {
		logs.Error("[Role][Add]Add Create Role DBError: ", err)
		this.Error(tsOpCode.OPERATION_DB_FAILED)
	}
	_ = tsRedis.Set(fmt.Sprintf(consts.KeyWEBAPISysRoleById, o.Id), string(data), 3600)
	this.Success("success")
}

// @Title 角色编辑
// @Description 角色编辑
// @Success 200 {"Code":200,"Data":"success"}
// @Param    Id    			formData    int64     true   角色ID
// @Param    Name    		formData    string    true   角色名称
// @Param    Permission     formData    string    true   权限
// @Param    Description    formData    string    true   描述
// @router / [put]
//
func (this *RoleController) Edit() {
	req := struct {
		Id          int64
		Name        string
		Permission  string
		Description string
		Node        string
		System      string
	}{}
	if err := this.ParseForm(&req); err != nil {
		this.Error(tsOpCode.OPERATION_REQUEST_FAILED)
	}
	//初始化
	db := tsDb.NewDbBase()
	o := sys.Role{}
	pass, err := o.VerifyRolePermission(this.AdminId, req.Permission)
	if err != nil {
		logs.Error("[Role][Add]VerifyRolePermission DBError: ", err)
		this.Error(tsOpCode.OPERATION_DB_FAILED)
	}
	if !pass {
		logs.Error("[Role][Add]VerifyRolePermission Not Pass")
		this.Error(tsOpCode.DONT_COVER_YOUR_PROMISS)
	}
	o.Id = req.Id
	o.Type = 2
	o.AdminId = this.RealAdminId
	adminUid := conf.SystemAdminId
	if this.RealAdminId != adminUid {
		err := db.DbGet(&o, "Id", "AdminId", "Type")
		if err != nil {
			logs.Error("[Role][Edit]不可修改非自建角色")
			this.Error(tsOpCode.NOT_SELF_ROLES)
		}
	}
	o.Name = req.Name
	o.Permission = req.Permission
	o.Description = req.Description
	o.Node = req.Node
	o.System = req.System
	o.UpdateTime = tsTime.CurrSe()
	err = db.DbUpdate(&o, "Name", "Permission", "Description", "Node", "System", "UpdateTime")
	if err != nil {
		this.Error(tsOpCode.OPERATION_DB_FAILED)
	}
	data, err := json.Marshal(o)
	if err != nil {
		logs.Error(err)
	}
	_ = tsRedis.Set(fmt.Sprintf(consts.KeyWEBAPISysRoleById, o.Id), string(data), 3600)
	this.Success("success")
}

// @Title 角色删除
// @Description 角色删除
// @Success 200 {"Code":200,"Data":"success"}
// @Param    id    path    int64    true   角色ID
// @router /:id [delete]
//
func (this *RoleController) Del() {
	//初始化对象
	var oRole sys.Role
	oRole.Id, _ = this.GetInt64(":id", 0)
	id := oRole.Id
	if oRole.Id == 0 {
		this.Error(tsOpCode.OPERATION_REQUEST_FAILED)
	}
	db := tsDb.NewDbBase()
	oRole.AdminId = this.RealAdminId
	adminUid := conf.SystemAdminId
	if this.RealAdminId != adminUid {
		err := db.DbGet(&oRole, "Id", "AdminId")
		if err != nil {
			logs.Error("[Role][Edit]不可修改非自建角色,ID:", id, " HandlerId:", this.AdminId)
			this.Error(tsOpCode.NOT_SELF_ROLES)
		}
	}

	//count := oRole.GetRoleUseCount(tsString.FromInt64(id))
	count := oRole.GetRoleUseCount2(this.AdminId,tsString.FromInt64(id))
	if count > 0 {
		logs.Error("[Role][Del]权限被占用无法删除")
		this.Error(tsOpCode.EXIST_CHILD_DEPENDENCY)
	}
	oRole.Deleted = 1
	err := tsDb.NewDbBase().DbUpdate(&oRole, "Deleted")
	if err != nil {
		logs.Error("[Role][Del]DbDel error:", err)
		this.Error(tsOpCode.OPERATION_DB_FAILED)
	}

	_ = tsRedis.Del(fmt.Sprintf(consts.KeyWEBAPISysRoleById, id))
	this.Success("success")
}
