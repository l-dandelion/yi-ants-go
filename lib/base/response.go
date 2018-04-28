package base

import (
	"bytes"
	"encoding/json"
	"github.com/l-dandelion/crawler_lib/constant"
	"html/template"
	"net/http"
	"strings"
)

type RespData struct {
	ErrMsg string
	Data   interface{}
}

func (c *BaseController) reply() {
	//标识已经有输出了，躲过执行Process
	//c.Ctx.ResponseWriter.Started = true
	//动态站所有请求，要求浏览器端不做缓存
	c.Ctx.Output.Header("Cache-Control", "no-store, no-cache, must-revalidate, post-check=0, pre-check=0")
	c.Ctx.Output.Header("Expires", "Mon, 26 Jul 1997 05:00:00 GMT")

	switch c.InputData.OutputType {
	case constant.OutputTypeJson: //json,jsonp输出
		c.replyAjax()
	case constant.OutputTypeHtml: //模板输出
		c.replyHtml()
	}
}

func (c *BaseController) replyAjax() {
	var data interface{}
	data = c.OutMapData

	rt := &RespData{
		ErrMsg: c.ErrMsg,
		Data:   data,
	}

	callback := strings.TrimSpace(c.Ctx.Input.Query("jsonpcallback"))
	if callback == "" {
		c.replyJson(rt)
	} else {
		c.replyJsonp(callback, rt)
	}
}
func (c *BaseController) replyJson(rt interface{}) {

	content, err := json.Marshal(rt)
	if err != nil { //rt中的data有可能出现转换失败的情况；譬如: json: unsupported type: map[interface {}]interface {}
		c.SetErrMsg(err.Error())
		rt := &RespData{
			ErrMsg: c.ErrMsg,
		}
		c.replyJson(rt)
		//http.Error(c.Ctx.ResponseWriter, err.Error(), http.StatusInternalServerError)
		return
	}

	c.Ctx.Output.Header("Content-Type", "application/json; charset=utf-8")
	c.Ctx.Output.Body(content)
}

func (c *BaseController) replyHtml() {
	if c.ErrMsg != "" {
		c.replyErrorPage()
		return
	}
	c.replyRender()
	return
}

func (c *BaseController) replyErrorPage() {
	c.TplName = "error/index" + "." + c.TplExt
	c.Data["ErrMsg"] = c.ErrMsg
	//渲染
	err := c.Render()
	if err != nil {
		http.Error(c.Ctx.ResponseWriter, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *BaseController) replyRender() {
	for key, value := range c.OutMapData {
		c.Data[key] = value
	}

	//渲染
	err := c.Render()
	if err != nil {
		http.Error(c.Ctx.ResponseWriter, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *BaseController) replyJsonp(callback string, rt interface{}) {
	c.Ctx.Output.Header("Content-Type", "application/javascript; charset=utf-8")
	content, err := json.Marshal(rt)

	if err != nil { //rt中的data有可能出现转换失败的情况；譬如: json: unsupported type: map[interface {}]interface {}
		c.SetErrMsg(err.Error())
		rt := &RespData{
			ErrMsg: c.ErrMsg,
		}
		c.replyJsonp(callback, rt)
		return
	}

	callbackContent := bytes.NewBufferString(" " + template.JSEscapeString(callback))
	callbackContent.WriteString("(")
	callbackContent.Write(content)
	callbackContent.WriteString(");\r\n")
	c.Ctx.Output.Body(callbackContent.Bytes())
}
