package sys

import "web-game-api/models/dto"

type CommStruct struct {
	Id            int64  `orm:"description(主键)"`
	AdminId       int64  `orm:"description(所有人)"`
	CreateAdminId int64  `orm:"description(创建人)"`
	CreateTime    uint64 `orm:"description(创建时间)"`
	UpdateTime    uint64 `orm:"description(修改时间)"`
	Deleted       int    `orm:"description(0:正常  1:删除)"`
}

// 公共签名，对struct中的数据进行签名
type CommSign struct {
	Sign string
}

type CommStructInterface interface {
	// 将req转为 struct 的抽象接口
	FromReq(adminId, createAdminId int64, channelId string, req dto.CommReqInterface)
	SetCreateTime()
	SetUpdateTime()
	SetId(id int64) // 更新时使用
	SetDel()
}

type LoginToken struct {
	Id         int64
	Exp        int64
	CreateTime int64 // 创建时间，用于校验
}
