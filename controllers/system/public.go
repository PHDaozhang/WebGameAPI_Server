//不需要权限判断时候所使用的模块
package system

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"gopkg.in/mgo.v2/bson"
	"tsEngine/tsCrypto"
	"tsEngine/tsDb"
	"tsEngine/tsFile"
	"tsEngine/tsOpCode"
	"tsEngine/tsRand"
	"tsEngine/tsRedis"
	"tsEngine/tsString"
	"tsEngine/tsTime"
	"web-game-api/conf"
	"web-game-api/core/consts"
	"web-game-api/models/sys"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/validation"
)

type PublicController struct {
	PermissionController
}

func (this *PublicController) Prepare() {
	this.CheckLogin()
}

// @Title 获取配置
// @Description desc
// @Success 200 {object} {}
// @Param    type    query    string    false   获取配置的类型
// @router /config [get]
//
func (this *PublicController) Config() {
	db := tsDb.NewDbBase()
	//获取数据信息
	types := this.GetString("type")

	switch types {
	case "Admin":
		o := sys.Admin{}
		//需要获取的字段
		fields := []string{"Id", "Name", "Role", "Sex", "Photo"}
		//排序方式
		order := []string{"Id"}
		list, _ := db.DbListFields(&o, fields, order)
		this.Success(list)
	case "Role":
		o := sys.Role{}
		list, _ := db.DbList(&o)
		this.Success(list)
	case "Mode":
		o := sys.Mode{}
		list, _ := db.DbList(&o)
		this.Success(list)
	case "Node":
		o := sys.Node{}
		order := []string{"Sort", "Id"}
		list, _ := db.DbListOrder(&o, order)
		this.Success(list)
	default:
		pageNum, _ := beego.AppConfig.Int64("PageNum")
		oRole := sys.Role{}
		//roleList, _ := db.DbList(&oRole)
		roleList, _ := oRole.GetRoleList(this.RealAdminId)

		oMode := sys.Mode{}
		modeList, _ := db.DbList(&oMode)

		order := []string{"Sort", "-Id"}
		oNode := sys.Node{}
		nodeList, _ := db.DbListOrder(&oNode, order)

		this.Success(bson.M{
			"PageNum": pageNum,
			"Role":    roleList, // 过滤后的角色
			//"Admin":         admin_list, // TODO 不理解
			"Mode": modeList,
			"Node": nodeList,
		})
	}
}

// @Title 管理员信息修改
// @Description 管理员信息修改
// @Success 200 {object} admin.Admin
// @Param    Username           formData    string    true   用户名
// @Param    Password           formData    string    true   密码
// @Param    ConfirmPassword    formData    string    true   确认密码
// @Param    Name               formData    string    true   昵称
// @Param    Sex                formData    int64     true   性别
// @Param    Birthday           formData    string    true   生日
// @Param    Email              formData    string    true   eMail
// @Param    IdentityId         formData    string    true   身份证号
// @Param    Mobile             formData    string    true   手机号
// @Param    Address            formData    string    true   地址
// @Param    Note               formData    string    true   签名
// @Param    Photo              formData    string    true   照片
// @router /info [post]
//
func (this *PublicController) Info() {
	req := struct {
		Username        string
		Password        string
		ConfirmPassword string
		Name            string
		Sex             int64
		Birthday        string
		Email           string
		IdentityId      string
		Mobile          string
		Address         string
		Note            string
		Photo           string
	}{}
	if err := this.ParseForm(&req); err != nil {
		this.Error(tsOpCode.OPERATION_REQUEST_FAILED)
	}
	//过滤开发者账号
	uid := conf.SystemAdminId
	if this.AdminId == uid {
		this.Error(tsOpCode.ACCOUNT_DENIDE)
	}
	//初始化
	db := tsDb.NewDbBase()
	o := sys.Admin{}
	o.Id = this.AdminId

	err := db.DbRead(&o)
	if err != nil {
		this.Error(tsOpCode.OPERATION_DB_FAILED)
	}

	oldPassword := o.Password

	o.Username = req.Username
	//过滤开发者账号
	if o.Username == beego.AppConfig.String("Username") {
		if this.AdminId == uid {
			this.Error(tsOpCode.ACCOUNT_DENIDE)
		}
	}

	o.Password = req.Password
	o.Name = req.Name
	o.Sex = req.Sex
	o.Birthday = tsTime.StringToSe(req.Birthday, 4)
	o.Email = req.Email
	o.IdentityId = req.IdentityId
	o.Mobile = req.Mobile
	o.Address = req.Address
	o.Note = req.Note

	//****************************************************
	//数据验证
	valid := validation.Validation{}
	//密码验证
	if o.Password != "" {
		valid.Required(o.Password, "Password").Message("10014")
		valid.MinSize(o.Password, 6, "PasswordMin").Message("10015")
		valid.MaxSize(o.Password, 50, "PasswordMax").Message("10016")
		if o.Password != req.ConfirmPassword {
			this.Error(tsOpCode.PASSWORD_CONFIRM_FAILED)
		}
		o.Password = tsCrypto.GetMd5([]byte(o.Password + conf.PasswordSalt))
	} else {
		o.Password = oldPassword
	}
	if o.Name != "" {
		valid.MaxSize(o.Name, 20, "Name").Message("10019")
	}
	if o.Sex > 0 {
		valid.Range(int(o.Sex), 1, 2, "Sex").Message("10020")
	}
	if o.Email != "" {
		valid.Email(o.Email, "Email").Message("10021")
	}
	if o.Mobile != "" {
		valid.Mobile(o.Mobile, "Mobile").Message("%v", tsOpCode.MOBILE_ERROR)
	}
	if valid.HasErrors() {
		// 如果有错误信息，证明验证没通过
		// 打印错误信息
		for _, err := range valid.Errors {
			logs.Warn("valid:", err)
			this.Error(tsString.ToInt(err.Message))
		}
	}
	if len(req.Photo) > 255 && req.Photo != o.Photo {
		filename := tsCrypto.GetMd5([]byte(fmt.Sprintf("%d%d", tsTime.CurrMs(), tsRand.RandInt(0, 10000))))
		path, err := tsFile.WriteImgFile("./static/upload/", filename, req.Photo)
		if err != nil {
			this.Error(tsOpCode.SAVE_UPLOAD_FILE_ERROR)
		}
		//上传成功后删除原始图片
		if o.Photo != "" {
			tsFile.DelFile("." + o.Photo) //删除文件
		}
		o.Photo = path
	}

	if o.Mobile != "" {
		count, err := o.GetFilterCount("Mobile", o.Mobile)
		if err != nil {
			this.Error(tsOpCode.OPERATION_DB_FAILED)
		}
		if count > 0 {
			this.Error(tsOpCode.CONTACT_INF_EXISITS)
		}
	}
	o.UpdateTime = tsTime.CurrSe()
	err = db.DbUpdate(&o, "Password", "Name", "Email", "Photo", "Sex", "Birthday", "Address", "IdentityId", "Mobile", "UpdateTime", "Note")
	if err != nil {
		this.Error(tsOpCode.OPERATION_DB_FAILED)
	}
	this.Success(o)
}

