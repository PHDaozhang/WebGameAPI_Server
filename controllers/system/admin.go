//管理员管理
package system

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"gopkg.in/mgo.v2/bson"
	"tsEngine/tsCrypto"
	"tsEngine/tsDb"
	"tsEngine/tsOpCode"
	"tsEngine/tsRedis"
	"tsEngine/tsString"
	"tsEngine/tsTime"
	"web-game-api/conf"
	"web-game-api/core/consts"
	"web-game-api/models/dto"
	"web-game-api/models/sys"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/validation"
)

type AdminController struct {
	PermissionController
}

// @Title 管理员列表
// @Description 管理员列表
// @Success 200 {object} admin.Admin
// @Param    Keyword    query    string    false  搜索词
// @Param    Page       query    string    true   页码
// @Param    PageSize   query    string    true   单页数据量
// @Param    BeginTime  query    string    false  过滤开始时间
// @Param    EndTime    query    string    false  过滤结束时间
// @router   /list [get]
func (this *AdminController) List() {
	var req dto.ReqSearch
	if err := this.ParseForm(&req); err != nil {
		this.Error(tsOpCode.OPERATION_REQUEST_FAILED)
	}
	var beginTime, endTime string
	if req.BeginTime > 0 {
		beginTime = tsTime.SeToString(uint64(req.BeginTime), "2006-01-02 15:04:05")
	}

	if req.EndTime > 0 {
		endTime = tsTime.SeToString(uint64(req.EndTime), "2006-01-02 15:04:05")
	}

	o := sys.Admin{}
	items, pagination, err := o.ListChildSlim(req.Page, req.PageSize, beginTime, endTime, req.Keyword)

	if err != nil {
		logs.Error(err)
		this.Error(tsOpCode.OPERATION_DB_FAILED)
	}

	this.Success(bson.M{"Items": items, "Pagination": pagination})
}

// @Title 添加管理员
// @Description 添加管理员
// @Success 200 {object} admin.Admin
// @Param    Username    formData    string    true   登陆名
// @Param    Password    formData    string    true   用户密码
// @Param    Role        formData    string    false  角色
// @Param    Name        formData    string    false  名字
// @Param    ContactInf  formData    string    false  联系方式
// @Param    Mobile      formData    string    false  电话
// @Param    AdminType   formData    int       false  类型
// @router   / [post]
func (this *AdminController) Add() {
	var req dto.ReqAddAdmin
	if err := this.ParseForm(&req); err != nil {
		logs.Debug("ParseForm:", err)
		this.Error(tsOpCode.OPERATION_REQUEST_FAILED)
	}

	var err error
	o := sys.Admin{
		Username:  req.Username,
		Password:  req.Password,
		Status:    1,
		Role:      req.Role,
		Name:      req.Name,
		Mobile:    req.Mobile,
		AdminType: req.AdminType, // 1代理 2推广 3币商 4子账户
	}

	//过滤开发者账号
	if o.Username == beego.AppConfig.String("Username") {
		logs.Debug("Username:", o.Username)
		this.Error(tsOpCode.OPERATION_REQUEST_FAILED)
	}

	//ConfirmPassword := this.GetString("ConfirmPassword")

	if req.Role != "" {
		o.Role = "," + req.Role + ","
	}

	//****************************************************
	//数据验证
	valid := validation.Validation{}

	//用户名验证
	valid.Required(o.Username, "Username").Message("10010")
	valid.MinSize(o.Username, 2, "UserNameMin").Message("10011")
	valid.MaxSize(o.Username, 20, "UserNameMax").Message("10012")
	valid.AlphaDash(o.Username, "UserNameAlphaDash").Message("10013")
	//密码验证
	valid.Required(o.Password, "Password").Message("10014")
	valid.MinSize(o.Password, 6, "PasswordMin").Message("10015")
	valid.MaxSize(o.Password, 50, "PasswordMax").Message("10016")
	//if o.Password != ConfirmPassword {
	//	this.Error(tsOpCode.PASSWORD_INCONSISTENT)
	//}
	valid.Range(int(o.Status), 1, 2, "Status").Message("10018")

	if o.Name != "" {
		valid.MaxSize(o.Name, 20, "Name").Message("10019")
	}

	if o.Mobile != "" {
		valid.Mobile(o.Mobile, "Mobile").Message("%v", tsOpCode.MOBILE_ERROR)
	}

	if valid.HasErrors() {
		// 如果有错误信息，证明验证没通过
		// 打印错误信息
		for _, err := range valid.Errors {
			this.Error(tsString.ToInt(err.Message))
		}
	}

	o.Password = tsCrypto.GetMd5([]byte(o.Password + conf.PasswordSalt))
	o.CreateTime = tsTime.CurrSe()
	o.UpdateTime = o.CreateTime
	o.LoginTime = o.CreateTime

	db := tsDb.NewDbBase()

	logs.Trace(o)

	if o.Username != "" {
		// 检测账号唯一性
		isOnly := o.CheckUsernameOnly(o.Username, 0)
		logs.Trace(isOnly)
		if !isOnly {
			this.Error(tsOpCode.USER_NAME_EXIST)
		}
	}

	if o.Mobile != "" {
		isOnly := o.CheckPhoneOnly(o.Mobile)
		if !isOnly {
			this.Error(tsOpCode.TEL_NUMBER_EXISTED)
		}
	}

	// 设置当前
	o.ParentId = -1
	o.ParentTree = ",-1,"

	_ = db.Transaction()
	defer db.TransactionEnd()

	adminId, err := db.DbInsert(&o)
	if err != nil {
		db.SetRollback(true)
		logs.Error(err)
		this.Error(tsOpCode.OPERATION_DB_FAILED)
	}

	//更新ParentTree，链接自身的Id
	o.ParentTree += fmt.Sprintf("%d,", adminId)

	err = db.DbUpdate(&o, "ParentTree")
	if err != nil {
		db.SetRollback(true)
		logs.Error(err)
		this.Error(tsOpCode.OPERATION_DB_FAILED)
	}

	data, err := json.Marshal(o)
	if err != nil {
		logs.Error(err)
	}
	_ = tsRedis.Set(fmt.Sprintf(consts.KeyWEBAPISysAdminByName, o.Username), string(data), 3600)
	_ = tsRedis.Set(fmt.Sprintf(consts.KeyWEBAPISysAdminById, o.Id), string(data), 3600)
	this.Success(nil)
}

