package sys

import (
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/astaxie/beego/validation"
	"strings"
	"tsEngine/tsCrypto"
	"tsEngine/tsDb"
	"tsEngine/tsOpCode"
	"tsEngine/tsPagination"
	"tsEngine/tsString"
	"tsEngine/tsTime"
	"web-game-api/conf"
	"web-game-api/core/consts"
	"web-game-api/core/utility"
	"web-game-api/models/dto"
)

// 管理员表模型
type Admin struct {
	Id           int64  `orm:"description(编号)"`
	Username     string `orm:"description(用户名，登陆账户)"`
	WithdrawPwd  string `orm:"description(提现密码)"`
	Password     string `orm:"description(用户密码)"`
	Role         string `orm:"description(角色)"`
	CreateRole   string `orm:"size(2048);description(可创建角色列表（,分割）)"`
	Name         string `orm:"description(真实姓名)"`
	Sex          int64  `orm:"description(性别 1：男 2：女)"`
	Birthday     uint64 `orm:"description(出生日期)"`
	Photo        string `orm:"description(照片)"`
	Email        string `orm:"description(邮箱地址)"`
	Address      string `orm:"description(员工地址)"`
	IdentityId   string `orm:"description(身份证号)"`
	ContactInf   string `orm:"description(联系方式)"`
	Mobile       string `orm:"description(手机号码)"`
	LoginIp      string `orm:"description(登陆IP)"`
	Status       int64  `orm:"description(1:启用 2：停用)"`
	Note         string `orm:"description(备注)"`
	ParentId     int64  `orm:"description(父节点)"`
	ParentTree   string `orm:"description(父节点树)"`
	AgentId      int64  `orm:"description(所属代理ID)"`
	AdminType    int64  `orm:"description(1代理 2推广 3币商 4子账户 5平台账户，用于对账)"`
	AppId        string `orm:"description(绑定的APP_ID)"`
	AppAccount   string `orm:"description(绑定的账户)"`
	AppSubSecret string `orm:"description(cloud提供的secret)"`
	AppRemark    string `orm:"description(app对应的名字,用于备注展示)"`
	PolicyId     int64  `orm:"description(关联策略组)"`
	CreateTime   uint64 `orm:"description(创建时间)"`
	LoginTime    uint64 `orm:"description(登陆时间)"`
	UpdateTime   uint64 `orm:"description(更新时间)"`
	Deleted      int    `orm:"description(删除)"`
}

type AdminAccount struct {
	Id               int64  // 编号
	Username         string // 用户名
	Password         string // 用户密码
	Role             string // 角色
	CreateRole       string // 角色
	Name             string // 真实姓名
	Sex              int64  // 性别 1：男 2：女
	Birthday         uint64 // 出生日期
	Photo            string // 照片
	Email            string // 邮箱地址
	Address          string // 员工地址
	IdentityId       string // 身份证号
	ContactInf       string // 联系方式
	Mobile           string // 手机号码
	LoginIp          string // 登陆IP
	Status           int64  // 1:启用 2：停用
	Note             string // 备注
	ParentId         int64  // 父节点
	ParentTree       string // 父节点树
	AgentId          int64  // 所属代理ID
	AdminType        int64  // 1代理 2推广 3币商 4子账户
	CreateTime       uint64 // 创建时间
	LoginTime        uint64 // 登陆时间
	UpdateTime       uint64 // 更新时间
	AdminId          int64
	Scores           int64
	TotalBuyScores   int64
	TotalSellScores  int64
	BuyScoreRate     float64
	Gold             int64
	TotalGold        int64
	TotalWithdraw    int64
	TotalTax         int64
	TotalContribute  int64
	Level            int64
	ReturnRate       float64
	SettleType       int
	SettleRate       int
	SettlePrice      int
	TotalCharge      int64
	TotalWater       int64
	TotalCount       int64
	TotalPerformance int64
}

