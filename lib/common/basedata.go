package common

// 依赖appconfig里的提供的几个配置 switch.enablehttps
type InputData struct {
	OutputType  int8
	IsNeedLogin bool
	CheckXsrf   bool
}

type ContextInterface interface {
	SetErrMsg(string)
	SetError(error)
	SetOutMapData(string, interface{})
}
