package system

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/beego/i18n"
	"reflect"
	_ "strconv"
	"strings"
	"tsEngine/tsDb"
	"tsEngine/tsJson"
	"tsEngine/tsOpCode"
	"tsEngine/tsRedis"
	"tsEngine/tsString"
	"tsEngine/tsTime"
	"tsEngine/tsToken"
	"tsEngine/tsValid"
	"web-game-api/conf"
	"web-game-api/logic/errorCode"
	"web-game-api/models/dto"
	"web-game-api/models/sys"
)

type BaseController struct {
	beego.Controller
	i18n.Locale
	Code          int
	Msg           string
	Result        interface{}
	AdminId       int64
	AgentId       int64
	AdminUsername string
	AdminName     string
	RealAdminId   int64 // 如果当前是子账号登入，这个值是代理或推广账号ID
	PersonInfo    sys.Admin
	Role          string
}

var langTypes []string // Languages that are supported.

func init() {
	logs.Trace("初始化控制器")
	//获取语言包列表
	langTypes = beego.AppConfig.Strings("LangTypes")
	for _, lang := range langTypes {
		if lang != "" {
			logs.Trace("载入语言包: " + lang)
			if err := i18n.SetMessage(lang, "conf/"+"locale_"+lang+".ini"); err != nil {
				logs.Error("Fail to set message file:", err)
				return
			}
		}
	}
}

// 在这里处理后，其他函数中就不需要雷同代码了。
func (this *BaseController) Prepare() {
	logs.Info("BaseController Prepare")
	logs.Info("Request.Method:", this.Ctx.Input.Method())
	logs.Info("Request.Header:", tsJson.ToJson(this.Ctx.Request.Header))
	logs.Info("Request.Host:", this.Ctx.Input.Host())
	logs.Info("Request.URL:", tsJson.ToJson(this.Ctx.Request.URL))
	logs.Info("Request.RequestBody:", string(this.Ctx.Input.RequestBody))
}
func (this *BaseController) Success(obj interface{}) {
	this.Code = tsOpCode.OPERATION_SUCCESS
	this.Result = obj
	this.TraceJson()
}

func (this *BaseController) APISuccess(obj interface{}) {
	this.Code = errorCode.SUCC
	if obj == nil {
		this.Result = map[string]interface{}{}
	} else {
		this.Result = obj
	}
	this.Msg = "ok"
	this.TraceJson()
}

/**
调用此函数，会终止当前函数内其他的逻辑处理（跳出）
*/
func (this *BaseController) Error(code int, msg ...string) {
	if code < 1 {
		code = tsOpCode.SERVER_ERROR
	}
	this.Code = code

	if beego.BConfig.RunMode == beego.DEV && len(msg) > 0 {
		this.Msg = msg[0]
	} else {
		this.CheckLanguage()
		if this.Code > 0 {
			this.Msg = this.Tr("error." + fmt.Sprintf("%d", this.Code))
		}
		logs.Warn("[BaseController] error msg:", msg)
	}

	this.TraceJson()
}

//json 输出
func (this *BaseController) TraceJson() {
	this.Data["json"] = &map[string]interface{}{"Code": this.Code, "Msg": this.Msg, "Data": this.Result}
	this.ServeJSON()
	this.StopRun()
}

//检测语言包
func (this *BaseController) CheckLanguage() {

	//设置语言
	this.Lang = ""

	// 1. 获取 'Accept-Language' 值
	al := this.Ctx.Request.Header.Get("Accept-Language")
	if len(al) > 4 {
		al = al[:5] // Only compare first 5 letters.
		if i18n.IsExist(al) {
			this.Lang = al
		}
	}

	// 2. 默认为中文
	if len(this.Lang) == 0 {
		this.Lang = "zh-CN"
	}

}

func (this *BaseController) GetUserFormToken() (loginToken sys.LoginToken) {
	token := this.Ctx.Input.Header("Token")

	tokenMap := tsToken.FromToken(token, conf.TokenSalt)
	if tokenMap == nil {
	}

	adminId, _ := tokenMap["Id"]
	logs.Trace("adminId:", adminId)

	if adminId == nil {
		loginToken = sys.LoginToken{
			Id: 0,
		}
	} else {
		loginToken = sys.LoginToken{
			Id: int64(adminId.(float64)),
		}
	}

	return
}