type AdminSlim struct {
	Id         int64  // 编号
	Username   string // 用户名
	Name       string // 真实姓名
	CreateRole string
	ParentId   int64  // 父节点
	AdminType  int64  // 1代理 2推广 3币商 4子账户
	AgentId    int64  // 所属代理ID
	CreateTime uint64 // 创建时间
}

func init() {
	orm.RegisterModel(new(Admin))
}

func (this *Admin) TableName() string {
	return "system_admin"
}

// 内部方法：映射排序查询字段
func (this *Admin) sqlMapToCloum(key string) string {
	maps := map[string]string{
		"Scores":       "B.scores",
		"BuyScoreRate": "B.buy_score_rate",
		"ReturnRate":   "B.return_rate",
	}

	if v, ok := maps[key]; ok {
		return v
	} else {
		return key
	}
}

// 获取全部用户
func (this *Admin) List(page int64, pageSize int64, parentId, startTime, endTime string, adminType int64) (data []orm.Params, pagination *tsPagination.Pagination, err error) {

	op := orm.NewOrm().QueryTable(this)

	if parentId != "" {
		op = op.Filter("ParentId", parentId)
	}

	if adminType > 0 {
		op = op.Filter("AdminType", adminType)
	}

	if startTime != "" {
		op = op.Filter("CreateTime__gte", startTime)
	}
	if endTime != "" {
		op = op.Filter("CreateTime__lte", endTime)
	}

	if this.Username != "" {
		op = op.Filter("Username__icontains", this.Username)
	}

	if this.Name != "" {
		op = op.Filter("Name__icontains", this.Name)
	}

	if this.ContactInf != "" {
		op = op.Filter("ContactInf__icontains", this.ContactInf)
	}

	if this.Mobile != "" {
		op = op.Filter("Mobile", this.Mobile)
	}

	count, _ := op.Count()

	pagination = tsPagination.NewPagination(page, pageSize, count)

	op = op.Limit(pageSize, pagination.GetOffset())

	op = op.OrderBy("-Id")

	_, err = op.Values(&data)

	return data, pagination, err
}

// 计算count
type CountInfo struct {
	Count int64
}