// @Title 获取详情
// @Description 获取详情
// @Success 200 {object} models.Mode
// @Param    id   path    int       true   id
// @router   /:id [get]
func (this *AdminController) Get() {
	this.CheckRoot() //校验是否为开发者

	//初始化
	db := tsDb.NewDbBase()
	o := sys.Admin{}

	//获取get数据
	o.Id, _ = this.GetInt64(":id", 0)

	data, _ := tsRedis.Get(fmt.Sprintf(consts.KeyWEBAPISysAdminById, o.Id))
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
		this.Error(tsOpCode.OPERATION_DB_FAILED)
	}
	sData, err := json.Marshal(o)
	if err != nil {
		logs.Error(err)
	}
	_ = tsRedis.Set(fmt.Sprintf(consts.KeyWEBAPISysAdminById, o.Id), string(sData), 3600)
	_ = tsRedis.Set(fmt.Sprintf(consts.KeyWEBAPISysAdminByName, o.Username), string(sData), 3600)
	this.Success(o)
}

// @Title 编辑管理员
// @Description 编辑管理员
// @Success 200 {object} admin.Admin
// @Param    Id               formData    int       true   主键
// @Param    Username         formData    string    true   登陆名
// @Param    Password         formData    string    true   用户密码
// @Param    ConfirmPassword  formData    string    true   用户确认密码
// @Param    Status           formData    int       false  状态
// @Param    Role             formData    string    true   角色
// @Param    Name             formData    string    true   名称
// @Param    Mobile           formData    string    true   联系电话
// @router   / [put]
func (this *AdminController) Edit() {

	//获取post数据
	id, _ := this.GetInt64("Id", 0)
	var req dto.ReqAddAdmin
	if err := this.ParseForm(&req); err != nil {
		this.Error(tsOpCode.OPERATION_REQUEST_FAILED)
	}

	db := tsDb.NewDbBase()
	o := sys.Admin{
		Id: id,
	}

	err := db.DbRead(&o)
	if err != nil {
		logs.Error(err)
		this.Error(tsOpCode.DATA_NOT_EXIST)
	}

	oldPassword := o.Password
	oldUsername := o.Username
	oldMobile := o.Mobile

	o.Name = req.Name
	o.Mobile = req.Mobile
	// o.Username = req.Username // 应该不允许变更的
	o.Password = req.Password
	o.Status = req.Status
	o.Role = req.Role

	//过滤开发者账号
	if o.Username == beego.AppConfig.String("Username") {
		this.Error(tsOpCode.OPERATION_REQUEST_FAILED)
	}

	ConfirmPassword := this.GetString("ConfirmPassword")

	if req.Role != "" {
		o.Role = "," + req.Role + ","
	}

	//****************************************************
	//数据验证
	valid := validation.Validation{}
	//用户名验证
	valid.Required(o.Username, "Username").Message("10010")
	valid.MinSize(o.Username, 2, "UserNameMin").Message("10011")
	valid.MaxSize(o.Username, 20, "UserNameMax").Message("10012")
	valid.AlphaDash(o.Username, "UserNameAlphaDash").Message("10013")
	//密码验证
	if o.Password != "" {
		valid.Required(o.Password, "Password").Message("10014")
		valid.MinSize(o.Password, 6, "PasswordMin").Message("10015")
		valid.MaxSize(o.Password, 50, "PasswordMax").Message("10016")
		if o.Password != ConfirmPassword {
			this.Error(tsOpCode.PASSWORD_CONFIRM_FAILED)
		}

		o.Password = tsCrypto.GetMd5([]byte(o.Password + conf.PasswordSalt))

	} else {
		o.Password = oldPassword
	}
	valid.Range(int(o.Status), 1, 2, "Status").Message("10018")

	if o.Name != "" {
		valid.MaxSize(o.Name, 20, "Name").Message("10019")
	}

	if o.Mobile != "" {
		valid.Mobile(o.Mobile, "Mobile").Message("%v", tsOpCode.MOBILE_ERROR)
	}

	if valid.HasErrors() {
		// 如果有错误信息，证明验证没通过
		// 打印错误信息
		for _, err := range valid.Errors {
			this.Error(tsString.ToInt(err.Message))
		}
	}

	admin := sys.Admin{}
	admin.Id = o.Id

	if o.Username != oldUsername && o.Username != "" {
		// 检测账号唯一性
		isOnly := o.CheckUsernameOnly(o.Username, o.Id)
		if !isOnly {
			this.Error(tsOpCode.ACCOUNT_ERROR)
		}
	}

	if o.Mobile != oldMobile && o.Mobile != "" {
		// 检测账号唯一性
		isOnly := o.CheckPhoneOnly(o.Mobile)
		if !isOnly {
			this.Error(tsOpCode.TEL_NUMBER_EXISTED)
		}
	}

	db.Transaction()
	defer db.TransactionEnd()

	o.UpdateTime = tsTime.CurrSe()
	err = db.DbUpdate(&o, "Username", "Password", "Name", "Role", "Mobile")
	if err != nil {
		db.SetRollback(true)
		logs.Error(err)
		this.Error(tsOpCode.OPERATION_DB_FAILED)
	}

	// 如果修改买金比率和推广收益比率
	//BuyScoreRate, _ := this.GetFloat("BuyScoreRate")
	//ReturnRate, _ := this.GetFloat("ReturnRate")
	//
	//mA := admin.Account{}
	//mA.AdminId = o.Id
	//db.DbGet(&mA, "AdminId")
	//
	//if BuyScoreRate > 0 {
	//	mA.BuyScoreRate, _ = this.GetFloat("BuyScoreRate")
	//}
	//
	//if ReturnRate > 0 {
	//	if ReturnRate < mA.ReturnRate {
	//		db.SetRollback(true)
	//		this.Error(tsOpCode.RATE_CAN_NOT_HEIGH_THEN_SELF)
	//	}
	//
	//	mA.ReturnRate, _ = this.GetFloat("ReturnRate")
	//}
	//
	//if BuyScoreRate > 0 || ReturnRate > 0 {
	//	mA.UpdateTime = tsTime.CurrSe()
	//	db.DbUpdate(&mA, "BuyScoreRate", "ReturnRate", "UpdateTime")
	//}

	data, err := json.Marshal(o)
	if err != nil {
		logs.Error(err)
	}

	_ = tsRedis.Set(fmt.Sprintf(consts.KeyWEBAPISysAdminByName, o.Username), string(data), 3600)
	_ = tsRedis.Set(fmt.Sprintf(consts.KeyWEBAPISysAdminById, o.Id), string(data), 3600)
	this.Success(o)
}

// @Title 删除管理员
// @Description 删除管理员
// @Success 200 {object} admin.Admin
// @Param    id          path    int    true   管理员ID
// @router   /:id [delete]
func (this *AdminController) Del() {
	db := tsDb.NewDbBase()
	o := sys.Admin{}
	Id, _ := this.GetInt64(":id")
	o.Id = Id
	err := db.DbRead(&o)
	if err != nil {
		logs.Debug(err)
		this.Error(tsOpCode.DATA_NOT_EXIST)
	}

	_, err = db.DbDel(&o)

	if err != nil {
		logs.Error(err)
		this.Error(tsOpCode.OPERATION_DB_FAILED)
	}
	_ = tsRedis.Del(fmt.Sprintf(consts.KeyWEBAPISysAdminByName, o.Username))
	_ = tsRedis.Del(fmt.Sprintf(consts.KeyWEBAPISysAdminById, Id))
	this.Success(nil)
}
