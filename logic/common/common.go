package common

//请求消息体
type ReturnMsgStruct struct {
	Code			int;
	Msg				string;
	Data			map[string]interface{};
}