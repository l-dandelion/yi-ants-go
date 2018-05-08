package action

import (
	"github.com/l-dandelion/yi-ants-go/core/module/data"
	"github.com/l-dandelion/yi-ants-go/core/node"
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
	//complile a spider
	ComplileSpider(req *RpcSpiderName, resp *RpcBase) error
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
	//sign a reuqest
	SignRequest(req RpcRequest, resp *RpcError) error
	//init a spider
	InitSpider(req *RpcSpiderName, resp *RpcError) error
	//delete a spider
	DeleteSpider(req *RpcSpiderName, resp *RpcError) error
	//get spiders status
	SpiderStatusList(req *RpcBase, resp *RpcBase) error
	//get distribute queue size
	GetDistributeQueueSize(req *RpcBase, resp *RpcNum) error
	//can start spider
	CanInitSpider(req *RpcSpiderName, resp *RpcError) error
	//get spider status by spider name
	GetSpiderStatusBySpiderName(req *RpcSpiderName, resp *RpcError) error
	//get node score
	GetNodeScore(req *RpcBase, resp *RpcNum) error
	//accept requests by myself
	AcceptRequests(req *RpcRequestList, resp *RpcBase) error
	//filter requests
	FilterRequests(req *RpcRequestList, resp *RpcRequestList) error
	//sign requests
	SignRequests(req *RpcRequestList, resp *RpcError) error
}

type RpcServerCluster interface {
	LetMeIn(req *RpcBase, resp *RpcBase) error
	GetAllNode(req *RpcBase, resp *RpcNodeInfoList) error
	GetNodeInfo(req *RpcBase, resp *RpcNodeInfoList) error
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
	//call node.RpcServer.AcceptRequests(req)
	DistributeRequests(nodeName string, reqs []*data.Request) *constant.YiError
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
	//call all node to sign a request
	SignRequest(req *data.Request) *constant.YiError
	//call all node to sign requests
	SignRequests(reqs []*data.Request)
	//call all node to init a spider
	InitSpider(spiderName string) *constant.YiError
	//call all node to complile a spider
	ComplileSpider(spiderName string) *constant.YiError
	//call all node to delete a spider
	DeleteSpider(spiderName string) *constant.YiError
	//get all spider status from all node
	SpiderStatusList() ([]*spider.SpiderStatus, *constant.YiError)
	//get spider status from node named nodeName
	GetSpiderStatusListByNodeName(nodeName string) ([]*spider.SpiderStatus, *constant.YiError)
	//get distribute queue size
	GetDistributeQueueSize(nodeName string) (uint64, *constant.YiError)
	//get information of node named nodeName
	GetNodeInfoByNodeName(nodeName string) (*node.NodeInfo, *constant.YiError)
	//can we start spider
	CanWeInitSpider(spiderName string) *constant.YiError
	//get spider status map from all node
	GetSpiderStatusMapBySpiderName(spiderName string) (map[string]*spider.SpiderStatus, *constant.YiError)
	//get node score
	GetNodeScore(nodeName string) (uint64, *constant.YiError)
	//filter requests
	FilterRequests([]*data.Request) ([]*data.Request)
}

type RpcClientAnts interface {
	RpcClient
	RpcClientCrawl
	RpcClientCluster
}
