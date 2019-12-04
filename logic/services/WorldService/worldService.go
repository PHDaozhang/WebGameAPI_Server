package WorldService

import (
	"gopkg.in/mgo.v2/bson"
	"strconv"
	"web-game-api/core/mulMongo"
	"web-game-api/logic/gameEnum"
)

//模仿ts 代码，根据玩家账号获取一个可用的服务器地址
func GetWorldHttpPath(account string)(path string){
	serverInfo := struct {
		WorldId 			int64  `bson:"WorldId"`
		IsConnect			int64  `bson:"IsConnect"`
		Account				string `bson:"Account"`
	}{}

	query := bson.M{}
	query["Account"] = account;
	query["WorldId"] = bson.M{"$gt":0}  //此打开，会导致下面的err 出现 not found

	err := mulMongo.FindOne(gameEnum.DB_NAME_ACCOUNT,gameEnum.TABLE_COLLECTION_SERVER_INFO,&query,bson.M{},&serverInfo)
	if err != nil && err.Error() != "not found" {
		return
	}

	if serverInfo.WorldId != 0 {
		serverListResult := struct {
			ServerIp 		string  `bson:"ServerIp"`
			ServerId		int64  `bson:"ServerId"`
		}{}
		serverListQuery := bson.M{}
		serverListQuery["ServerId"] = serverInfo.WorldId
		serverListQuery["Status"] = 1

		err = mulMongo.FindOne(gameEnum.DB_NAME_CONFIG,gameEnum.TABLE_COLLECTION_SERVERLIST,serverListQuery,bson.M{},&serverListResult)

		if err != nil  && err.Error() != "not found"  {
			return
		}
		port := serverListResult.ServerId + 1
		path = "http://" + serverListResult.ServerIp + ":" + strconv.FormatInt(port,10)
		return
	}

	//如果没有服务器可用，则使用第一个即可
	serverListResult := struct {
		ServerIp 		string  `bson:"ServerIp"`
		ServerId		int64  `bson:"ServerId"`
		ServerType		string  `bson:"ServerType"`
		Status			int  `bson:"Status"`
		Host			string  `bson:"Host"`
	}{}
	err = mulMongo.FindOne(gameEnum.DB_NAME_CONFIG,gameEnum.TABLE_COLLECTION_SERVERLIST,bson.M{"ServerType":3,"Status":1},bson.M{},&serverListResult)
	if err != nil  && err.Error() != "not found"  {
		return
	}

	port := serverListResult.ServerId + 1
	path = "http://" + serverListResult.ServerIp + ":" + strconv.FormatInt(port,10)
	return
}

