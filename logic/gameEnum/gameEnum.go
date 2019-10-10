package gameEnum

const(
	TRANS_OTHER_TO_US				=	1					//乙方到甲方
	TRANS_US_TO_OTHERS				=	2					//甲方到乙方
)

const(
	DB_NAME_ACCOUNT =	"AccountDB" //账号数据库名字
	DB_NAME_PAYMENT =	"PaymentDB" //订单数据库
	DB_NAME_CONFIG	=	"ConfigDB"	//用于获取登陆游戏的时候ｉｐ
	DB_NAME_PLAYER	=	"PlayerDB_DWC"			//用于玩家实际上的玩家信息

	ACCOUNT_COLLECTION_TABLE        =	"AccountTable" //账号集合
	TABLE_PAY_COLLECTION_ORDER		=	"Order"			//上下分集合，两张表合一张表
	TABLE_COLLECTION_SERVERLIST	    = "ServerList"
	TABKE_COLLECTION_PLAYERINFO		=	"PlayerInfo"		//游戏玩家的信息
	TABLE_COLLECTION_SERVER_INFO	=	"ServerInfo"		//当前玩家在某个服务器上面列表
)

const(
	REDIS_ORDER							=	"API:ORDER:STR:"
	REDIS_FILTER_TIME_GAP				=	"API:FILTER_TIME_GAP:HM"
)

const(
	ORDER_REDIS_STATUS_UN_HANDLER	= iota				//未处理
	ORDER_REDIS_STATUS_HANDING							//正在处理
	ORDER_REDIS_STATUS_HANDED							//已处理
	ORDER_REDIS_STATUS_HANDLER_ERR						//处理失败
)