// @Title 获取管理员数据
// @Description 获取管理员数据
// @Success 200 {object} admin.Admin
// @router /admin [get]
//
func (this *PublicController) Admin() {
	o := sys.Admin{}
	//需要获取的字段
	fields := []string{"Id", "Name", "Role", "Sex", "Photo"}
	//排序方式
	order := []string{"Id"}
	db := tsDb.NewDbBase()
	items, err := db.DbListFields(&o, fields, order)
	if err != nil {
		this.Error(tsOpCode.OPERATION_DB_FAILED)
	}
	this.Success(items)
}

// @Title 修改密码
// @Description 修改密码
// @Success 200 {object} admin.Admin
// @Param    OldPassword      formData    string    true   旧密码
// @Param    NewPassword      formData    string    true   新密码
// @Param    ConfirmPassword  formData    string    true   新密码2
// @router   /modifypassword [put]
func (this *PublicController) ModifyPassword() {

	db := tsDb.NewDbBase()
	o := sys.Admin{
		Id: this.AdminId,
	}

	err := db.DbRead(&o)
	if err != nil {
		logs.Error(err)
		this.Error(tsOpCode.DATA_NOT_EXIST)
	}
	oldPassword := o.Password

	//过滤开发者账号
	if o.Username == beego.AppConfig.String("Username") {
		this.Error(tsOpCode.OPERATION_REQUEST_FAILED)
	}

	oPassword := this.GetString("OldPassword")
	oPassword = tsCrypto.GetMd5([]byte(oPassword + conf.PasswordSalt))
	if oldPassword != oPassword {
		this.Error(tsOpCode.PASSWORD_CONFIRM_FAILED)
	}
	NewPassword := this.GetString("NewPassword")
	ConfirmPassword := this.GetString("ConfirmPassword")
	o.Password = NewPassword

	//****************************************************
	//数据验证
	valid := validation.Validation{}
	//密码验证
	if o.Password != "" {
		valid.Required(o.Password, "Password").Message("%v", tsOpCode.PASSWORD_CANT_EMPTY)
		valid.MinSize(o.Password, 6, "PasswordMin").Message("%v", tsOpCode.PASSWORD_SHORTLY)
		valid.MaxSize(o.Password, 50, "PasswordMax").Message("%v", tsOpCode.PASSWORD_LONG)
		if o.Password != ConfirmPassword {
			this.Error(tsOpCode.PASSWORD_INCONSISTENT)
		}

		o.Password = tsCrypto.GetMd5([]byte(o.Password + conf.PasswordSalt))

	} else {
		o.Password = oldPassword
	}

	if valid.HasErrors() {
		// 如果有错误信息，证明验证没通过
		// 打印错误信息
		for _, err := range valid.Errors {
			this.Error(tsString.ToInt(err.Message))
		}
	}

	db.Transaction()
	defer db.TransactionEnd()

	o.UpdateTime = tsTime.CurrSe()
	err = db.DbUpdate(&o, "Password")
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
	this.Success(o)
}
