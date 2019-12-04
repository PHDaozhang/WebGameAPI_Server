package saApi

import (
	"encoding/xml"
	"github.com/astaxie/beego/httplib"
	"net/url"
	"strconv"
	"tsEngine/tsCrypto"
	"tsEngine/tsRand"
	"web-game-api/controllers/system"
)

const(
	api_url = "http://10.1.12.50:14004"
)

type SAPlatformUser struct {
	system.BaseController
}

type Books struct {
	XMLName xml.Name `xml:"books"`;
	Nums    int      `xml:"nums,attr"`;
	Book    []Book   `xml:"book"`;
}

type Book struct {
	XMLName xml.Name `xml:"book"`;
	Name    string   `xml:"name,attr"`;
	Author  string   `xml:"author"`;
	Time    string   `xml:"time"`;
}

// @Title sa平台登陆
// @Description 管理员列表
// @Success 200 {object} admin.Admin
// @router   /login [post]
func (this *SAPlatformUser) Login() {
	response := LoginRequestResponse{}
	response.ErrorMsgId = 0
	response.Token = "token......."
	response.DisplayName = "love";
	response.GameURL = "http://www.baidu.com"

	this.Data["xml"] = response
	this.ServeXML()
	this.StopRun()
}



// @Title sa平台登陆
// @Description 管理员列表
// @Success 200 {object} admin.Admin
// @router   /GetUserBalance [post]
func (this *SAPlatformUser)GetUserBalance(){
	userName := this.GetString("username","userName")
	currency := this.GetString("currency","CNY")

	requestUrl := api_url + "/api/sa/GetUserBalance"
	req := httplib.Post(requestUrl)
	req.Header("Content-Type", "application/x-www-form-urlencoded")

	paramsMap := map[string]interface{}{}
	paramsMap["username"] = userName
	paramsMap["currency"] = currency

	postStr := BuildParamAndDes(paramsMap)
	req.Body(postStr)

	responseStr,err := req.String()
	if err != nil {
		this.Error(100,"向c++服务器请求错误")
	}

	this.Ctx.Output.Body([]byte(responseStr))
}


// @Title sa平台登陆
// @Description 管理员列表
// @Success 200 {object} admin.Admin
// @router   /PlaceBet [get]
func (this *SAPlatformUser)PlaceBet(){
	userName := this.GetString("username","userName")
	currency := this.GetString("currency","CNY")
	amount,_ := this.GetInt("amount",100)
	txnid,_ := this.GetInt("txnid",-1)
	if txnid == -1 {
		txnid =  tsRand.RandInt(100000,200000)
	}
	gameid,_ := this.GetInt("gameid",-1)
	if gameid == -1 {
		gameid = tsRand.RandInt(1,10000)
	}

	requestUrl := api_url + "/api/sa/PlaceBet"
	req := httplib.Post(requestUrl)
	req.Header("Content-Type", "application/x-www-form-urlencoded")

	paramsMap := map[string]interface{}{}
	paramsMap["username"] = userName
	paramsMap["currency"] = currency
	paramsMap["amount"] = strconv.Itoa(amount)
	paramsMap["txnid"] = strconv.Itoa(txnid)
	paramsMap["gametype"] = "rot"
	paramsMap["platform"] = 1			//0:桌面版
	paramsMap["hostid"] = 10
	paramsMap["gameid"] = gameid

	postStr := BuildParamAndDes(paramsMap)
	req.Body(postStr)

	responseStr,err := req.String()
	if err != nil {
		this.Error(100,"向c++服务器请求错误")
	}

	this.Ctx.Output.Body([]byte(responseStr))
}


// @Title sa平台登陆
// @Description 管理员列表
// @Success 200 {object} admin.Admin
// @router   /PlayerWin [get]
func (this *SAPlatformUser)PlayerWin(){
	userName := this.GetString("username","userName")
	currency := this.GetString("currency","CNY")
	amount,_ := this.GetInt("amount",100)
	txnid,_ := this.GetInt("txnid",-1)
	if txnid == -1 {
		txnid =  tsRand.RandInt(100000,200000)
	}
	gameid,_ := this.GetInt("gameid",-1)
	if gameid == -1 {
		gameid = tsRand.RandInt(1,10000)
	}

	requestUrl := api_url + "/api/sa/PlayerWin"
	req := httplib.Post(requestUrl)
	req.Header("Content-Type", "application/x-www-form-urlencoded")


	paramsMap := map[string]interface{}{}
	paramsMap["username"] = userName
	paramsMap["currency"] = currency
	paramsMap["amount"] = strconv.Itoa(amount)
	paramsMap["txnid"] = strconv.Itoa(txnid)
	paramsMap["gametype"] = "rot"
	paramsMap["hostid"] = 10
	paramsMap["gameid"] = gameid
	paramsMap["Payouttime"] = "2019-10-30 12:34:56"			//0:桌面版


	postStr := BuildParamAndDes(paramsMap)
	req.Body(postStr)

	responseStr,err := req.String()
	if err != nil {
		this.Error(100,"向c++服务器请求错误")
	}

	this.Ctx.Output.Body([]byte(responseStr))
}



