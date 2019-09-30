// @APIVersion 1.0.0
// @Title webgame API
// @Description 用于连接web服务器与h5服务器
// @Contact daozhang@88888.com
package routers

import (
	"web-game-api/controllers"
	"web-game-api/controllers/admin"
	"github.com/astaxie/beego"
	"web-game-api/controllers/api"
	"web-game-api/controllers/system"
	"web-game-api/controllers/webApi"
)

func init() {
    beego.Router("/", &controllers.MainController{})
    //--------------------------与外部接口start----------------------------------------------
    apiNs := beego.NewNamespace("/apiurl",
		// 玩家登陆
    	beego.NSNamespace("/user", beego.NSInclude(&api.UserController{})),
    	beego.NSNamespace("/money", beego.NSInclude(&api.MoneyController{})),
    	beego.NSNamespace("/history", beego.NSInclude(&api.HistoryController{})),
    	beego.NSNamespace("/test", beego.NSInclude(&api.TestController{})),
    )
    beego.AddNamespace(apiNs)


	//--------------------------与外部接口end----------------------------------------------

	//--------------------------与后台内容start----------------------------------------------
	webNs := beego.NewNamespace("/webapi",
		beego.NSNamespace("/agent", beego.NSInclude(&webApi.AgentController{})),
		beego.NSNamespace("/order", beego.NSInclude(&webApi.OrderController{})),
		beego.NSNamespace("/player", beego.NSInclude(&webApi.PlayerController{})),
	)
	beego.AddNamespace(webNs)
	//--------------------------与后台内容end------------------------------------------------

	/*********************系统路由********************************/
	systemNs := beego.NewNamespace("/webapi/system",
		// 登录
		beego.NSNamespace("/user", beego.NSInclude(&system.LoginController{})),
		// 节点
		beego.NSNamespace("/node", beego.NSInclude(&system.NodeController{})),
		// 模块管理
		beego.NSNamespace("/mode", beego.NSInclude(&system.ModeController{})),
		// 管理员
		beego.NSNamespace("/admin", beego.NSInclude(&system.AdminController{})),
		// 角色
		beego.NSNamespace("/role", beego.NSInclude(&system.RoleController{})),
		// ip屏蔽
		beego.NSNamespace("/ipban", beego.NSInclude(&system.IpbanController{})),
		// 日志
		beego.NSNamespace("/logs", beego.NSInclude(&system.LogsController{})),
		// 公共
		beego.NSNamespace("/public", beego.NSInclude(&system.PublicController{})),
	)
	beego.AddNamespace(systemNs)
	//404 等错误处理
	beego.ErrorController(&system.ErrorController{})
	/*********************系统路由 end ***************************/

	adminNs := beego.NewNamespace("/webapi/admin",
		beego.NSNamespace("/sonuser", beego.NSInclude(&admin.SonUserController{})),
	)


	beego.AddNamespace(adminNs)

	//bee run -gendoc=true -downdoc=true
}
