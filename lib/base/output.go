package base

// 设置key value 形式输出数据 json，jsonp，tpl方式渲染适用
func (c *BaseController) SetOutMapData(key string, value interface{}) {
	if key == "" {
		return
	}
	c.OutMapData[key] = value
}
