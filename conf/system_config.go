package conf


const (
	SystemAdminId = int64(-1)

	// 登录加密字串
	PasswordSalt = "iloveyourmorethanicansay"
	AesKey       = 1580081714143943634

	TokenSalt         = "iloveyourmorethanicansaytoo"
	TokenExpMinute    = 10
	TokenMaxExpSecond = 864000 // 60*60*24*10  免登陆最长时间 10天

	// ###################### 数据库配置 ############################
	DBHost     = "127.0.0.1"
	DBPort     = "3306"
	DBUser     = "root"
	DBPassword = "123456"
	DBName     = "api"
	// ###################### Redis配置 ############################

	SpecialPermissions = "127,128,129,130,131"
)