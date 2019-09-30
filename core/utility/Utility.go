package utility

import (
	"strings"
	"tsEngine/tsMicro"
	"tsEngine/tsString"
)

func SortToStdSql(sort string) (string, string) {
	index := strings.Index(sort, "-")
	if index < 0 {
		return sort, "ASC"
	}
	sort = strings.Replace(sort, "-", "", -1)
	return sort, "DESC"
}

type CountInfo struct {
	Count int64
}

func SortToUSort(sort string) (isSign bool, uSort string) {
	index := strings.Index(sort, "-")
	if index != -1 {
		isSign = true
		uSort = tsString.Substr(sort, index+1, len(sort))
		return
	}
	return false, sort
}

func FormatSort(sort string, exStr string) string {
	is, uSort := SortToUSort(sort)
	sort = tsString.CoverCamelToSnake(uSort)
	if is {
		sort = "-" + sort
	}
	sort, order := SortToStdSql(sort)
	sortSql := exStr + sort + " " + order
	return sortSql
}

type AccountCond struct {
	AdminId    int64
	AdminName  string
	Name       string
	SettleType string
	Status     int
	Role       string
	AgentId    int64
	//...
}

func JointAdminCond(accType int, keyword string) (string, []interface{}) {
	sqlStr := ""
	args := []interface{}{}
	if accType == 2 {
		sqlStr += "OR "
		sqlStr += "A.username LIKE ? OR A.name LIKE ? OR C.settle_type LIKE ?"
		if keyword != "" {
			args = append(args, keyword, keyword, keyword)
		}
	}
	return sqlStr, args
}

func IpToInt(ip string) int64 {
	ipArr := strings.Split(ip, ".")
	if len(ipArr) < 4 {
		return 0
	}
	times := 3
	result := int64(0)
	for index, ipd := range ipArr {
		t := tsString.ToInt64(ipd)
		for i := 0; i < times; i++ {
			t *= 255
		}
		result += t
		times -= index + 1
	}
	return result
}

type ReqCreateOrder struct {
	tsMicro.MicroSign
	OrderId  string //客户端生成的订单号(最大32位)
	Platform string //充值平台 alipay/alipay-web/wechat/bank
	Amount   int64  //充值金额(分，最低100)
	Vip      bool   //是否使用vip通道
	Callback string //回调地址(带http)
}

type RespCreatOrder struct {
	tsMicro.MicroSign
	OrderId       string `json:"orderId"`
	PayUrl        string `json:"payUrl"`
	Status        int    `json:"status"` //0-已创建 1-已提交 2-成功 3-失败
	Amount        int64  `json:"amount"`
	OwnWebBrowser bool   `json:"ownWebBrowser"` // 是否可用自带浏览器
}

type ReqCallback struct {
	OrderId   string
	PayStatus int   // 支付状态 2:成功 3:失败 4:取消
	PayTime   int64 // 秒
}

type RespGame struct {
	Code int
	Msg  string
}

type ResQueryPaymentOrder struct {
	Code int
	Data QueryPayOrderData
}

type QueryPayOrderData struct {
	OrderId string `json:"orderId"`
	Status  int    `json:"status"` // 支付状态 2:成功 3:失败 4:取消
	PayUrl  string `json:"payUrl"` // 支付链接
	Amount  int64  `json:"amount"` // 金额
}

func IfInt64(b bool, t, f int64) int64 {
	if b {
		return t
	}
	return f
}

func IfString(b bool, t, f string) string {
	if b {
		return t
	}
	return f
}

type ReqTrans struct {
	tsMicro.MicroSign
	OrderId     string `json:"orderId"`
	Platform    string `json:"platform"`
	Amount      int64  `json:"amount"`
	ExpTime     int64  `json:"expTime"`
	Account     string `json:"account"`
	RealName    string `json:"realName"`
	IdCard      string `json:"idCard"`
	Telephone   string `json:"telephone"`
	BankName    string `json:"bankName"`
	CallbackUrl string `json:"callbackUrl"`
}

type RespTrans struct {
	Code int
	Data RespTransOrder
}

type RespTransOrder struct {
	OrderId   string `json:"orderId"`
	Status    int    `json:"status"`
	Timestamp uint64 `json:"timestamp"`
}

type RespWithdrawCreatOrder struct {
	Code int
	Data RespTransOrder
}
