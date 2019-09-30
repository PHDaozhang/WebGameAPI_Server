package system

type ErrorController struct {
	BaseController
}

func (this *ErrorController) Error404() {
	this.Error(404)
}

func (this *ErrorController) Error501() {
	this.Error(501)
}

func (this *ErrorController) ErrorDb() {
	this.Error(500)
}