// 合并查询用户表和账户表
func (this *Admin) ListOfAdminAccount(req dto.ReqSearch, parentId string, adminType int64, accountType int, accountCond *utility.AccountCond) (data []AdminAccount, pagination *tsPagination.Pagination, err error) {
	var conds []interface{}
	req.Page = utility.IfInt64(req.Page > 0 && req.Page != -1, req.Page, 1)
	req.PageSize = utility.IfInt64(req.PageSize > 0 && req.Page != -1, req.PageSize, 20)

	qbSelect, _ := orm.NewQueryBuilder(consts.CloudDBEngin)
	qbSelect.Select("A.*")
	qbFrom, _ := orm.NewQueryBuilder(consts.CloudDBEngin)
	qbFrom.From("system_admin AS A")

	qbWhere, _ := orm.NewQueryBuilder(consts.CloudDBEngin)
	qbWhere.Where("1=1 ").And("A.deleted=0")
	qbWhere.And("A.parent_tree LIKE ? ")
	conds = append(conds, "%,"+tsString.FromInt64(accountCond.AdminId)+",%")
	if accountCond.AdminId != conf.SystemAdminId {
		qbWhere.And("A.id!=?")
		if adminType > 0 && accountType != consts.AdminTypePromoter {
			qbWhere.And("(A.agent_id=? OR A.parent_id=?)")
		} else {
			qbWhere.And("A.agent_id=?").And("A.parent_id=?")
		}
		conds = append(conds, accountCond.AdminId, accountCond.AgentId, accountCond.AdminId)
	}
	//if parentId != "" {
	//	qbWhere.And("A.parent_id = ? ")
	//	conds = append(conds, parentId)
	//}
	if adminType > 0 {
		qbWhere.And("A.admin_type = ? ")
		conds = append(conds, adminType)
	}

	if req.Keyword != "" {
		keyword := "%" + req.Keyword + "%"
		exWhere, exArgs := utility.JointAdminCond(accountType, keyword)
		qbWhere.And("(A.username LIKE ? OR A.name LIKE ? OR A.note LIKE ? " + exWhere + ") ")
		conds = append(conds, keyword, keyword, keyword, exArgs)
	}

	if accountCond.AdminName != "" {
		qbWhere.And("A.username = ? ")
		conds = append(conds, accountCond.AdminName)
	}

	if accountCond.Name != "" {
		qbWhere.And("A.name = ? ")
		conds = append(conds, accountCond.Name)
	}
	if accountCond.Status > 0 {
		qbWhere.And("A.status = ? ")
		conds = append(conds, accountCond.Status)
	}
	if accountCond.SettleType != "" {
		settleTypes := strings.Split(accountCond.SettleType, ",")
		or := ""
		for i, s := range settleTypes {
			or += "C.settle_type=?"
			if i < len(settleTypes)-1 {
				or += " OR "
			}
			conds = append(conds, s)
		}
		qbWhere.And("(" + or + ") ")
	}
	if accountCond.Role != "" {
		qbWhere.And("A.role=? ")
		conds = append(conds, accountCond.Role)
	}
	if req.BeginTime > 0 {
		qbWhere.And("A.create_time >= ? ")
		conds = append(conds, req.BeginTime)
	}

	if req.EndTime > 0 {
		qbWhere.And("A.create_time <= ? ")
		conds = append(conds, req.EndTime)
	}

	var c CountInfo
	if err := orm.NewOrm().Raw(`SELECT COUNT(*) As count `+qbFrom.String()+" "+qbWhere.String(), conds).QueryRow(&c); err != nil {
		logs.Warn(err.Error())
	}
	pagination = tsPagination.NewPagination(req.Page, req.PageSize, c.Count)
	logs.Trace(c)

	qb3, _ := orm.NewQueryBuilder(consts.CloudDBEngin)

	if req.Sort != "" {
		qb3.OrderBy(utility.FormatSort(req.Sort, utility.IfString(strings.Contains(req.Sort, "CreateTime"), "A.", "B.")))
	} else {
		qb3.OrderBy(" A.id DESC")
	}

	//qb3.Limit(int(req.PageSize)).Offset(int(pagination.GetOffset()))
	if req.PageSize != -1 && req.Page != -1 {
		qb3.Limit(int(req.PageSize)).Offset(int(req.PageSize * (req.Page - 1)))
	}

	sql := qbSelect.String() + qbFrom.String() + qbWhere.String() + qb3.String()
	_, err = orm.NewOrm().Raw(sql, conds).QueryRows(&data)

	logs.Trace("获取账号列表：", sql, data)

	return data, pagination, err
}

// 检测账号唯一性
func (this *Admin) CheckUsernameOnly(userName string, adminId int64) (isOnly bool) {
	op := orm.NewOrm().QueryTable(this)

	if userName != "" {
		cond := orm.NewCondition().And("Username", userName)

		if adminId > 0 {
			cond = cond.AndNot("Id__in", adminId)
		}

		count, _ := op.SetCond(cond).Count()
		if count <= 0 {
			return true
		}
	}

	return
}

// 检测手机号唯一性
func (this *Admin) CheckPhoneOnly(phoneNo string) (isOnly bool) {
	db := tsDb.NewDbBase()
	admin := Admin{}
	if phoneNo != "" {
		count, err := db.DbCount(&admin, "Mobile", phoneNo)
		if err == nil {
			if count <= 0 {
				isOnly = true
			}
		}
	}

	return
}

// 检测联系的方式唯一性
func (this *Admin) CheckContactInfOnly(ContactInf string) (isOnly bool) {
	db := tsDb.NewDbBase()
	admin := Admin{}
	if ContactInf != "" {
		count, err := db.DbCount(&admin, "ContactInf", ContactInf)
		if err == nil {
			if count <= 0 {
				isOnly = true
			}
		}
	}

	return
}

func (this *Admin) GetCount(field string, value ...interface{}) (count int64, err error) {

	op := orm.NewOrm().QueryTable(this)
	op = op.Filter(field, value)
	op = op.Exclude("Id", this.Id)
	count, err = op.Count()
	return count, err
}

