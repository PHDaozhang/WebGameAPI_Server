package api

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type AccountDBInfo struct {
	Id_					bson.ObjectId `bson:"_id"`
	RandKey				int32 `bson:"RandKey"`
	LastTime			int64 `bson:"LastTime"`
	LastIp				string `bson:"LastIp"`
	Account				string `bson:"Account"`
	AccPhone			string `bson:"AccPhone"`
	AccPwd				string `bson:"AccPwd"`
	AccDev				string `bson:"AccDev"`
	DeviceId			string `bson:"DeviceId"`
	DevicePlatform		string `bson:"DevicePlatform"`
	Platform			string `bson:"Platform"`
	ChannelId			string `bson:"ChannelId"`						//渠道号
	ClientChannelId		string `bson:"ClientChannelId"`					//包渠道号
	RegisterTime		time.Time `bson:"RegisterTime"`					//此处需要插入date数据
	RegisterIp			string `bson:"RegisterIp"`
	Level				int64 `bson:"Level"`
	LocationParam		string `bson:"LocationParam"`
	PhoneNum			string `bson:"PhoneNum"`
}

type OrderInfo struct {
	Id          bson.ObjectId `bson:"_id"`
	OrderID     string        `bson:"OrderID"`
	Account     string        `bson:"Account"`
	PlayerID    string        `bson:"PlayerID"`
	PlayerName  string        `bson:"PlayerName"`
	ChannelID   string        `bson:"ChannelID"`
	BindID      string        `bson:"BindID"`
	PayType     int           `bson:"PayType"`
	RMB         int64           `bson:"RMB"`								//支付金额，单位 （元）
	RealRMB     int64           `bson:"RealRMB"`							//支付金额，单位 （元）
	PayCode     int64           `bson:"PayCode"`							//到账金币	RMB * 100
	Custom      int           `bson:"Custom"`								//
	PayPlatform string        `bson:"PayPlatform"`
	State				int `bson:"State"`								//支付状态  默认0：  1：  2：支付成功		[3：下分失败] 未与服务器配对
	CreateTime			time.Time `bson:"CreateTime"`
	UpdateTime			time.Time `bson:"UpdateTime"`
	PayTime				time.Time `bson:"PayTime"`
	Process				bool `bson:"Process"`
	Result				map[string]interface{} `bson:"Result"`
	Type                int `bson:"Type"`								//订单类型			1：乙 --> 甲  2：甲 -->乙
}

type ServerListInfo struct {
	Id          bson.ObjectId `bson:"_id"`
	ServerId	int64  `bson:"ServerId"`
	ServerType	int64  `bson:"ServerType"`
	Status		int64  `bson:"Status"`
	ServerIp	string  `bson:"ServerIp"`
	GateType	string  `bson:"GateType"`
	Host		string  `bson:"Host"`
}