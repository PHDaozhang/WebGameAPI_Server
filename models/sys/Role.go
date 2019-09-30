package sys

import (
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"strings"
	"tsEngine/tsContain"
	"tsEngine/tsDb"
	"tsEngine/tsPagination"
	"tsEngine/tsString"
	"tsEngine/tsTime"
	"web-game-api/conf"
)

//用户表模型
type Role struct {
	CommStruct
	Name        string
	Permission  string `orm:"size(2048)"`
	Node        string `orm:"size(2048)"`
	System      string
	Description string `orm:"size(1024)"`
	Type        int    `orm:"size(1);description(1:平台默认，2:代理自建)"`
}

func init() {
	orm.RegisterModel(new(Role))
}

func (this *Role) TableName() string {
	return "system_role"
}

func (this *Role) List(adminId, page, pageSize int64, keyword string) (data []Role, pagination *tsPagination.Pagination, err error) {
	adminUid := conf.SystemAdminId
	qbSelect, _ := orm.NewQueryBuilder("mysql")
	qbSelect.Select("*")
	qbCount, _ := orm.NewQueryBuilder("mysql")
	qbCount.Select("COUNT(*) count")
	qbQuery, _ := orm.NewQueryBuilder("mysql")
	args := []interface{}{}
	qbQuery.From("system_role")
	qbQuery.Where("deleted=0")
	if adminId != adminUid {
		qbQuery.And("admin_id=?")
		args = append(args, adminId)
		//	qbQuery.And("type=?")
		//	args = append(args, 2)
	}
	if keyword != "" {
		qbQuery.And("name LIKE ?")
		args = append(args, "%"+keyword+"%")
	}
	c := CountInfo{}
	op := orm.NewOrm()
	err = op.Raw(qbCount.String()+" "+qbQuery.String()+" ", args).QueryRow(&c)
	if err != nil {
		return
	}
	pagination = tsPagination.NewPagination(page, pageSize, c.Count)
	qbOrder, _ := orm.NewQueryBuilder("mysql")
	qbOrder.OrderBy("id").Desc()
	qbOrder.Limit(int(pageSize))
	qbOrder.Offset(int(pagination.GetOffset()))
	_, err = op.Raw(qbSelect.String()+" "+qbQuery.String()+" "+qbOrder.String(), args).QueryRows(&data)
	if err != nil {
		return
	}
	return data, pagination, err

}

func (this *Role) VerifyRolePermission(adminId int64, permissions string) (pass bool, err error) {
	pass = false
	if adminId == conf.SystemAdminId {
		return true, nil
	}
	admin := Admin{}
	role := Role{}
	admin.Id = adminId
	db := tsDb.NewDbBase()
	err = db.DbGet(&admin)
	if err != nil {
		return
	}
	adminRole := admin.Role
	adminRoleArr := tsString.CoverStringToArray(adminRole, ",", false)
	adminRolePermission := ""
	for _, arp := range adminRoleArr {
		role.Id = tsString.ToInt64(arp)
		if err = db.DbGet(&role); err != nil {
			return
		}
		adminRolePermission += role.Permission
	}

	permissionsArr := tsString.CoverStringToArray(permissions, ",", false)
	adminPermissionsArr := tsString.CoverStringToArray(adminRolePermission, ",", false)
	for _, permission := range permissionsArr {
		if !tsContain.InArrayString(adminPermissionsArr, permission) || this.SpecialPermission(permission) {
			return
		}
	}
	pass = true
	return
}

func (this *Role) SpecialPermission(needle string) bool {
	return tsContain.InArrayString(strings.Split(conf.SpecialPermissions, ","), needle)
}

func (this *Role) CopyRoleTemplate(adminId int64) (ids []int64, err error) {
	oRoles := []Role{}
	qb, _ := orm.NewQueryBuilder("mysql")
	qb.Select("*")
	qb.From("system_role")
	qb.Where("type=1")
	o := orm.NewOrm()
	_, err = o.Raw(qb.String()).QueryRows(&oRoles)
	if err != nil {
		return
	}
	for _, oRole := range oRoles {
		iRoles := Role{
			CommStruct{
				0,
				adminId,
				adminId,
				tsTime.CurrSe(),
				0,
				0,
			},
			oRole.Name,
			oRole.Permission,
			oRole.Node,
			oRole.System,
			oRole.Description,
			2,
		}
		id, err2 := o.Insert(&iRoles)
		err = err2
		if err != nil {
			return
		}
		ids = append(ids, id)
	}
	return
}

func (this *Role) GetRoleList(adminId int64) (rl []Role, err error) {
	qb, _ := orm.NewQueryBuilder("mysql")
	qb.Select("*").From("system_role").Where("admin_id=?")
	_, err = orm.NewOrm().Raw(qb.String()).SetArgs(adminId).QueryRows(&rl)
	return
}


func (this *Role) GetRoleUseCount2(adminId int64, roleId string) (count int64) {
	o := orm.NewOrm()
	qb, _ := orm.NewQueryBuilder("mysql")
	qb.Select("COUNT(*) count").From("cloud_sys_admin").Where("(role LIKE ? OR create_role LIKE ?)").And("deleted=0").And("id != ?")
	like := "%," + roleId + ",%"
	c := CountInfo{}
	_ = o.Raw(qb.String(), like, like, adminId).QueryRow(&c)
	return c.Count
}


func (this *Role) RolesExists(roles []string) bool {
	if len(roles) == 0 {
		logs.Error("[Role][RolesExists]Input Roles is empty")
		return false
	}
	count, err := orm.NewOrm().QueryTable(this).Filter("id__in", roles).Count()
	if err != nil {
		logs.Error("[Role][RolesExists]DB Error:", err)
		return false
	}
	if int(count) == len(roles) {
		return true
	}
	return false
}