func (this *Admin) GetFilterCount(field string, value interface{}) (count int64, err error) {

	op := orm.NewOrm().QueryTable(this)
	op = op.Filter(field, value)
	op = op.Exclude("Id", this.Id)
	count, err = op.Count()
	return count, err
}

// 展示自己孩子，只查询某些字段 - 廋身版本（废弃）
func (this *Admin) ListChildSlim(page int64, pageSize int64, startTime, endTime, keyWord string) (data []orm.Params, pagination *tsPagination.Pagination, err error) {

	op := orm.NewOrm().QueryTable(this)
	cond := orm.NewCondition()

	if startTime != "" {
		cond = cond.And("CreateTime__gte", startTime)
	}
	if endTime != "" {
		cond = cond.And("CreateTime__lte", endTime)
	}

	if keyWord != "" {
		cond = cond.AndCond(orm.NewCondition().And("Name__contains", keyWord).Or("Username__contains", keyWord))
	}

	op = op.SetCond(cond)

	count, _ := op.Count()

	pagination = tsPagination.NewPagination(page, pageSize, count)

	op = op.Limit(pageSize, pagination.GetOffset())

	op = op.OrderBy("-Id")

	_, err = op.Values(&data, "Id", "Username", "Name", "ContactInf", "Mobile", "LoginTime", "Status", "CreateTime", "ParentId", "AdminType", "Note")

	return data, pagination, err
}

// 合并查询用户表和账户表
func (this *Admin) GetAdminJoinAccountById(id int64) (data AdminAccount, err error) {

	qbSelect, _ := orm.NewQueryBuilder(consts.CloudDBEngin)
	qbSelect.Select("A.*")
	qbFrom, _ := orm.NewQueryBuilder(consts.CloudDBEngin)
	qbFrom.From("system_admin AS A")

	qbWhere, _ := orm.NewQueryBuilder(consts.CloudDBEngin)
	qbWhere.Where("1=1 ").And("A.id=?")

	sql := qbSelect.String() + qbFrom.String() + qbWhere.String()
	err = orm.NewOrm().Raw(sql, id).QueryRow(&data)

	logs.Trace("获取账号列表：", qbSelect.String(), data)

	return data, err
}

// 根据ID获取当前用户信息
func (this *Admin) GetAdminById(adminId, realAdminId int64) (data AdminAccount, err error) {
	sql := "SELECT A.* FROM system_admin A WHERE 1 = 1"

	if adminId > 0 {
		sql += fmt.Sprintf(" AND A.id = %d", adminId)
	}

	if realAdminId > 0 {
		sql += " AND A.parent_tree like '%" + fmt.Sprintf(",%d,", realAdminId) + "%'"
	}

	err = orm.NewOrm().Raw(sql).QueryRow(&data)
	logs.Trace("子渠道数据AdminId：", adminId, data)

	return
}

// 根据父ID获取全部子渠道id
func (this *Admin) GetChildAdminId(adminId int64) (items []int64) {
	data := []Admin{}

	o := orm.NewOrm()
	sql := "SELECT id FROM system_admin WHERE 1=1"
	sql += " AND parent_tree like '%" + fmt.Sprintf("%d", adminId) + "%'"
	_, err := o.Raw(sql).QueryRows(&data)

	if err == nil {
		// 将子渠道的数据取出
		for i := range data {
			items = append(items, data[i].Id)
		}
	}

	logs.Trace("子渠道数据AdminId：", adminId, items)
	return
}

// 根据ID获取当前渠道一条线上全部用户
func (this *Admin) GetChildChannelInfo(adminId int64) (data []orm.Params) {
	op := orm.NewOrm().QueryTable(this)

	op = op.Filter("ParentTree__contains", fmt.Sprintf(",%d,", adminId))

	_, err := op.Values(&data, "Name", "Username")
	if err != nil {
		logs.Error(err)
	}

	logs.Trace("子渠道数据：", adminId, data)
	return
}

