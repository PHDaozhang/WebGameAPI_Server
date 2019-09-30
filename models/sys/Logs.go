package sys

import (
	"github.com/astaxie/beego/orm"
	"tsEngine/tsPagination"
)

//日志表模型
type Logs struct {
	Id          int64
	AdminId     int64
	TemplateId  int64
	Mode        int64
	Action      int64
	Pass        int
	Ip          string
	Content     string
	Description string
	CreateTime  uint64
}

func init() {
	orm.RegisterModel(new(Logs))
}

func (this *Logs) TableName() string {
	return "system_logs"
}

/************************************************************/

func (this *Logs) List(page int64, pageSize int64, keyword string) (data []orm.Params, pagination *tsPagination.Pagination, err error) {

	op := orm.NewOrm().QueryTable(this)

	if keyword != "" {
		op = op.Filter("Content__icontains", keyword)
	}

	count, _ := op.Count()

	pagination = tsPagination.NewPagination(page, pageSize, count)

	op = op.Limit(pageSize, pagination.GetOffset())

	op = op.OrderBy("-Id")

	_, err = op.Values(&data)

	return data, pagination, err
}
