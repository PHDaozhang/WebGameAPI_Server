package sys

import (
	"fmt"
	"github.com/astaxie/beego/orm"
)

//用户表模型
type Node struct {
	Id          int64
	ParentId    int64
	ParentTree  string
	Icon        string
	Sort        int64
	Name        string
	Url         string
	LangCn      string `orm:"description(简体)"`
	LangTw      string `orm:"description(繁体)"`
	LangUs      string `orm:"description(英语)"`
	Description string
}

func init() {
	orm.RegisterModel(new(Node))
}

func (this *Node) TableName() string {
	return "system_node"
}

/************************************************************/

func (this *Node) Del() (err error) {

	op := orm.NewOrm()
	res, err := op.Raw("DELETE FROM `system_node` WHERE `parent_tree` Like ?", "%"+fmt.Sprintf(",%d,", this.Id)+"%").Exec()

	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	_, err = res.RowsAffected()
	return

}