// @Title sa平台登陆
// @Description 管理员列表
// @Success 200 {object} admin.Admin
// @router   /PlayerLost [get]
func (this *SAPlatformUser)PlayerLost(){
	userName := this.GetString("username","userName")
	currency := this.GetString("currency","CNY")
	txnid,_ := this.GetInt("txnid",-1)
	if txnid == -1 {
		txnid =  tsRand.RandInt(100000,200000)
	}
	gameid,_ := this.GetInt("gameid",-1)
	if gameid == -1 {
		gameid = tsRand.RandInt(1,10000)
	}

	requestUrl := api_url + "/api/sa/PlayerLost"
	req := httplib.Post(requestUrl)
	req.Header("Content-Type", "application/x-www-form-urlencoded")


	paramsMap := map[string]interface{}{}
	paramsMap["username"] = userName
	paramsMap["currency"] = currency
	paramsMap["txnid"] = strconv.Itoa(txnid)
	paramsMap["gametype"] = "rot"
	paramsMap["Payouttime"] =  "2019-10-30 12:34:56"
	paramsMap["hostid"] = 10
	paramsMap["gameid"] = gameid

	postStr := BuildParamAndDes(paramsMap)
	req.Body(postStr)

	responseStr,err := req.String()
	if err != nil {
		this.Error(100,"向c++服务器请求错误")
	}

	this.Ctx.Output.Body([]byte(responseStr))
}


// @Title sa平台登陆
// @Description 管理员列表
// @Success 200 {object} admin.Admin
// @router   /PlaceBetCancel [get]
func (this *SAPlatformUser)PlaceBetCancel(){
	userName := this.GetString("username","userName")
	currency := this.GetString("currency","CNY")
	amount,_ := this.GetInt("amount",100)
	txnid,_ := this.GetInt("txnid",-1)
	if txnid == -1 {
		txnid =  tsRand.RandInt(100000,200000)
	}
	gameid,_ := this.GetInt("gameid",-1)
	if gameid == -1 {
		gameid = tsRand.RandInt(1,10000)
	}
	txn_reverse_id,_ := this.GetInt("txn_reverse_id",-1)
	if txn_reverse_id == -1 {
		this.Error(-1,"需要txn_reverse_id参数")
	}

	requestUrl := api_url + "/api/sa/PlaceBetCancel"
	req := httplib.Post(requestUrl)
	req.Header("Content-Type", "application/x-www-form-urlencoded")


	paramsMap := map[string]interface{}{}
	paramsMap["username"] = userName
	paramsMap["currency"] = currency
	paramsMap["amount"] = strconv.Itoa(amount)
	paramsMap["txnid"] = strconv.Itoa(txnid)
	paramsMap["gametype"] = "rot"
	paramsMap["hostid"] = 10
	paramsMap["gameid"] = gameid
	//paramsMap["platform"] = 1			//0:桌面版
	paramsMap["txn_reverse_id"] = txn_reverse_id

	postStr := BuildParamAndDes(paramsMap)
	req.Body(postStr)

	responseStr,err := req.String()
	if err != nil {
		this.Error(100,"向c++服务器请求错误")
	}

	this.Ctx.Output.Body([]byte(responseStr))
}

// @Title sa平台登陆
// @Description 管理员列表
// @Success 200 {object} admin.Admin
// @router   /TestSaApiLogin [get]
func  (this *SAPlatformUser)TestSaApiLogin(){

	/**
	q := "6vHmhhReDPs7OY30aqM23mGuz%2fx4oaKMVzb7x7oTQX0Ht9B1Q7BviQx8trt7oRK5oaABMNqO%0ab1RWCjkvg5oTeKjldAAfcIONtxfKGBJe5DuGQ1m88c%2f3TtBdfu58%2fp5DOfphjc6rd7j7pYai%0aZt7hKg%3d%3d%0a"
	s := "e279828934732bf11ea0549170536882"

	requestUrl :=  "http://sai-api.sa-apisvr.com/api/api.aspx"
	req := httplib.Post(requestUrl)
	req.Header("Content-Type", "application/x-www-form-urlencoded")
	req.Param("q",q)
	req.Param("s",s)


	responseStr,err := req.String()
	if err != nil {
		this.Error(100,"向c++服务器请求错误")
	}

	this.Ctx.Output.Body([]byte(responseStr))
	**/

	postStr := "MDQGx7yYdma2DRSxY3abZ64yK7TjN20S0wxxLhWHk8E%3d";
	newPostStr,err := url.Parse(postStr)
	//postStr = tsCrypto.
	//postStr2 := tsCrypto.Base64Decode(newPostStr.Path);

	//fmt.Println(postStr2)

	des := tsCrypto.Des{}
	des.Strkey = DESKEY
	des.EncodeType = tsCrypto.EncodeBase64
	des.PadType = tsCrypto.PadString
	des.Iv = DESKEY

	data,err :=  des.DecryptCBC(newPostStr.Path)

	if err != nil {
		this.Error(100);
	}

	afterDes := string(data)

	newStr := tsCrypto.Base64Decode(afterDes);

	this.Success(newStr);
}