// 获取全部推广渠道
func (this *Admin) GetChildInfoByParentId(parentId string) (data []orm.Params) {
	op := orm.NewOrm().QueryTable(this)

	op = op.Filter("ParentId", parentId).Filter("AdminType", 2)

	_, err := op.Values(&data, "Name", "Username")
	if err != nil {
		logs.Error(err)
	}

	logs.Trace("子渠道数据：", parentId, data)
	return
}

/**
合并校验请求admin传入的参数
*/
func (this *Admin) validateAdmin(o Admin, realMob bool,reqPwd string) int64 {
	//****************************************************
	// 数据验证

	valid := validation.Validation{}

	// 用户名验证
	valid.Required(o.Username, "Username").Message("10010")
	valid.MinSize(o.Username, 2, "UserNameMin").Message("10011")
	valid.MaxSize(o.Username, 20, "UserNameMax").Message("10012")
	valid.AlphaDash(o.Username, "UserNameAlphaDash").Message("10013")
	// 密码验证
	valid.Required(o.Password, "Password").Message("10014")
	valid.MinSize(o.Password, 6, "PasswordMin").Message("10015")
	valid.MaxSize(o.Password, 50, "PasswordMax").Message("10016")
	valid.Range(int(o.Status), 1, 2, "Status").Message("10018")

	if o.Name != "" {
		valid.MaxSize(o.Name, 20, "Name").Message("10019")
	}

	if o.Mobile != "" && realMob {
		valid.Mobile(o.Mobile, "Mobile").Message("%v", tsOpCode.MOBILE_ERROR)
	}

	if valid.HasErrors() {
		// 如果有错误信息，证明验证没通过
		// 打印错误信息
		for _, err := range valid.Errors {
			return tsString.ToInt64(err.Message)
		}
	}

	return 0
}

func (this *Admin) GetAdminByName(name string) (err error) {
	this.Name = name
	err = tsDb.NewDbBase().DbGet(this, "Name")
	return
}

//func (this *Admin) DeleteAdmin(adminId int64) (err error) {
//	qbAdmin, _ := orm.NewQueryBuilder(consts.CloudDBEngin)
//	qbAdmin.Update("system_admin").Set("deleted=1").Where("id=?")
//	_, err = orm.NewOrm().Raw(qbAdmin.String()).SetArgs(adminId).Exec()
//	if err != nil {
//		logs.Error("[DeleteAdmin]system_admin:DB Error: ", err)
//		return
//	}
//	return
//}

func (this *Admin) DeleteAdmin(adminId, realAdminId int64) (err error) {
	qbAdmin, _ := orm.NewQueryBuilder(consts.CloudDBEngin)
	qbAdmin.Update("cloud_sys_admin").Set("deleted=1").Where("id=?").And("parent_tree like " + "'%" + fmt.Sprintf(",%d,", realAdminId) + "%'")
	res, err := orm.NewOrm().Raw(qbAdmin.String()).SetArgs(adminId).Exec()
	if err != nil {
		logs.Error("[DeleteAdmin]cloud_sys_admin:DB Error: ", err)
		return
	}

	effectRows, _ := res.RowsAffected()
	logs.Trace("更新的行数为:", effectRows)
	if effectRows == 0 {
		logs.Error("[DeleteAdmin]cloud_sys_admin:DB Error: 记录不存在")
		return
	}

	qbAcc, _ := orm.NewQueryBuilder(consts.CloudDBEngin)
	qbAcc.Update("cloud_data_account").Set("deleted=1").Where("admin_id=?")
	_, err = orm.NewOrm().Raw(qbAcc.String()).SetArgs(adminId).Exec()
	if err != nil {
		logs.Error("[DeleteAdmin]cloud_data_account:DB Error: ", err)
		return
	}
	return
}

/**
- 用户单独显示服务前端下拉框选择不同代理
*/
type PromoteAdminInfo struct {
	Id       int64
	Username string
}

