package data

import (
	"net/http"
	"encoding/gob"
	"errors"
)

/*
 * Request
 * httpReq: the http request
 * depth: crawl depth
 * proxy: use proxy if not empty
 * extra: additional information(used for context)
 */
type Request struct {
	RNodeName   string
	RSpiderName string
	RHttpReq    *http.Request          // the http request
	RDepth      uint32                 // crawl depth
	RProxy      string                 // use proxy if not empty
	Extra      map[string]interface{} // additional information(used for context)
}

/*
 * get spider name
 */
func (req *Request) SpiderName() string {
	return req.RSpiderName
}

/*
 * get node name
 */
func (req *Request) NodeName() string {
	return req.RNodeName
}

/*
 * set node name
 */
func (req *Request) SetNodeName(nodeName string) {
	req.RNodeName = nodeName
}

/*
 * get http request
 */
func (req *Request) HTTPReq() *http.Request {
	return req.RHttpReq
}

/*
 * get crawl depth
 */
func (req *Request) Depth() uint32 {
	return req.RDepth
}

/*
 * get extra infomation
 */
func (req *Request) SetExtra(key string, val interface{}) {
	req.Extra[key] = val
}

/*
 * set crawl depth
 */
func (req *Request) SetDepth(depth uint32) {
	req.RDepth = depth
}

/*
 * check the request
 */
func (req *Request) Valid() bool {
	return req.RHttpReq != nil && req.RHttpReq.URL != nil
}

/*
 * New an instance of Request
 */
func NewRequest(httpReq *http.Request, extras ...map[string]interface{}) *Request {
	var extra map[string]interface{}
	if len(extras) != 0 {
		extra = extras[0]
	} else {
		extra = map[string]interface{}{}
	}
	return &Request{
		RHttpReq: httpReq,
		Extra:   extra,
	}
}

/*
 * add cookie
 */
func (req *Request) AddCookie(key, value string) {
	c := &http.Cookie{
		Name:  key,
		Value: value,
	}
	req.RHttpReq.AddCookie(c)
}

/*
 * set header
 */
func (req *Request) SetHeader(key, value string) {
	req.RHttpReq.Header.Set(key, value)
}

/*
 * set user agent
 */
func (req *Request) SetUserAgent(ua string) {
	req.SetHeader("User-Agent", ua)
}

/*
 * set referer
 */
func (req *Request) SetReferer(referer string) {
	req.SetHeader("referer", referer)
}

/*
 * set proxy
 */
func (req *Request) SetProxy(proxy string) {
	req.RProxy = proxy
}

/*
 * set spider name
 */
func (req *Request) SetSpiderName(name string) {
	req.RSpiderName = name
}

func init() {
	gob.Register(errors.New(""))
}