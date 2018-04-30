package base

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/l-dandelion/yi-ants-go/lib/common"
)

// 依赖appconfig里的提供的几个配置 switch.enablehttps
type BaseController struct {
	beego.Controller
	InputData *common.InputData
	// 输出数据，可以在子类中直接设定；也可以调用SetOutMapData设定
	OutMapData map[string]interface{}
	// 处理过程中出现错误时，错误消息内容保存到这个变量(展示给用户看的)
	ErrMsg string
	cName string
	aName string
	IsApi bool
}

// Init generates default values of controller operations.
func (c *BaseController) Init(ctx *context.Context, controllerName, actionName string, app interface{}) {
	c.Controller.Init(ctx, controllerName, actionName, app)
	c.TplExt = "tpl"
	c.InputData = c.initInputData()
	c.OutMapData = make(map[string]interface{})
	c.ErrMsg = ""
	c.TplName = c.cName + "/" + c.aName + "." + c.TplExt
}

// Finish runs after request function execution.
func (c *BaseController) Finish() {
	if !c.Ctx.ResponseWriter.Started {
		c.reply()
	}

	if c.InputData != nil {
		inputDataPool.Put(c.InputData)
	}
}
