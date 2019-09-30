package models

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"tsEngine/tsDb"
	"tsEngine/tsJson"
	"tsEngine/tsTime"
)

//操作日志
type HandleLogs struct {
	CommStruct
	Username   string `orm:"Description(操作人账号)"`
	TemplateId int64  `orm:"Description(模板ID)"`
	Ip         string `orm:"Description(IP)"`
	Params     string `orm:"Description(参数Json字符串)"`
}

func init() {
	orm.RegisterModel(new(HandleLogs))
}

func (this *HandleLogs) TableName() string {
	return "cloud_data_handle_logs"
}

func (this *HandleLogs) Log(templateId, agentId, adminId int64, username, ip string, params beego.M) {
	go func() {
		this.TemplateId = templateId
		this.AdminId = agentId
		this.Username = username
		this.CreateAdminId = adminId
		this.CreateTime = tsTime.CurrSe()
		this.Params = tsJson.ToJson(params)
		this.Ip = ip
		_, err := tsDb.NewDbBase().DbInsert(this)
		if err != nil {
			logs.Error("[HandleLog][Log]Db Insert Error:", err)
		}
	}()
}
