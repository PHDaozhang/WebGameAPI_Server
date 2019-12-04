package saApi

import (
	"github.com/astaxie/beego/logs"
	"net/url"
	"strconv"
	"strings"
	"tsEngine/tsCrypto"
	"tsEngine/tsString"
)

const(
	SecretKey="9E42915BBCC345EC918AED667D12C3E2"
	MDKEY = "GgaIMaiNNtg"
	DESKEY = "g9G16nTs"
	LV	=	"g9G16nTs"
)


//加密QS
func BuildQS(methodName string,time uint64,paramsMap map[string]interface{}) string {
	 paramsStr := ""

	 for k,v := range paramsMap{
	 	switch v.(type) {
		case string:
			paramsStr += "&" + k + "=" + v.(string)
			break
		case int:
			paramsStr += "&" + k + "=" + strconv.Itoa( v.(int) )
			break
		default:
			paramsStr += "&" + k + "=" + v.(string)
			break
		}
	 }

	 fmtStr := "method=" +  methodName + "&Key=" + SecretKey + "&Time=" + strconv.FormatUint(time,10)  + paramsStr

	 des := tsCrypto.Des{}
	 des.Iv = DESKEY
	 des.Strkey = DESKEY
	 des.PadType = tsCrypto.PadString
	 des.PadType = tsCrypto.EncodeBase64

	returnRes,err := des.EncryptCBC([]byte(fmtStr))
	if err != nil {
		logs.Error("加密错误:" + fmtStr)
		return ""
	}

	 return returnRes
}


func UnBuildQS(qs string)(paramsMap map[string]string) {
	logs.Info("q:" + qs)

	des := tsCrypto.Des{}
	des.Strkey = DESKEY
	des.EncodeType = tsCrypto.EncodeBase64
	des.PadType = tsCrypto.PadString
	des.Iv = DESKEY

	cryptoQS,err := des.DecryptCBC(qs)

	if err != nil {
		logs.Error("解密失败")
	}

	list := strings.Split(string(cryptoQS),"&")

	for i := 0; i < len(list);i++ {
		itemStr := list[i]

		itemList := strings.Split(itemStr,"=")

		if len(itemList) != 2 {
			logs.Info("the itemStr is not has the len 2" + itemStr)
			continue
		}

		if paramsMap == nil {
			paramsMap = map[string]string{}
		}

		paramsMap[ itemList[0] ] = itemList[1]
	}

	return
}


func BuildMD5( qs string,ymdhms string ) string {
	str := qs + MDKEY + ymdhms + DESKEY
	logs.Info("calc md5:" + str)
	md5 := tsCrypto.GetMd5([]byte(str))
	//md5 = strings.ToUpper(md5)
	logs.Info("after md5:" + md5)
	return md5
}

func ParseRequestBody(request string)map[string]string{
	values,err := url.ParseQuery(request)
	if err != nil {
		return nil
	}

	returnMap := map[string]string{}

	for k,v := range values {
		if len(v[0]) == 0 {
			if idx1 := strings.Index(k,"&"); idx1 != -1 {
				list := strings.Split(k,"&")
				for i := 0; i < len(list);i++ {
					itemStr := list[i]
					idx := strings.Index(itemStr,"=")
					newK := tsString.Substr(itemStr,0,idx)
					newV := tsString.Substr(itemStr,idx+1, len(k))
					returnMap[newK] = newV
				}
			} else {
				idx := strings.Index(k,"=")
				newK := tsString.Substr(k,0,idx)
				newV := tsString.Substr(k,idx+1, len(k))
				returnMap[newK] = newV
			}
		} else {
			returnMap[k] = v[0]
		}

	}

	if len(returnMap) > 0 {
		return returnMap
	}


		list := strings.Split(request,"&")
	if len(list) == 0 {
		return nil
	}



	for i:= 0; i < len(list);i++ {
		innerList := strings.Split(list[i],"=")
		if len(innerList) < 2 {
			continue;
		}

		returnMap[ innerList[0] ] = innerList[1]
	}

	return returnMap
}

func paraseURLDecode(encodeStr string) string {
	return ""
}


func BuildParamAndDes(paramMap map[string]interface{})string{
	paramStr := ""

	for k,v := range paramMap{
		switch v.(type) {
		case string:
			paramStr += "&" + k + "=" + v.(string)
			break
		case int:
			paramStr += "&" + k + "=" + strconv.Itoa( v.(int) )
			break
		}
	}

	paramStr = tsString.Substr(paramStr,1, len(paramStr)-1)

	des := tsCrypto.Des{}
	des.Iv = DESKEY
	des.Strkey = DESKEY
	des.PadType = tsCrypto.PadString
	des.PadType = tsCrypto.EncodeBase64

	desStr,err := des.EncryptCBC([]byte(paramStr))

	if err != nil {
		return ""
	}

	desStr = url.QueryEscape(desStr)

	return desStr
}