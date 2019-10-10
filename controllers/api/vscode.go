package api

import (
	"web-game-api/controllers/system"
)

//abc
type Abc struct {
	system.BaseController
}

func (this *Abc) test() {
	this.Code = 1

	this.Success("go this...1")
}
