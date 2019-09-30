package api

import (
	"fmt"
	"github.com/astaxie/beego/orm"
	"tsEngine/tsPagination"
)

type Agent struct{
	Id			int64
	Name		string `orm:"name"`
	AppKey		string
	Desc		string `orm:"desc"`
}

func init(){
	orm.RegisterModel(new(Agent))

}

func (this *Agent)TableName()string {
	return "agent"
}

type CountInfo struct {
	Count int64
}

//
func (this *Agent)List(page,pageSize int64,keyword string,filterKey string)(data []Agent,pagination *tsPagination.Pagination,err error) {
	qbSelect, _ := orm.NewQueryBuilder("mysql")
	qbSelect.Select("*")
	qbCount, _ := orm.NewQueryBuilder("mysql")
	qbCount.Select("COUNT(*) count")
	qbQuery, _ := orm.NewQueryBuilder("mysql")
	args := []interface{}{}
	qbQuery.From("agent")
	//qbQuery.Where("deleted=0")
	qbQuery.Where("id>0")
	if keyword != "" && keyword != "All" {
		fieldName := "name"
		if filterKey == "Name" {
			fieldName = "name"
		} else if filterKey == "Id" {
			fieldName = "id"
		} else if filterKey == "AppKey" {
			fieldName = "app_key"
		}

		qbQuery.And(  fieldName + " LIKE ?")
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
	qbOrder.OrderBy("id").Asc()
	qbOrder.Limit(int(pageSize))
	qbOrder.Offset(int(pagination.GetOffset()))

	queryStr := qbSelect.String()+" "+qbQuery.String()+" "+qbOrder.String()

	fmt.Println("queryStr:" + queryStr)

	_, err = op.Raw(queryStr, args).QueryRows(&data)
	if err != nil {
		return
	}
	return data, pagination, err
}