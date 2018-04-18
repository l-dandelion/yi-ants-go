package scheduler

import (
	"strings"

	"github.com/l-dandelion/yi-ants-go/core/module/data"
	"github.com/l-dandelion/yi-ants-go/lib/library/log"
	"github.com/l-dandelion/yi-ants-go/lib/constant"
)

/*
 * send request to request buffer pool
 */
func (sched *myScheduler) sendReq(req *data.Request) bool {
	if req == nil {
		return false
	}
	if sched.canceled() {
		return false
	}
	httpReq := req.HTTPReq()
	if httpReq == nil {
		log.Warnln("Ignore the request! Its HTTP request is invalid!")
		return false
	}
	reqURL := httpReq.URL
	if reqURL == nil {
		log.Warnln("Ignore the request! Its URL is invalid!")
		return false
	}
	scheme := strings.ToLower(reqURL.Scheme)
	if scheme != "http" && scheme != "https" {
		log.Warnf("Ignore the request! Its URL scheme is %q, but should be %q or %q. (URL: %s)\n", scheme, "http", "https", reqURL)
		return false
	}
	if v := sched.urlMap.Get(reqURL.String()); v != nil {
		log.Warnf("Ignore the request! Its URL is repeated. (URL: %s)\n", reqURL)
		return false
	}
	pd, _ := getPrimaryDomain(httpReq.Host)
	if sched.acceptedDomainMap.Get(pd) == nil {
		log.Warnf("Ignore the request! Its host %q is not in accepted primary domain map. (URL: %s)\n",
			httpReq.Host, reqURL)
		return false
	}
	if req.Depth() > sched.maxDepth {
		log.Warnf("Ignore the request! Its depth %d is greater than %d. (URL: %s)\n", req.Depth(), sched.maxDepth, reqURL)
		return false
	}
	go func(req *data.Request) {
		if err := sched.reqBufferPool.Put(req); err != nil {
			log.Warnln("The request buffer pool was closed. Ignore request sending.")
		}
	}(req)
	sched.urlMap.Put(reqURL.String(), struct{}{})
	return true
}


/*
 * send response to response buffer pool
 */
func (sched *myScheduler) sendResp(resp *data.Response) bool {
	respBufferPool := sched.respBufferPool
	if resp == nil || respBufferPool == nil || respBufferPool.Closed() {
		return false
	}
	go func(resp *data.Response) {
		if err := respBufferPool.Put(resp); err != nil {
			log.Warnln("The response buffer pool was closed. Ignore response sending.")
		}
	}(resp)
	return true
}


/*
 * send item to item buffer pool
 */
func (sched *myScheduler) sendItem(item data.Item) bool {
	itemBufferPool := sched.itemBufferPool
	if item == nil || itemBufferPool == nil || itemBufferPool.Closed() {
		return false
	}
	go func(item data.Item) {
		if err := itemBufferPool.Put(item); err != nil {
			log.Warnln("The item buffer pool was closed. Ignore item sending.")
		}
	}(item)
	return true
}


/*
 * send error to error buffer pool
 */
func (sched *myScheduler)sendError(yierr *constant.YiError) bool {
	errorBufferPool := sched.errorBufferPool
	if yierr == nil || errorBufferPool == nil || errorBufferPool.Closed() {
		return false
	}
	if errorBufferPool.Closed() {
		return false
	}
	go func(yierr *constant.YiError) {
		if err := errorBufferPool.Put(yierr); err != nil {
			log.Warnln("The error buffer pool was closed. Ignore error sending.")
		}
	}(yierr)
	return true
}
