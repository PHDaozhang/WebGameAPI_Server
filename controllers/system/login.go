package system

import (
	"encoding/json"
	"fmt"
	"tsEngine/tsCrypto"
	"tsEngine/tsDb"
	"tsEngine/tsJson"
	"tsEngine/tsOpCode"
	"tsEngine/tsRedis"
	"tsEngine/tsString"
	"tsEngine/tsTime"
	"tsEngine/tsToken"
	"web-game-api/conf"
	"web-game-api/core/consts"
	"web-game-api/models/dto"
	"web-game-api/models/sys"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
)

//登录
type LoginController struct {
	PermissionController
}

func (this *LoginController) Prepare() {
	// 这个类里面不需要做权限检测。覆盖掉父类的方法
	logs.Debug("LoginController Prepare")
}

// @Title 登陆
// @Description 登陆
// @Success 200 {object} admin.Admin
// @Param    Username    formData    string    true   登陆名
// @Param    Password    formData    string    true   用户密码
// @Param    Captcha     formData    string    false  验证码
// @router   /login [post]
func (this *LoginController) Login() {
	passwordSalt := conf.PasswordSalt

	form := dto.ReqLogin{}
	if err := this.ParseForm(&form); err != nil {
		logs.Error("login", err)
		this.Error(tsOpCode.OPERATION_REQUEST_FAILED)
	}

	oAdmin := sys.Admin{
		Username: form.Username,
		Password: tsCrypto.GetMd5([]byte(form.Password + passwordSalt)),
	}

	db := tsDb.NewDbBase()
	timestamp := tsTime.CurrSe()
	err := db.DbRead(&oAdmin, "Username", "Password")
	if err != nil {
		logs.Error(err, tsJson.ToJson(oAdmin))
		this.Error(tsOpCode.PASSWORD_ERROR)
	}
	this.AdminId = oAdmin.Id
	ex, _ := tsRedis.Exists(fmt.Sprintf(consts.KeyWEBAPIIsLogin, this.AdminId))
	if ex == 1 {
		logs.Error("[Login][Login]账户已登录 adminId: ", this.AdminId)
		//TODO wisp
		//this.Error(tsOpCode.ACCOUNT_ALREADY_LOGIN)
	}

	//非开发者时候记录登录时间和IP
	oAdmin.LoginTime = timestamp
	oAdmin.LoginIp = this.Ctx.Input.IP()

	//更新用户登录时间不是关键数据不需要判断错误
	go db.DbUpdate(&oAdmin, "LoginTime", "LoginIp")

	// 免登陆最长时间 60*60*24*10
	tokenMaxExp := beego.AppConfig.DefaultInt64("TokenMaxExpSecond", 864000)
	logs.Debug("set login redis:", tsString.FromInt64(this.AdminId), tokenMaxExp)

	//rAdmin := oAdmin
	if oAdmin.AdminType == consts.AdminTypeSonuser {
		// 直接获取父亲（因为子账号只能由公司账号创建）

		// 支持无限极向上查
		//oa := admin.Account{}
		//realAdminId, _ = oa.GetSonParentAdmin(oAdmin.Id, oAdmin.ParentTree)

		//rAdmin = admin.Admin{
		//	//Id: oAdmin.ParentId,
		//	Id: realAdminId,
		//}
		//err := db.DbRead(&rAdmin, "Id")
		if err != nil {
			logs.Error(err)
			this.Error(tsOpCode.DATA_NOT_EXIST)
		}
	}
	go tsRedis.Set(tsString.FromInt64(this.AdminId), "login", tokenMaxExp)

	// 获取节点和角色
	oNav, oRole, createRoles, err := GetNavPermission(oAdmin)
	if err != nil {
		this.Code = tsOpCode.OPERATION_DB_FAILED
		this.TraceJson()
	}

	token := getToken(oAdmin.Id, int64(timestamp))

	this.Success(map[string]interface{}{"Admin": oAdmin, "Nav": oNav, "Role": oRole, "CRole": createRoles, "Token": token})
}

// @Title 退出登陆
// @Description 退出登陆
// @Success 200 {"Code":200,"Data":null}
// @router  /logout [get]
func (this *LoginController) Logout() {
	loginAdmin := this.GetUserFormToken()
	logs.Trace("IsLogin 获取 adminId 值", loginAdmin.Id)
	go tsRedis.Del(fmt.Sprintf(consts.KeyWEBAPIIsLogin, loginAdmin.Id))
	//go tsRedis.Del(tsString.FromInt64(loginAdmin.Id))
	this.Success(nil)
}

// @Title 检测登录状态
// @Description 检测登录状态，获取最新的Token
// @Success 200 {object} admin.Admin
// @Param    Token    header    string    true   token
// @router  /islogin [get]
func (this *LoginController) IsLogin() {
	loginAdmin := this.GetUserFormToken()
	logs.Trace("IsLogin 获取 adminId 值", loginAdmin.Id)

	loginId := loginAdmin.Id
	if loginId == 0 {
		this.Error(tsOpCode.TIME_OUT)
	}

	var oAdmin sys.Admin
	oAdmin.Id = loginId

	db := tsDb.NewDbBase()

	err := db.DbRead(&oAdmin)
	if err != nil {
		logs.Error(err)
		this.Error(tsOpCode.OPERATION_DB_FAILED)
	} else if oAdmin.Username == "" {
		this.Error(tsOpCode.USER_NOT_LOGIN)
	}
	//}
	//区分子账号
	realAdminId := oAdmin.Id
	rAdmin := oAdmin
	if oAdmin.AdminType == consts.AdminTypeSonuser {
		realAdminId = oAdmin.ParentId
		rAdmin = sys.Admin{
			//Id: oAdmin.ParentId,
			Id: realAdminId,
		}
		err := db.DbRead(&rAdmin, "Id")
		if err != nil {
			logs.Error(err)
			this.Error(tsOpCode.PASSWORD_ERROR)
		}
	}
	//获取节点和权限
	oNav, oRole, createRoles, err := GetNavPermission(oAdmin)
	if err != nil {
		this.Error(tsOpCode.OPERATION_DB_FAILED)
	}

	// 获取account信息
	account := map[string]interface{}{}
	account["AdminId"] = realAdminId

	data, err := json.Marshal(rAdmin)
	if err != nil {
		logs.Error(err)
		this.Error(tsOpCode.JSON_UNMARSHAL_FAILED)
	}
	_ = tsRedis.SetNX(fmt.Sprintf(consts.KeyWEBAPISysAdminById, rAdmin.Id), string(data), 3600)
	_ = tsRedis.SetNX(fmt.Sprintf(consts.KeyWEBAPISysAdminByName, rAdmin.Username), string(data), 3600)
	result := map[string]interface{}{
		"Admin":   oAdmin,
		"Nav":     oNav,
		"Role":    oRole,
		"CRole":   createRoles,
		"Account": account,
		"Token":   getToken(oAdmin.Id, int64(tsTime.CurrSe())),
	}
	this.Success(result)
}

func getToken(adminId, timestamp int64) string {
	go tsRedis.Set(fmt.Sprintf(consts.KeyWEBAPIIsLogin, adminId), 1, conf.TokenExpMinute*60)
	return tsToken.ToToken(beego.M{"Id": adminId, "CreateTime": timestamp}, conf.TokenSalt, conf.TokenExpMinute)
}
