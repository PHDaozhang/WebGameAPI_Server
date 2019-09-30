package system

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"strings"
	"tsEngine/tsCrypto"
	"tsEngine/tsDb"
	"tsEngine/tsJson"
	"tsEngine/tsOpCode"
	"tsEngine/tsRedis"
	"tsEngine/tsString"
	"tsEngine/tsTime"
	"web-game-api/conf"
	"web-game-api/core/consts"
	"web-game-api/models/sys"
)

type PermissionController struct {
	BaseController
}

// 在这里处理后，其他函数中就不需要雷同代码了。
func (this *PermissionController) Prepare() {
	logs.Info("PermissionController Prepare")
	// 权限判断
	this.CheckPermission()
}

//权限判断
func (this *PermissionController) CheckPermission() {

	//params := strings.Split(strings.ToLower(this.Ctx.Request.RequestURI), "/")
	controller, action := this.GetControllerAndAction()
	controller = strings.Replace(strings.ToLower(controller), "controller", "", -1)
	action = strings.ToLower(action)

	//// 注意，有编辑（edit）权限，就会有获取详情权限
	//if action == "get" {
	//	action = "edit"
	//}
	if controller == "error" {
		logs.Error("权限判断：", controller, this.Ctx.Request.RequestURI)
	} else {
		logs.Trace("权限判断：", controller, action)
	}

	// 赋值全局共用参数
	this.CheckLogin()
	db := tsDb.NewDbBase()

	//如果是开发者直接返回
	uid := conf.SystemAdminId
	if this.AdminId == uid {
		this.Role = "admin"
		return
	}

	var err error

	adminRoles := tsString.CoverStringToArray(this.PersonInfo.Role, ",", false)
	logs.Trace("[Base]接口验证:", controller, action)
	md5 := tsCrypto.GetMd5([]byte(controller + action))
	pass := 0
	var permissions []string
	permissionStr := ""
	if len(adminRoles) == 0 {
		this.Error(tsOpCode.NO_PERMISSION)
	}

	var (
		oRole sys.Role
		keys  []string
	)
	for _, id := range adminRoles {
		keys = append(keys, fmt.Sprintf(consts.KeyWEBAPISysRoleById, id))
	}
	roles := tsRedis.MGet(keys...)
	if len(roles) <= 0 {
		for _, role := range roles {
			err = json.Unmarshal([]byte(role.(string)), &oRole)
			if err != nil {
				continue
			}
			permissionStr += oRole.Permission
		}
	} else {
		list, err := db.DbInIds(&oRole, "Id", adminRoles)
		if err != nil {
			this.Error(tsOpCode.OPERATION_DB_FAILED)
		}
		for _, v := range list {
			permissionStr += v["Permission"].(string)
		}
	}
	permissions = tsString.CoverStringToArray(permissionStr, ",", false)
	var oMode sys.Mode
	err = oMode.GetModeByMD5(md5)
	if err != nil {
		logs.Error(err)
		this.Error(tsOpCode.NO_PERMISSION)
	}
	for _, permission := range permissions {
		if oMode.Id == tsString.ToInt64(permission) {
			pass = 1
			break
		}
	}

	logs.Trace("pass：", pass)

	if oMode.Logs == 1 {
		logs.Trace("参数", this.Ctx.Request.PostForm)
		go func(currentMode sys.Mode, userName string) {
			//记录操作日志
			var oLogs sys.Logs
			oLogs.Mode = oMode.ParentId
			oLogs.Action = oMode.Id
			oLogs.AdminId = this.PersonInfo.Id
			oLogs.Pass = pass
			oLogs.CreateTime = tsTime.CurrSe()
			// 用户:{user}{action}了{menu}的数据
			oLogs.TemplateId = 1
			if currentMode.ParentId == -1 {
				oLogs.Content = tsJson.ToJson(beego.M{"User": userName, "Action": "打开", "Menu": currentMode.Name})
			} else {
				parentMode := sys.Mode{Id: currentMode.ParentId}
				if tsDb.NewDbBase().DbRead(&parentMode, "Id") != nil {
					return
				}
				oLogs.Content = tsJson.ToJson(&beego.M{"User": userName, "Action": currentMode.Name, "Menu": parentMode.Name})
			}

			oLogs.Description = tsJson.ToJson(this.Ctx.Request.PostForm)

			db := tsDb.NewDbBase()
			db.DbInsert(&oLogs)
		}(oMode, this.AdminUsername)
	}

	if pass == 1 {
		return
	}
	this.Error(tsOpCode.NO_PERMISSION)
}

//校验是否为开发者
func (this *PermissionController) CheckRoot() {
	// TODO
	return

	loginAdmin := this.GetUserFormToken()
	this.AdminId = loginAdmin.Id

	if this.AdminId == 0 {

		this.Error(tsOpCode.OPERATION_SUCCESS)
	}

	uid := conf.SystemAdminId
	if this.AdminId != uid {
		this.Error(tsOpCode.NO_PERMISSION)
	}
}

//IP过滤判断
func (this *PermissionController) CheckIp() {

	ip := this.Ctx.Input.IP()
	//ip过滤
	var oIpban sys.Ipban

	//获取秒级时间戳
	nowTime := tsTime.CurrSe()
	db := tsDb.NewDbBase()
	data, err := tsRedis.Get(fmt.Sprintf(consts.KeyWEBAPISysIpbanByIp, ip))
	if data != "" {
		err = json.Unmarshal([]byte(data), &oIpban)
		if err != nil {
			this.Error(tsOpCode.JSON_UNMARSHAL_FAILED)
		}
	}
	if err != nil {
		list, err := db.DbList(&oIpban, "Ip", ip)
		if err != nil {
			logs.Error(err)
			this.Error(tsOpCode.OPERATION_DB_FAILED)
		}

		for _, v := range list {
			if v["Start"].(uint64) < nowTime && v["End"].(uint64) > nowTime {
				this.Error(tsOpCode.IP_BLOCKED)
			}
		}
	}

	if oIpban.Start < nowTime && oIpban.End > nowTime {
		this.Error(tsOpCode.IP_BLOCKED)
	}
}

//权限判断
func (this *PermissionController) CheckLogin() {
	loginAdmin := this.GetUserFormToken()
	logs.Trace("获取 adminId 值", loginAdmin.Id)

	if loginAdmin.Id == 0 {
		this.Error(tsOpCode.TIME_OUT)
	}

	db := tsDb.NewDbBase()

	//ip检测
	this.CheckIp()
	var oAdmin sys.Admin
	this.AdminId = loginAdmin.Id
	data, err := tsRedis.Get(fmt.Sprintf(consts.KeyWEBAPISysAdminById, this.AdminId))
	if err != nil {
		oAdmin.Id = this.AdminId
		// 读取个人信息
		err = db.DbRead(&oAdmin)
	} else {
		err = json.Unmarshal([]byte(data), &oAdmin)
	}
	if err != nil {
		this.Error(tsOpCode.JSON_UNMARSHAL_FAILED)
		return
	}

	this.PersonInfo = oAdmin
	this.RealAdminId = this.AdminId
	this.AdminUsername = oAdmin.Username
	this.AdminName = oAdmin.Name
	this.AgentId = oAdmin.AgentId

	// 如果是子账号，则用父节点账号
	if this.PersonInfo.AdminType == consts.AdminTypeSonuser {
		this.RealAdminId = oAdmin.ParentId
	}

	return
}
