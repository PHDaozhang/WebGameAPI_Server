package sys

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/orm"
	"tsEngine/tsDb"
	"tsEngine/tsRedis"
	"web-game-api/core/consts"
)

//用户表模型
type Mode struct {
	Id          int64
	ParentId    int64
	NodeId      int64
	Name        string
	Key         string
	Md5         string
	Type        int64
	Logs        int64
	Sort        int64
	Description string
}

func init() {
	orm.RegisterModel(new(Mode))
}

func (this *Mode) TableName() string {
	return "system_mode"
}

/************************************************************/

func (this *Mode) List(keyword string) (data []orm.Params, err error) {

	op := orm.NewOrm().QueryTable(this)

	if keyword != "" {
		cond := orm.NewCondition().And("Name__icontains", keyword).Or("Description__icontains", keyword).Or("ParentId__gt", -1)
		op = op.SetCond(cond)
	}

	op = op.OrderBy("Type", "Sort", "-Id")

	_, err = op.Values(&data)

	return data, err
}

func (this *Mode) GetModeByMD5(md5 string) (err error) {
	modeStr, err := tsRedis.Get(fmt.Sprintf(consts.KeyWEBAPISysModeByMD5, md5))
	if err != nil {
		this.Md5 = md5
		err = tsDb.NewDbBase().DbRead(this, "Md5")
		if err != nil {
			return err
		}
		err = nil
	} else {
		err = json.Unmarshal([]byte(modeStr), this)
		if err != nil {
			return err
		}
	}
	return
}
