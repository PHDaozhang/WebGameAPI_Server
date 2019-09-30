package system

import (
	"github.com/astaxie/beego/logs"
	"gopkg.in/mgo.v2/bson"
	"tsEngine/tsDb"
	"tsEngine/tsOpCode"
	"web-game-api/models/dto"
	"web-game-api/models/sys"
)

type LogsController struct {
	PermissionController
}

// @Title log列表
// @Description log列表
// @Success 200 {object} admin.Logs
// @Param    Keyword    query    string    false  搜索词
// @Param    Sort       query    string    false  "排序(示例: 按照渠道正向排序 ‘ChannelId’，按照渠道反向排序 ‘-ChannelId’)。如果为空，默认按照Id反向排序"
// @Param    Page       query    string    true   页码
// @Param    PageSize   query    string    true   单页数据量
// @Param    BeginTime  query    string    false  过滤开始时间
// @Param    EndTime    query    string    false  过滤结束时间
// @router   /list [get]
func (this *LogsController) List() {
	var req dto.ReqSearch
	if err := this.ParseForm(&req); err != nil {
		this.Error(tsOpCode.OPERATION_REQUEST_FAILED)
	}

	o := sys.Logs{}
	items, pagination, err := o.List(req.Page, req.PageSize, req.Keyword)

	if err != nil {
		logs.Error(err)
		this.Error(tsOpCode.OPERATION_DB_FAILED)
	}

	this.Success(bson.M{"Items": items, "Pagination": pagination})
}

// @Title 清除log
// @Description 清除log
// @Success 200 {"Code":200,"Data":null}
// @router   /clear [get]
func (this *LogsController) Clear() {
	var oLogs sys.Logs
	db := tsDb.NewDbBase()
	_, err := db.DbDel(&oLogs, "Id__gt", "0")
	if err != nil {
		logs.Error(err)
		this.Error(tsOpCode.OPERATION_DB_FAILED)
	}
	this.Success(nil)
}
