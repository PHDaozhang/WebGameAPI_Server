package manager

import (
	"github.com/astaxie/beego/logs"
	"io/ioutil"
	"os"
	"time"
	"tsEngine/tsCrypto"
)

var instance *UrlConfigMgr
var urlConfigDatas map[string]UrlConfig

type UrlConfigMgr struct {

}

type UrlConfig struct {
	Config 			string
	Key  			string
	encryptMsg		string
}

func GetInstance() *UrlConfigMgr {
	if instance == nil {
		instance = &UrlConfigMgr{}
		urlConfigDatas = map[string]UrlConfig{}
		go instance.OnTimer()
	}

	return instance
}

func (this* UrlConfigMgr)OnTimer(){
	tick := time.NewTicker(60 * time.Second)
	defer tick.Stop()

	for {
		<- tick.C
		logs.Info("....onTimer")
		this.Init()
	}
}

//每隔一分钟获取一次配置表
func  (this* UrlConfigMgr)Init() {
	files,_ := ioutil.ReadDir("./conf/serverConf")

	for _,f := range files {
		if f.IsDir() {
			continue
		}

		file,err := os.Open("./conf/serverConf" + "/" + f.Name())
		if err != nil {
			continue
		}
		contentTxt,err := ioutil.ReadAll(file)

		logs.Info(string(contentTxt))

		cfg := UrlConfig{}
		cfg.Config = string(contentTxt)
		cfg.Key = tsCrypto.GetMd5( contentTxt )
		cfg.encryptMsg = tsCrypto.Base64EncodeByte(contentTxt)

		data,ok := urlConfigDatas[ f.Name() ]
		if ok == false {
			urlConfigDatas[  f.Name() ] = cfg
		} else if data.Key != cfg.Key {
			urlConfigDatas [ f.Name() ] = cfg
		}
	}

}


func (this* UrlConfigMgr)GetJsonUrlDiffer(key string) string {
	fileName := "UrlConfig.json"

	if data,ok := urlConfigDatas[ fileName ];ok && data.Key != key {
		return data.encryptMsg
	}

	return ""
}