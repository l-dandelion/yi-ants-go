package rpc

import (
	"github.com/l-dandelion/yi-ants-go/core/spider"
	"github.com/l-dandelion/yi-ants-go/lib/constant"
)

//合并不同节点的爬虫信息
func megerSpiderStatus(a *spider.SpiderStatus, b *spider.SpiderStatus) *spider.SpiderStatus {
	if a == nil {
		return b
	}
	if b.CompilingStatus != constant.COMPLILING_STATUS_COMPLILED {
		a.CompilingStatus = b.CompilingStatus
		a.ComplilingError = b.ComplilingError
	}
	a.Crawled += b.Crawled
	a.Running += b.Running
	a.Success += b.Success
	a.Waiting += b.Waiting
	return a
}
