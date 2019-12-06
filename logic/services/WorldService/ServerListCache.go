package WorldService

import (
	"github.com/astaxie/beego/logs"
	"gopkg.in/mgo.v2/bson"
	"strconv"
	"time"
	"tsEngine/tsRand"
	"web-game-api/core/mulMongo"
	"web-game-api/logic/gameEnum"
)

var Instance	*ServerListCache


type ServerListCache struct {
	ServerCache map[int64]ServerInfo
}

type ServerInfo struct {		//对应的是ServerList时的数据
	ServerId		int64 	`bson:"ServerId"`
	SeverType		string  `bson:"SeverType"`
	ServerIp		string 	`bson:"ServerIp"`
	Status			int64	`bson:"Status"`
	GateType		string	`bson:"GateType"`

	GateId			int64
	Host			string
	GateIp			string			//合成的字段
}

type ServerAccountInfo struct {		//对应的是ServerInfo里的数据
	Account 		string		`bson:"Account"`
	GateId			int64		`bson:"GateId"`
	IsConnect		int64		`bson:"IsConnect"`
	WorldId			int64		`bson:"WorldId"`
	LastTime 		time.Time	`bson:"LastTime"`
}

func GetInstance() *ServerListCache {
	if Instance == nil {
		Instance = &ServerListCache{}
		Instance.ServerCache = map[int64]ServerInfo{}
		go Instance.OnTimer()			//另开一个结程执行计时
	}
	return Instance
}

func (this* ServerListCache)Init(){
	this.ServerCache =  map[int64]ServerInfo{}		//清空数据
	serverInfoList := []ServerInfo{}
	err := mulMongo.FindAll(gameEnum.DB_NAME_CONFIG,gameEnum.TABLE_COLLECTION_SERVERLIST,bson.M{"ServerType":1,"GateType":"web"},bson.M{},&serverInfoList)
	if err != nil {
		logs.Error("ServerList Data error!")
		return
	}

	for _,v := range serverInfoList {
		v.GateId = v.ServerId
		if v.Host != "" {
			v.GateIp = v.Host
		} else {
			v.GateIp = v.ServerIp + ":" + strconv.FormatInt(v.GateId,10)
		}

		this.ServerCache[ v.GateId ] = v
	}
}

func (this* ServerListCache)OnTimer(){
	tick := time.NewTicker(60 * time.Second)
	defer tick.Stop()

	for {
		<- tick.C
		this.Init()
	}
}


func (this* ServerListCache)AllocServer(account string) (ServerInfo,bool){
	serverAccountInfo := ServerAccountInfo{}
	err := mulMongo.FindOne(gameEnum.DB_NAME_ACCOUNT,gameEnum.TABLE_COLLECTION_SERVER_INFO,bson.M{"Account":account},bson.M{},&serverAccountInfo)
	if err != nil && err.Error() != "not found" {
		return ServerInfo{},false
	}

	if data,ok := this.ServerCache[ serverAccountInfo.GateId ]; ok {
		if data.Status == 1 && data.GateType == "web" {
			return data,true
		}
	}

	//随机取一个数据
	gateIdList :=[ ]int64{}

	for k,_ := range this.ServerCache {
		gateIdList = append(gateIdList, k)
	}

	randIdx := tsRand.RandInt(0,len(gateIdList))
	randGateId := gateIdList[ randIdx ]

	serverAccountInfo.GateId = randGateId
	serverAccountInfo.Account = account
	serverAccountInfo.IsConnect = 0
	serverAccountInfo.LastTime = time.Now()
	mulMongo.Upsert(gameEnum.DB_NAME_ACCOUNT,gameEnum.TABLE_COLLECTION_SERVER_INFO,bson.M{"Account":account},&serverAccountInfo)

	return this.ServerCache[ randGateId ],true
}
