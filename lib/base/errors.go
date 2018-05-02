package base

import "github.com/l-dandelion/yi-ants-go/lib/constant"

// 设定当前系统定义的错误号，一般设置完后，需要直接从process中return
func (c *BaseController) SetErrMsg(errMsg string) {
	c.ErrMsg = errMsg
}

func (c *BaseController) SetError(err error) {
	c.ErrMsg = err.Error()
}

func (c *BaseController) SetErrorf(errno int, errdesc string, args ...interface{}) {
	yierr := constant.NewYiErrorf(errno, errdesc, args...)
	c.SetError(yierr)
}