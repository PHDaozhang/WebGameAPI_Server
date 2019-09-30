package admin

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"tsEngine/tsContain"
	"tsEngine/tsDb"
	"tsEngine/tsOpCode"
	"tsEngine/tsString"
	"web-game-api/controllers/system"
	"web-game-api/core/utility"
	"web-game-api/models/admin"
	"web-game-api/models/dto"
	"web-game-api/models/sys"
)


type SonUserController struct {
	system.PermissionController
}

func (this *SonUserController) Prepare() {
	// 权限判断
	this.CheckPermission()
	//this.CheckLogin()
}

// @Title 获取列表
// @Description 获取列表
// @Success 200 {object} respSonList
// @Param    Keyword    query    string    false  搜索词
// @Param    Page       query    string    true   页码
// @Param    PageSize   query    string    true   单页数据量
// @Param    BeginTime  query    int       false  过滤开始时间
// @Param    EndTime    query    int       false  过滤结束时间
// @Param    Role	    query    string    false  角色
// @Param    Status     query    int       false  状态
// @router /list [get]
func (this *SonUserController) List() {
	var req dto.ReqSearch
	if err := this.ParseForm(&req); err != nil {
		logs.Debug("ParseForm:", err)
		this.Error(tsOpCode.OPERATION_REQUEST_FAILED)
	}
	req.SetDefault()

	tAdmin := sys.Admin{}
	tPolicy := admin.DataChannelPolicy{}

	qbCount, _ := orm.NewQueryBuilder("mysql")
	qbList, _ := orm.NewQueryBuilder("mysql")
	qbCount.Select("count(id) count").From(tAdmin.TableName()).Where("admin_type=5").And("deleted=0")
	qbList.Select("A.id,A.name,A.app_remark,A.app_account,A.mobile,A.policy_id,A.create_time,B.name policy_name").
		From(tAdmin.TableName() + " A").
		LeftJoin(tPolicy.TableName() + " B").On("A.policy_id=B.id").
		Where("A.admin_type=5").And("A.deleted=0").OrderBy(utility.FormatSort(req.Sort, "")).
		Limit(int(req.PageSize)).Offset(int(req.PageSize * req.Page))

	dbOrm := orm.NewOrm()
	var count respCount
	var data []respSonList
	if err := dbOrm.Raw(qbCount.String()).QueryRow(&count); err != nil {
		this.Error(tsOpCode.DB_SELECT_ERROR, err.Error())
	}

	if _, err := dbOrm.Raw(qbList.String()).QueryRows(&data); err != nil {
		this.Error(tsOpCode.DB_SELECT_ERROR, err.Error())
	}

	this.Success(beego.M{"count": count.Count, "list": data})
}

// @Title 新增子账号
// @Description 新增子账号
// @Success 200 {object} models.Account
// @Param    Username         formData    string    true   登陆名
// @Param    Password         formData    string    true   用户密码
// @Param    ConfirmPassword  formData    string    true   用户确认密码
// @Param    Status           formData    int       false  状态
// @Param    Role             formData    string    true   角色
// @Param    CreateRole       formData    string    true   可创建角色列表
// @Param    Name             formData    string    true   名称
// @Param    Mobile           formData    string    true   联系电话
// @Param    ContactInf       formData    string    true   联系方式
// @router   / [post]
func (this *SonUserController) Add() {
	var req dto.ReqAddAdmin
	if err := this.ParseForm(&req); err != nil {
		logs.Debug("ParseForm:", err)
		this.Error(tsOpCode.OPERATION_REQUEST_FAILED)
	}

	// 默认为子账号
	req.AdminType = 4
	req.Status = 1
	//DONE 验证操作者身上是否有该权限
	mHandle := sys.Admin{}
	mHandle.Id = this.AdminId
	err := tsDb.NewDbBase().DbGet(&mHandle)
	if err != nil {
		logs.Error("[SonUser][获取账号角色]:DBError: ", err)
		this.Error(tsOpCode.OPERATION_DB_FAILED)
	}
	//TODO 是否需要角色判断
	roleStr := mHandle.CreateRole
	roleArr := tsString.CoverStringToArray(roleStr, ",", false)
	reqRoleArr := tsString.CoverStringToArray(req.CreateRole, ",", false)
	for _, reqRole := range reqRoleArr {
		if !tsContain.InArrayString(roleArr, reqRole) {
			logs.Error("[SonUser][检查自身角色] 自身不拥有该角色：SelfRole:", roleArr, " ReqRole:", reqRoleArr)
			this.Error(tsOpCode.CREATE_ROLE_CHILD_DENIDE)
		}
	}
	// 初始化
	mA := sys.Admin{}
	code, _ := mA.InitAdmin(req, this.PersonInfo, true, true)
	if this.Code != tsOpCode.OPERATION_SUCCESS {
		this.Error(int(code))
	}
	//handleLog := models.HandleLogs{}
	//handleLog.Log(consts.LogTemplateEmployeeCreate, this.AgentId, this.AdminId, this.AdminUsername, this.Ip, beego.M{"Username": req.Username})
	this.Success("Success")
}


// @Title 获取当前用户详情
// @Description 获取当前用户详情
// @Success 200 {object} admin.AdminAccount
// @Param    Id         	  formData    int    	true   用户ID
// @router   /:id [get]
func (this *SonUserController) Get() {
	o := sys.Admin{}

	id := tsString.ToInt64(this.Ctx.Input.Param(":id"))
	logs.Trace(id)

	if id <= 0 {
		this.Error(tsOpCode.DATA_NOT_EXIST)
	}

	adminAccountInfo, err := o.GetAdminById(id, this.RealAdminId)
	if err != nil {
		this.Error(tsOpCode.DATA_NOT_EXIST)
	}

	this.Success(adminAccountInfo)
}

