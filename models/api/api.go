package api

import (
	"github.com/astaxie/beego/orm"
	"time"
)




//无效
//type Game struct {
//	Agent		uint32 	 `orm:"column(Agent);pk"` // 设置主键 ;
//	Group		uint32;
//	Appkey		string;
//}
//
//
////无效
//type Order struct {
//	Id			uint32;
//	Order		string;
//	Timestamp	orm.DateField;
//	Agent		uint32;
//	Account		string;
//	Money       uint32;
//	Type		int8;
//	Status		int8;
//	Ac_Time		orm.DateField;
//}
//
//
////无效
//type User struct {
//	Id			uint32;
//	Account		string;
//	Agent		uint32;
//	Uid			string;
//}

type Order_log struct {
	Id			uint32
	Order		string
	Timestamp	time.Time
	Agent		int64
	Account		string
	Money       int64
	Type		int
	Status		int
	Process		int 						//0:未处理 1：已处理
	AcTime		time.Time
}

func init(){
	//orm.RegisterModel(new(Game))
	//orm.RegisterModel(new(Order))
	//orm.RegisterModel(new(User))
	orm.RegisterModel(new(Order_log))
}



