package consts

// 自定义用户类型
const (
	// 云平台默认db类型
	CloudDBEngin = "mysql"

	// 系统用户类型
	AdminTypeAgent       = 1 // 代理类型
	AdminTypePromoter    = 2 // 推广类型
	AdminTypeCoindearler = 3 // 币商类型
	AdminTypeSonuser     = 4 // 子账号类型
)


// redis key
const (
	KeyWEBAPISysAdminById   = "WEBAPI:SysAdminById:%v"
	KeyWEBAPISysAdminByName = "WEBAPI:SysAdminByName:%v"
	KeyWEBAPISysIpbanById   = "WEBAPI:SysIpban:%v"
	KeyWEBAPISysIpbanByIp   = "WEBAPI:SysIpbanByIp:%v"
	KeyWEBAPISysModeById    = "WEBAPI:SysMode:%v"
	KeyWEBAPISysModeByMD5   = "WEBAPI:SysModeByMD5:%v"
	KeyWEBAPISysNodeById    = "WEBAPI:SysNode:%v"
	KeyWEBAPISysRoleById    = "WEBAPI:SysRole:%v"
	KeyWEBAPIIsLogin        = "WEBAPI:IsLogin:%v" //是否已登录

	KeyWEBAPIAccountAmount = "WEBAPIAccountAmount:%d"
)


//web api redis key
const(
	KeyWEBAPIAgentById   = "WEBAPI:SysAdminById:%v"
)