/**
- 获取代理，全部推广adminId
*/
func (this *Admin) GetAllSonPromoteAdminId(realAdminId int64) (data []PromoteAdminInfo, ids []string, err error) {

	var conds []interface{}
	var SQL string

	// 如果是超管，则查询全部数据（这里要区分）- 不管有没有渠道，都查询出来就好了
	if realAdminId == -1 {
		SQL = `SELECT id, username FROM system_admin WHERE 1=1 AND id <> ?`
		conds = append(conds, conf.SystemAdminId)
	} else {
		// 这里匹配父节点，能匹配到自己
		SQL = `
SELECT
	id, username  FROM  system_admin
	WHERE parent_tree LIKE ? AND (id = ? OR (id <> ? AND admin_type = ?))
`
		conds = append(conds, "%,"+tsString.FromInt64(realAdminId)+",%", realAdminId, realAdminId, consts.AdminTypePromoter)
	}

	_, err = orm.NewOrm().Raw(SQL).SetArgs(conds).QueryRows(&data)
	if err != nil {
		logs.Error(err)
	} else {
		for _, x := range data {
			ids = append(ids, tsString.FromInt64(x.Id))
		}
	}

	logs.Trace(SQL)

	return
}

/**
- 获取全部子节点账号
*/
type ChildIds struct {
	Id int64
}

func (this *Admin) GetAllChildAdminId(realAdminId int64) (data []ChildIds, idsStr []string, err error) {
	// 这里匹配父节点，能匹配到自己
	qb, _ := orm.NewQueryBuilder("mysql")
	qb.Select("id")
	qb.From("system_admin")
	qb.Where("parent_tree LIKE ?")
	qb.And("id=?")
	sql := qb.String()

	_, err = orm.NewOrm().Raw(sql).SetArgs("%,"+tsString.FromInt64(realAdminId)+",%", realAdminId).QueryRows(&data)
	if err != nil {
		logs.Error(err)
	} else {
		for _, x := range data {
			idsStr = append(idsStr, tsString.FromInt64(x.Id))
		}
	}

	return
}

func (this *Admin) UpdateRole(adminId int64, roles []int64) (err error) {
	db := tsDb.NewDbBase()
	this.Id = adminId
	err = db.DbGet(this)
	if err != nil {
		return
	}
	roleStr := tsString.CoverInt64Arr2String(roles)
	this.Role = "," + strings.Join(roleStr, ",") + ","
	err = db.DbUpdate(this, "Role")
	return
}

func (this *Admin) UpdateCreateRole(adminId int64, roles []int64) (err error) {
	db := tsDb.NewDbBase()
	this.Id = adminId
	err = db.DbGet(this)
	if err != nil {
		return
	}
	roleStr := tsString.CoverInt64Arr2String(roles)
	this.CreateRole = "," + strings.Join(roleStr, ",") + ","
	err = db.DbUpdate(this, "CreateRole")
	return
}

func (this *Admin) AppendCreateRole(adminId int64, roles ...string) (err error) {
	db := tsDb.NewDbBase()
	this.Id = adminId
	err = db.DbGet(this)
	if err != nil {
		return
	}
	this.CreateRole += "," + strings.Join(roles, ",") + ","
	err = db.DbUpdate(this, "CreateRole")
	return
}


func (this *Admin) GetSampleAdminById(adminId int64) (err error) {
	this.Id = adminId
	//wisp delete
	//data, _ := tsRedis.Get(fmt.Sprintf(consts.KeyCloudSysAdminById, this.Id))
	//if data != "" {
	//	err = json.Unmarshal([]byte(data), this)
	//	return
	//}
	err = tsDb.NewDbBase().DbGet(this, "Id")
	return
}