// @Title 修改客户关联策略组
// @Description 编辑
// @Success 200 {nil}
// @Param    token          header      string    true   "auth token"
// @Param    Id             path        int       true   ""
// @Param    PolicyId       formData    int       true   "修改后的策略组ID"
// @router   /:id/policy [put]
func (this *SonUserController) Policy() {
	id, _ := this.GetInt64("Id")
	if id == 0 {
		this.Error(tsOpCode.OPERATION_REQUEST_FAILED)
	}

	var req reqEditPolicy
	if err := this.ParseForm(&req); err != nil {
		logs.Debug(err)
		this.Error(tsOpCode.OPERATION_REQUEST_FAILED)
	}

	// 验证
	policy := admin.DataChannelPolicy{Id: req.PolicyId}
	if err := tsDb.NewDbBase().DbGet(&policy); err != nil {
		this.Error(tsOpCode.DB_SELECT_ERROR, err.Error())
	}

	// 更新
	admin := sys.Admin{Id: id, PolicyId: req.PolicyId}
	if err := tsDb.NewDbBase().DbUpdate(&admin, "PolicyId"); err != nil {
		this.Error(tsOpCode.DB_UPDATE_ERROR, err.Error())
	} else {
		this.Success(nil)
	}
}

// @Title 编辑
// @Description 编辑
// @Success 200 {object} sys.Account
// @Param    Id         	  formData    int    	true   用户ID
// @Param    Username         formData    string    true   登陆名
// @Param    Password         formData    string    true   用户密码
// @Param    ConfirmPassword  formData    string    true   用户确认密码
// @Param    Status           formData    int       false  状态
// @Param    Role             formData    string    true   角色
// @Param    CreateRole       formData    string    true   可创建角色
// @Param    Name             formData    string    true   名称
// @Param    Mobile           formData    string    true   联系电话
// @router   / [put]
func (this *SonUserController) Edit() {
	var req dto.ReqEditAdmin
	if err := this.ParseForm(&req); err != nil {
		logs.Debug("ParseForm:", err)
		this.Error(tsOpCode.OPERATION_REQUEST_FAILED)
	}
	//DONE 验证操作者身上是否有该权限
	mHandle := sys.Admin{}
	mHandle.Id = this.AdminId
	err := tsDb.NewDbBase().DbGet(&mHandle)
	if err != nil {
		logs.Error("[SonUser][获取账号角色]:DBError: ", err)
		this.Error(tsOpCode.OPERATION_DB_FAILED)
	}
	//roleStr := mHandle.Role
	//roleArr := tsString.CoverStringToArray(roleStr, ",", false)
	//reqRoleArr := tsString.CoverStringToArray(req.Role, ",", false)
	//for _, reqRole := range reqRoleArr {
	//	if !tsContain.InArrayString(roleArr, reqRole) {
	//		logs.Error("[SonUser][检查权限设定] 自己权限不足：SelfRole:", roleArr, " ReqRole:", reqRoleArr)
	//		this.Error(tsOpCode.NO_HAVE_OPEN_CHANNEL)
	//	}
	//}
	a := sys.Admin{}
	o := sys.Admin{}
	err = a.GetSampleAdminById(req.Id)
	if err != nil {
		this.Error(tsOpCode.OPERATION_DB_FAILED)
	}
	code := o.EditAdmin(req, this.PersonInfo, this.RealAdminId)
	if code != tsOpCode.OPERATION_SUCCESS {
		this.Error(int(code))
	}
	//handleLog := models.HandleLogs{}
	//if a.Status != req.Status {
	//	handleLog.Log(consts.LogTemplateEmployeeDisable, this.AgentId, this.AdminId, this.AdminUsername, this.Ip, beego.M{"Username": a.Username, "Status": o.Status})
	//}
	//handleLog.Log(consts.LogTemplateEmployeeEdit, this.AgentId, this.AdminId, this.AdminUsername, this.Ip, beego.M{"Username": a.Username})
	this.Success(code)
}

// @Title 删除
// @Description 删除
// @Success 200 {object} sys.Account
// @Param    Id         	  formData    int    	true   用户ID
// @router /:id [delete]
func (this *SonUserController) Del() {
	id, _ := this.GetInt64(":id")
	if id < 1 {
		this.Error(tsOpCode.OPERATION_REQUEST_FAILED, "id")
	}

	o := sys.Admin{}
	if err := o.DeleteAdmin(id, this.RealAdminId); err != nil {
		this.Error(tsOpCode.OPERATION_DB_FAILED)
	}
	//handleLog := models.HandleLogs{}
	//handleLog.Log(consts.LogTemplateEmployeeDel, this.AgentId, this.AdminId, this.AdminUsername, this.Ip, beego.M{"Username": o.Username})
	this.Success("success")
}


type respSonList struct {
	Id         int64
	Name       string // 客户名称
	AppAccount string // 绑定的云平台代理账户
	AppRemark  string // 备注
	UserName   string // 登陆账户
	Mobile     string // 绑定手机
	PolicyId   int64  // 支付策略
	PolicyName string // 支付策略
	CreateTime int64
}

type reqEditPolicy struct {
	PolicyId int64
}

type respCount struct {
	Count int
}