// 通用的list列表
func (this *BaseController) CommList(model interface{}, keywordExp []string, sortOrder []string, showFields []string, andCond ...*orm.Condition) {
	var req dto.ReqSearch
	if err := this.ParseForm(&req); err != nil {
		logs.Error("ParseForm error:", err)
		this.Error(tsOpCode.OPERATION_REQUEST_FAILED)
	}

	var sort []string
	if len(req.Sort) > 0 {
		sort = append(sort, req.Sort)
	}
	sort = append(sort, sortOrder...)

	// where ChannelId = 'ss1001' or ChannelId = 'ss1002'
	if len(req.FilterKey) > 0 && len(req.FilterValue) > 0 {
		filterSlice := strings.Split(req.FilterValue, ",")
		cond := orm.NewCondition()

		for _, filter := range filterSlice {
			cond = cond.Or(req.FilterKey, filter)
		}
		if andCond != nil {
			andCond = append(andCond, cond)
		}
	}

	items, pagination, err := tsDb.NewDbBase().DbSearchPage(model, req.Page, req.PageSize, req.Keyword, keywordExp, sort, showFields, andCond...)

	if err != nil && err != orm.ErrNoRows {
		logs.Error(err)
		this.Error(tsOpCode.OPERATION_DB_FAILED)
	}

	this.Success(beego.M{"Items": items, "Pagination": pagination})
}

// 通用的添加处理
func (this *BaseController) CommAdd(adminId, createAdminId int64, channelId string, req dto.CommReqInterface, model sys.CommStructInterface, callback ...CommonAddHandler) {
	this.HandleAddReq(adminId, createAdminId, channelId, req, model)

	// 启动事务
	tsConn := tsDb.NewDbBase()
	tsConn.Transaction()
	defer tsConn.TransactionEnd()

	// other setting
	if len(callback) > 0 {
		for _, f := range callback {
			f(tsConn, model)
		}
	}

	// 插入
	if _, err := tsConn.DbInsert(model); err != nil {
		logs.Error("DbInsert error:", err)
		tsConn.SetRollback(true)
		this.Error(tsOpCode.OPERATION_DB_FAILED)
	}

	this.Success(model)
}

// 适合单表查询
func (this *BaseController) CommGet(model sys.CommStructInterface, fields ...string) {
	id, _ := this.GetInt64(":id", 0)
	if id < 1 {
		this.Error(tsOpCode.DATA_NOT_EXIST)
	}

	model.SetId(id)
	if len(fields) == 0 {
		fields = []string{"Id", "Deleted"}
	}
	if err := tsDb.NewDbBase().DbGet(model, fields...); err != nil {
		logs.Error("DbGet error:", err)
		this.Error(tsOpCode.OPERATION_DB_FAILED)
	}
	this.Success(model)
}

func (this *BaseController) CommEdit(req dto.CommReqInterface, model sys.CommStructInterface, updateField ...string) {
	this.HandleEditReq(req, model)
	if err := tsDb.NewDbBase().DbUpdate(model, updateField...); err != nil {
		logs.Error("DbUpdate error:", err)
		this.Error(tsOpCode.OPERATION_DB_FAILED)
	}

	// TODO 此处返回的数据是用户提交的数据，而不是更新后的数据
	this.Success(model)
}

func (this *BaseController) CommDel(model sys.CommStructInterface) {
	id, _ := this.GetInt64(":id")
	if id < 1 {
		this.Error(tsOpCode.DATA_NOT_EXIST)
	}

	model.SetId(id)
	model.SetDel()

	if err := tsDb.NewDbBase().DbUpdate(model, "Deleted"); err != nil {
		logs.Error("DbDel error:", err)
		this.Error(tsOpCode.OPERATION_DB_FAILED)
	}
	this.Success(nil)
}

//获取节点和权限
func GetNavPermission(oAdmin sys.Admin) (oNav interface{}, oRole interface{}, createRoles interface{}, err error) {

	db := tsDb.NewDbBase()

	roles := strings.Split(oAdmin.Role, ",")
	creatRoles := tsString.CoverStringToArray(oAdmin.CreateRole, ",", false)
	var ids []string
	for _, v := range roles {
		if v != "" {
			ids = append(ids, v)
		}
	}

	if len(ids) > 0 {

		var role sys.Role
		role_list, err := db.DbInIds(&role, "Id", ids)

		if err != nil {
			return oNav, oRole, createRoles, err
		}

		if len(role_list) > 0 {

			temp := ""
			for _, v := range role_list {
				if v["Node"].(string) != "" {
					temp += v["Node"].(string)
				}
			}
			if temp != "" {
				temp = strings.Trim(temp, ",")

				var oNode sys.Node
				list, err := db.DbInIdsOrder(&oNode, "Id", strings.Split(temp, ","), []string{"sort"})
				if err != nil {
					return oNav, oRole, createRoles, err
				}

				temp = ""
				for _, v := range list {
					if v["ParentTree"].(string) != "" {
						temp += strings.Trim(v["ParentTree"].(string), ",") + ","
					}
				}

				if temp != "" {
					temp = strings.Trim(temp, ",")
					oNav, err = db.DbInIdsOrder(&oNode, "Id", strings.Split(temp, ","), []string{"sort"})
					if err != nil {
						return oNav, oRole, createRoles, err
					}
				}
			}
			//设置权限数据
			oRole = role_list
		}

	}
	if len(creatRoles) > 0 {
		var role sys.Role
		createRoles, err = db.DbInIds(&role, "Id", creatRoles)
	}
	return oNav, oRole, createRoles, nil
}

