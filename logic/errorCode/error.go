package errorCode

//错误编码
const(
	SUCC							=	1
	PARAM_ERROR						=	10001			//参数错误
	INNER_ERROR						=	10002			//内部错误
	UNKNOW_ERROR					=	10003			//未知错误
	DB_OPER_ERROR					=	10004			//操作数据库错误
	SIGN_ERROR						=	10005			//签名不正确
	ORDER_REG_ERROR					=   10006			//订单号编码有误
	ORDER_REPEAT_ERROR				=	10007			//订单重复
	ORDER_TIME_EXPIRE				=	10008			//订单时间过期
	ORDER_UN_EXIST					=	10009			//订单不存在
	AGENT_UNEXIST					=	10010			//代理不存在
	ACCOUNT_UN_EXIST				=	10011			//此账号不存在
	ACCOUNT_NAME_ILLIGLE			=	10012			//账号名字含有非法字符
	ACCOUNT_ORDER_UN_OPER			= 	10013			//有未处理的上下分订单
	COOL_DOWN						=	10014			//接口冷却时间
	DB_NOT_EXIST					=	10015			//数据不存在
	ADMIN_NOT_PERMISSION			=	10016			//权限不够
	GATE_SERVER_ERROR				=	10017			//找不到Gate服务器地址

	ORDER_PROCESS_DONE				=	20000			//订单已处理
)



const(

)