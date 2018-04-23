package action

import (
	"github.com/l-dandelion/yi-ants-go/core/module/data"
	"github.com/l-dandelion/yi-ants-go/core/spider"
	"github.com/l-dandelion/yi-ants-go/lib/constant"
	"net/rpc"
)

type RpcServer interface {
	IsAlive(request *RpcBase, response *RpcBase) error
}

type RpcServerCrawl interface {
	//accept a request by myself
	AcceptRequest(req *RpcRequest, resp *RpcBase) error
	//start a spider named req.SpiderName by myself
	StartSpider(req *RpcSpiderName, resp *RpcBase) error
	//pause a spider named req.SpiderName by myself
	PauseSpider(req *RpcSpiderName, resp *RpcBase) error
	//recover a spider named req.SpiderName by myself
	RecoverSpider(req *RpcSpiderName, resp *RpcBase) error
	//stop a spider named req.SpiderName by myself
	StopSpider(req *RpcSpiderName, resp *RpcBase) error
	//add a spider named by myself
	AddSpider(req RpcSpider, resp *RpcBase) error
}

type RpcServerCluster interface {
	LetMeIn(req *RpcBase, resp *RpcBase) error
	GetAllNode(req *RpcBase, resp *RpcNodeInfoList) error
}

type RpcServerAnts interface {
	RpcServer
	RpcServerCrawl
	RpcServerCluster
}

type RpcClient interface {
	//connect to node(ip:port) and return an client
	Dial(ip string, port int) (*rpc.Client, *constant.YiError)
	//check the node list, remove the dead node
	Detect()
	//start Detect (cycle)
	Start()
}

type RpcClientCluster interface {
	// join cluster which node(ip:port) joined
	// Do: get node info list from node(ip:port) and connect one by one
	LetMeIn(ip string, port int) *constant.YiError
	// connect node(ip:port) and store
	Connect(ip string, port int) *constant.YiError
}

type RpcClientCrawl interface {
	//call node.RpcServer.AcceptRequest(req)
	Distribute(nodeName string, req *data.Request) *constant.YiError
	//call all node to start spider named spiderName
	StartSpider(spiderName string) *constant.YiError
	//call all node to stop spider named spiderName
	StopSpider(spiderName string) *constant.YiError
	//call all node to pause spider named spiderName
	PauseSpider(spiderName string) *constant.YiError
	//call all node to recover spider named spiderName
	RecoverSpider(spiderName string) *constant.YiError
	//call all node to add spider
	AddSpider(spider spider.Spider) *constant.YiError
}

type RpcClientAnts interface {
	RpcClient
	RpcClientCrawl
	RpcClientCluster
}
