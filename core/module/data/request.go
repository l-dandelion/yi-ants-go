package data

import (
	"net/http"
)

/*
 * Request
 * httpReq: the http request
 * depth: crawl depth
 * proxy: use proxy if not empty
 * extra: additional information(used for context)
 */
type Request struct {
	nodeName   string
	spiderName string
	httpReq    *http.Request          // the http request
	depth      uint32                 // crawl depth
	proxy      string                 // use proxy if not empty
	Extra      map[string]interface{} // additional information(used for context)
}

/*
 * get spider name
 */
func (req *Request) SpiderName() string {
	return req.spiderName
}

/*
 * get node name
 */
func (req *Request) NodeName() string {
	return req.nodeName
}

/*
 * set node name
 */
func (req *Request) SetNodeName(nodeName string) {
	req.nodeName = nodeName
}

/*
 * get http request
 */
func (req *Request) HTTPReq() *http.Request {
	return req.httpReq
}

/*
 * get crawl depth
 */
func (req *Request) Depth() uint32 {
	return req.depth
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
	req.depth = depth
}

/*
 * check the request
 */
func (req *Request) Valid() bool {
	return req.httpReq != nil && req.httpReq.URL != nil
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
		httpReq: httpReq,
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
	req.httpReq.AddCookie(c)
}

/*
 * set header
 */
func (req *Request) SetHeader(key, value string) {
	req.httpReq.Header.Set(key, value)
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
	req.proxy = proxy
}