/**
更新登录账号和account账号
req 		- 前端传入的全部参数
PersonInfo  - 个人admin和account表数据
*/
func (this *Admin) EditAdmin(req dto.ReqEditAdmin, PersonInfo Admin, realAdminId int64) int64 {
	db := tsDb.NewDbBase()

	var err error
	o := Admin{Id: req.Id}
	err = db.DbGet(&o)
	logs.Trace("更新账号信息", o)

	if err != nil || !strings.Contains(o.ParentTree, fmt.Sprintf(",%d,", realAdminId)) {
		logs.Error(err)
		return tsOpCode.DATA_NOT_EXIST
	}

	oldUserName := o.Username
	oldPassword := o.Password
	oldMobile := o.Mobile
	oldContactInf := o.ContactInf

	o.ContactInf = req.ContactInf
	o.Name = req.Name
	o.Note = req.Note
	o.CreateRole = req.CreateRole
	o.Username = req.Username
	o.Status = req.Status
	o.UpdateTime = tsTime.CurrSe()
	if req.Role != "" {
		role, err := this.FormatRoleString(req.Role)
		if err != nil {
			return tsOpCode.OPERATION_DB_FAILED
		}
		if role != "" {
			o.Role = role
		}
	}

	// 校验必填参数
	errCode := this.validateAdmin(o, false, req.Password)
	if errCode > 0 {
		logs.Error(errCode)
		return errCode
	}

	// 如果密码有修改就修改
	newPassword := tsCrypto.GetMd5([]byte(req.Password + conf.PasswordSalt))
	if req.Password != "" && newPassword != oldPassword {
		o.Password = newPassword
	}

	if req.Username != "" && req.Username != oldUserName {
		// 检测账号唯一性
		isOnly := o.CheckUsernameOnly(o.Username, 0)
		if !isOnly {
			return tsOpCode.USERNAME_IS_EXISTS
		}
	}

	if req.Mobile != "" && req.Mobile != oldMobile {
		isOnly := o.CheckPhoneOnly(o.Mobile)
		if !isOnly {
			return tsOpCode.TEL_NUMBER_EXISTED
		}
	}

	if req.ContactInf != "" && req.ContactInf != oldContactInf {
		isOnly := o.CheckContactInfOnly(o.ContactInf)
		if !isOnly {
			return tsOpCode.CONTACT_INF_EXISITS
		}
	}

	logs.Trace("更新账号信息", o)

	_ = db.Transaction()
	defer db.TransactionEnd()

	err = db.DbUpdate(&o, "ContactInf", "Mobile", "Name", "Note", "Role", "CreateRole", "Username", "Password", "Status", "UpdateTime")
	if err != nil {
		db.SetRollback(true)
		logs.Error(err)
		return tsOpCode.OPERATION_DB_FAILED
	}

	// 子账号编辑完成，直接退出
	if o.AdminType == consts.AdminTypeSonuser {
		return tsOpCode.OPERATION_SUCCESS
	}

	// 修改推广合作模式和合作金币
	//if o.AdminType == consts.AdminTypePromoter {
	//	mp := PromoterData{AdminId: o.Id}
	//	err = db.ReadForUpdate(&mp, "AdminId")
	//	if err != nil {
	//		db.SetRollback(true)
	//		logs.Error(err)
	//		return tsOpCode.DATA_NOT_EXIST
	//	}
	//
	//	mp.SettleType = req.SettleType
	//	mp.SettlePrice = req.SettlePrice
	//	err = db.DbUpdate(&mp, "SettleType", "SettlePrice")
	//	if err != nil {
	//		db.SetRollback(true)
	//		logs.Error(err)
	//		return tsOpCode.ACCOUNT_ERROR
	//	}
	//}

	// 修改代理服务费率
	//if o.AdminType == consts.AdminTypeAgent {
	//	// 更新account表
	//	mA := Account{AdminId: o.Id}
	//
	//	// 加锁
	//	err = db.ReadForUpdate(&mA, "AdminId")
	//	if err != nil {
	//		db.SetRollback(true)
	//		logs.Error(err)
	//		return tsOpCode.DATA_NOT_EXIST
	//	}
	//
	//	mA.ServiceRateType = req.ServiceRateType
	//	// 默认固定费率
	//	if req.ServiceRateType == 0 {
	//		mA.ServiceRateType = 1
	//	}
	//	mA.ServiceRate = req.ServiceRate
	//	mA.ServiceRateDynamic = req.ServiceRateDynamic
	//	err = db.DbUpdate(&mA, "ServiceRateType", "ServiceRateDynamic", "ServiceRate")
	//	if err != nil {
	//		db.SetRollback(true)
	//		logs.Error(err)
	//		return tsOpCode.ACCOUNT_ERROR
	//	}
	//}

	return tsOpCode.OPERATION_SUCCESS
}


