package action

import (
	"github.com/l-dandelion/yi-ants-go/core/module/data"
	"github.com/l-dandelion/yi-ants-go/core/node"
	"github.com/l-dandelion/yi-ants-go/core/spider"
	"github.com/l-dandelion/yi-ants-go/lib/constant"
	"github.com/l-dandelion/yi-ants-go/core/crawler"
)

type RpcBase struct {
	NodeInfo *node.NodeInfo
	Result   bool
}

type RpcRequest struct {
	RpcBase
	Req *data.Request
}

type RpcError struct {
	NodeInfo *node.NodeInfo
	Result bool
	Yierr *constant.YiError
}

type RpcSpiderName struct {
	RpcBase
	SpiderName string
}

type RpcSpider struct {
	RpcBase
	Spider spider.Spider
}

type RpcNodeInfoList struct {
	RpcBase
	NodeInfoList []*node.NodeInfo
}

type RpcSpiderStatusList struct {
	RpcBase
	SpiderStatusList []*spider.SpiderStatus
}

type RpcNum struct {
	Num uint64
	RpcBase
}

type RpcSpiderStatus struct {
	RpcBase
	SpiderStatus *spider.SpiderStatus
	Yierr *constant.YiError
}

type RpcRequestList struct {
	RpcBase
	Reqs []*data.Request
}

type RpcCrawlerSummary struct {
	RpcBase
	Summary *crawler.Summary
}