package base

import (
	"github.com/l-dandelion/yi-ants-go/lib/common"
	"github.com/l-dandelion/yi-ants-go/lib/constant"
	"github.com/l-dandelion/yi-ants-go/lib/utils"
)

func (c *BaseController) initInputData() (data *common.InputData) {
	data = inputDataPool.Get().(*common.InputData)
	data.OutputType = constant.OutputTypeHtml

	data.URI = utils.GetNoDomainUri(c.Ctx.Input.URI())
	c.IsApi, c.cName, c.aName = utils.GetControllerInfoFromUri(data.URI)
	if c.IsApi {
		data.OutputType = constant.OutputTypeJson
	}
	return data
}