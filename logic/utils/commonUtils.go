package utils

import (
	"regexp"
	"strconv"
	"strings"
	"time"
	"tsEngine/tsCrypto"
	"tsEngine/tsString"
)

//检查订单流水号是否合法
func CheckOrderId( orderId string,agent int64,account string ) bool{
	agentStr := strconv.FormatInt(agent,10)
	if !strings.HasPrefix(orderId,agentStr){
		return false
	}
	if !strings.HasSuffix(orderId,account){
		return false
	}

	timeStrLen := len(orderId) - len(agentStr) - len(account)
	newOrder := tsString.Substr(orderId,len(agentStr),timeStrLen)
	//timestampUnix := tsString.ToInt(newOrder)
	//timestampSec := timestampUnix / 1000
	//timestatmpMSec := (timestampUnix % 1000) * 1000 * 1000
	//
	//curTime := time.Unix( int64(timestampSec) ,int64(timestatmpMSec))


	//MustCompile解析并返回一个正则表达式。如果成功返回，该Regexp就可用于匹配文本。	yyyyMMddHHmmssSSS
	dateReg := regexp.MustCompile(`20\d{15}`)			//基本的验证下是不是时间戳字符
	matchFlag := dateReg.MatchString(newOrder)

	return matchFlag
}

/**
生成流水号
 agent + yyyymmddHHmmssSSS
 */
func GenOrderId(agent int,account string)(orderId string) {
	curTime := time.Now()
	year := curTime.Year()
	month := curTime.Month()
	day := curTime.Day()
	hour := curTime.Hour()
	minute := curTime.Minute()
	second := curTime.Second()
	mincSecond := curTime.UnixNano() % 1000000

	dateStr := strconv.Itoa(year) + getTwoNumStr( int(month) ) + getTwoNumStr(day) + getTwoNumStr(hour) + getTwoNumStr(minute) + getTwoNumStr(second) + getTwoNumStr(int(mincSecond) )
	orderId =  strconv.Itoa(agent)  + dateStr + account
	return
}

/**
当数字小于两位的时候前面补位0
 */
func getTwoNumStr(num int)string{
	if num < 10 {
		return "0" + strconv.Itoa(num)
	} else {
		return strconv.Itoa(num)
	}

}

/**
检查签名是否正确
 */
func CheckSign(agent int64,timestamp int64,appkey string,sign string) bool {
	preMd5Str := strconv.FormatInt(agent,10)  + strconv.FormatInt(timestamp,10)  + appkey
	myGenMD5 := tsCrypto.GetMd5([]byte(preMd5Str))

	return myGenMD5 == sign
}

/**
根据代理 + 时间戳 + appkey获得玩家的签名
 */
func GenSign(agent int,timestamp int64,appkey string) (md5 string){
	preMd5Str := strconv.Itoa(agent)  + strconv.FormatInt(timestamp,10)  + appkey
	md5 = tsCrypto.GetMd5([]byte(preMd5Str))
	return
}

/**
通过账号与代理号获得玩家在游戏中的account
 */
func GameAccount(agent int64,account string) string {
	return "api_" + strconv.FormatInt(agent,10)  + "_" + account
}