func (this *Admin) FormatRoleString(srcRole string) (destRole string, err error) {
	strArr := strings.Split(srcRole, ",")
	destArr := []string{}
	for _, s := range strArr {
		if s != "" {
			destArr = append(destArr, s)
		}
	}
	r := Role{}
	if !r.RolesExists(destArr) {
		return "", errors.New("[FormatRoleString]Roles invalid,please check")
	}
	destRole = "," + strings.Join(destArr, ",") + ","
	return destRole, err
}

/**
初始化登录账号
req 		- 前端传入的全部参数
PersonInfo  - 个人admin和account表数据
*/
func (this *Admin) InitAdmin(req dto.ReqAddAdmin, PersonInfo Admin, realMob, isChild bool) (int64, int64) {
	db := tsDb.NewDbBase()

	var err error
	o := Admin{
		Username:   req.Username,
		Password:   req.Password,
		Status:     req.Status,
		Name:       req.Name,
		Mobile:     req.Mobile,
		ContactInf: req.ContactInf,
		AdminType:  req.AdminType,
		Note:       req.Note,
		ParentId:   utility.IfInt64(isChild, PersonInfo.AgentId, PersonInfo.Id), // 改为目前所有的员工账号都属于代理
		ParentTree: PersonInfo.ParentTree,
		CreateRole: req.CreateRole,
	}

	logs.Trace("个人缓存数据：%#v,%#v", req, PersonInfo)

	o.Role, err = this.FormatRoleString(req.Role)
	if err != nil {
		return tsOpCode.OPERATION_DB_FAILED, 0
	}

	// 过滤开发者账号
	if o.Username == beego.AppConfig.String("Username") {
		return tsOpCode.OPERATION_REQUEST_FAILED, 0
	}

	// 校验必填参数
	errCode := this.validateAdmin(o, realMob, req.Password)
	if errCode > 0 {
		return errCode, 0
	}

	o.Password = tsCrypto.GetMd5([]byte(o.Password + conf.PasswordSalt))
	o.CreateTime = tsTime.CurrSe()
	o.UpdateTime = o.CreateTime
	o.LoginTime = o.CreateTime

	if o.Username != "" {
		// 检测账号唯一性
		isOnly := o.CheckUsernameOnly(o.Username, 0)
		if !isOnly {
			return tsOpCode.USER_NAME_EXIST, 0
		}
	}

	if o.Mobile != "" {
		isOnly := o.CheckPhoneOnly(o.Mobile)
		if !isOnly {
			return tsOpCode.TEL_NUMBER_EXISTED, 0
		}
	}

	_ = db.Transaction()
	defer db.TransactionEnd()

	adminId, err := db.DbInsert(&o)
	if err != nil {
		db.SetRollback(true)
		logs.Error(err)
		return tsOpCode.OPERATION_DB_FAILED, 0
	}

	//更新ParentTree，链接自身的Id
	o.ParentTree += fmt.Sprintf("%d,", adminId)
	if req.AdminType == consts.AdminTypeAgent {
		o.AgentId = adminId
	} else {
		o.AgentId = PersonInfo.AgentId
	}
	err = db.DbUpdate(&o, "ParentTree", "AgentId")
	if err != nil {
		db.SetRollback(true)
		logs.Error(err)
		return tsOpCode.OPERATION_DB_FAILED, 0
	}

	// 非子账号
	if req.AdminType == consts.AdminTypeSonuser {
		return tsOpCode.OPERATION_SUCCESS, 0
	}


	return tsOpCode.OPERATION_SUCCESS, adminId
}