type CommonAddHandler func(tsConn *tsDb.DbBase, model sys.CommStructInterface)

type Filter struct {
	Key   string
	Value string
}

//过滤条件
func (this *BaseController) FilterToCondition(filterStr string) (status bool, c []*orm.Condition) {
	//[{key:xxx,value:[xxx, xxx]},...]
	status = true
	res := []Filter{}
	if filterStr == "" {
		return
	}
	err := json.Unmarshal([]byte(filterStr), &res)
	if err != nil {
		status = false
		return
	}

	//c = &orm.NewCondition()
	for _, kv := range res {
		key := kv.Key
		valueStr := kv.Value
		if key == "" || valueStr == "" {
			continue
		}
		values := strings.Split(kv.Value, ",")
		t := orm.NewCondition()
		t = t.Or(key, values)
		c = append(c, t)
	}
	return
}

func (this *BaseController) FilterToSQL(filterStr string, as string) (status bool, or []string) {
	//[{key:xxx,value:[xxx, xxx]},...]
	status = true
	res := []Filter{}
	if filterStr == "" {
		return
	}
	err := json.Unmarshal([]byte(filterStr), &res)
	if err != nil {
		status = false
		return
	}

	for _, kv := range res {
		key := kv.Key
		valueStr := kv.Value
		if key == "" || valueStr == "" {
			continue
		}
		values := strings.Split(kv.Value, ",")
		sql := "("
		vL := len(values)
		for i, value := range values {
			kv.Key = tsString.CoverCamelToSnake(kv.Key)
			sql += fmt.Sprintf("%v%v=`%v`", as, kv.Key, value)
			if i < vL-1 {
				sql += " or "
			}
		}
		sql += ")"
		or = append(or, sql)
	}
	return
}

func (this *BaseController) HandleAddReq(adminId, createAdminId int64, channelId string, req dto.CommReqInterface, model sys.CommStructInterface) {
	if err := this.ParseForm(req); err != nil {
		logs.Error("ParseForm:", err)
		this.Error(tsOpCode.OPERATION_REQUEST_FAILED)
	}

	// 校验输入
	if err := tsValid.Validate(req); err != nil {
		logs.Error("Validate error:", err)
		this.Error(tsOpCode.OPERATION_REQUEST_FAILED, err.Error())
	}

	hash := req.GetHash()
	reqJson, _ := tsJson.ToString(req)
	logs.Debug("CommAdd CheckRequest:", reflect.TypeOf(req).String(), reqJson)

	// 校验输入，通过hash达到避免重复插入
	if !req.CheckRequest() || len(hash) == 0 {
		logs.Warn("CommAdd CheckRequest error:", reflect.TypeOf(req).String(), reqJson)
		this.Error(tsOpCode.OPERATION_REQUEST_FAILED)
	}

	// 对一些字段进行赋值、校正
	req.SetDefault()

	// 将请求数据转化为数据库层对象
	model.FromReq(adminId, createAdminId, channelId, req)
	model.SetCreateTime()

	// 是否重复提交
	controllerName, _ := this.GetControllerAndAction()
	redisKey := fmt.Sprintf("%s-%d-%s", controllerName, this.AdminId, hash)
	if v, e := tsRedis.Get(redisKey); e != nil {
		logs.Warn("redis Get:", redisKey, " error:", e)
	} else if len(v) > 0 {
		// v = tsTime.CurrSe()
		this.Error(tsOpCode.REPEATED_SUBMIT)
	}
	// 插入数据
	go func() {
		_ = tsRedis.Set(redisKey, tsTime.CurrSe(), beego.AppConfig.DefaultInt64("repeatInsertExp", 60))
	}()
}

func (this *BaseController) HandleEditReq(req dto.CommReqInterface, model sys.CommStructInterface) {
	id, _ := this.GetInt64("Id", 0)
	if id < 1 {
		this.Error(tsOpCode.DATA_NOT_EXIST)
	}

	if err := this.ParseForm(req); err != nil {
		reqJson, _ := tsJson.ToString(req)
		logs.Warn("CommEdit CheckRequest error:", reflect.TypeOf(req).String(), reqJson)
		this.Error(tsOpCode.OPERATION_REQUEST_FAILED)
	}

	// 校验输入
	if err := tsValid.Validate(req); err != nil {
		logs.Error("Validate error:", err)
		this.Error(tsOpCode.OPERATION_REQUEST_FAILED, err.Error())
	}

	if !req.CheckRequest() {
		logs.Error("CheckRequest error:", req)
		this.Error(tsOpCode.OPERATION_REQUEST_FAILED)
	}

	// 将请求数据转化为数据库层对象
	model.FromReq(0, 0, "", req)
	model.SetUpdateTime()
	model.SetId(id)
}
