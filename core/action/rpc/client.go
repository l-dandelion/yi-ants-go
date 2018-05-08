package rpc

import (
	"github.com/l-dandelion/yi-ants-go/core/action"
	"github.com/l-dandelion/yi-ants-go/core/cluster"
	"github.com/l-dandelion/yi-ants-go/core/module/data"
	"github.com/l-dandelion/yi-ants-go/core/node"
	"github.com/l-dandelion/yi-ants-go/core/spider"
	"github.com/l-dandelion/yi-ants-go/lib/constant"
	log "github.com/sirupsen/logrus"
	"net/rpc"
	"strconv"
	"time"
	"fmt"
)

type RpcClient struct {
	node    node.Node
	cluster cluster.Cluster
	connMap map[string]*rpc.Client
	//TODO:connMap 并发安全
}

func NewRpcClient(node node.Node, cluster cluster.Cluster) *RpcClient {
	connMap := make(map[string]*rpc.Client)
	return &RpcClient{
		node:    node,
		cluster: cluster,
		connMap: connMap,
	}
}

//connect to node(ip:port) and return an client
func (this *RpcClient) Dial(ip string, port int) (*rpc.Client, *constant.YiError) {
	client, err := rpc.Dial(RPC_TYPE, ip+":"+strconv.Itoa(port))
	if err != nil {
		return nil, constant.NewYiErrore(constant.ERR_RPC_CLIENT_DIAL, err)
	}
	return client, nil
}

//check the node list, remove the dead node
func (this *RpcClient) Detect() {
	request := new(action.RpcBase)
	response := new(action.RpcBase)
	for key, conn := range this.connMap {
		err := conn.Call("RpcServer.IsAlive", request, response)
		if err != nil {
			log.Errorf("Node %s is dead, so remove it. Error: %s", err)
			delete(this.connMap, key)
			this.cluster.DeleteDeadNode(key)
		}
	}
}

//start Detect (cycle)
func (this *RpcClient) Start() {
	go func() {
		for {
			this.Detect()
			time.Sleep(5 * time.Second)
		}
	}()
}

// join cluster which node(ip:port) joined
// Do: get node info list from node(ip:port) and connect one by one
func (this *RpcClient) LetMeIn(ip string, port int) *constant.YiError {
	client, yierr := this.Dial(ip, port)
	if yierr != nil {
		return yierr
	}
	req := new(action.RpcBase)
	req.NodeInfo = this.node.GetNodeInfo()
	resp := new(action.RpcNodeInfoList)
	err := client.Call("RpcServer.GetAllNode", req, resp)
	client.Close()
	if err != nil {
		return constant.NewYiErrorf(constant.ERR_RPC_CALL,
			"Get all node fail when join: %s, IP: %s, Port: %d", err, ip, port)
	}

	if resp.Result {
		for _, nodeInfo := range resp.NodeInfoList {
			if !this.node.IsMe(nodeInfo.Name) {
				yierr := this.Connect(nodeInfo.Ip, nodeInfo.Port)
				if yierr != nil {
					log.Error(yierr)
					continue
				}
				req := new(action.RpcBase)
				req.NodeInfo = this.node.GetNodeInfo()
				resp := new(action.RpcBase)
				client, ok := this.connMap[nodeInfo.Name]
				if !ok || client == nil {
					log.Error("Get Node(%s) connect fail", nodeInfo.Name)
					continue
				}
				err := client.Call("RpcServer.LetMeIn", req, resp)
				if err != nil {
					yierr = constant.NewYiErrorf(constant.ERR_RPC_CALL,
						"Connect to node fail: %s, IP: %s, Port: %d", err, nodeInfo.Ip, nodeInfo.Port)
					log.Error(yierr)
				}
			}
		}
	}
	return nil
}

// connect node(ip:port) and store
func (this *RpcClient) Connect(ip string, port int) *constant.YiError {
	client, yierr := this.Dial(ip, port)
	if yierr != nil {
		return yierr
	}
	req := new(action.RpcBase)
	req.NodeInfo = this.node.GetNodeInfo()
	resp := new(action.RpcBase)
	err := client.Call("RpcServer.IsAlive", req, resp)
	log.Infof("NodeInfo: %v", resp.NodeInfo)
	if err == nil {
		this.connMap[resp.NodeInfo.Name] = client
		this.cluster.AddNode(resp.NodeInfo)
		return nil
	}
	client.Close()
	return constant.NewYiErrorf(constant.ERR_RPC_CALL,
		"Connect to node fail: %s, IP: %s, Port: %d", err, ip, port)
}

//call node.RpcServer.AcceptRequest(req)
func (this *RpcClient) Distribute(nodeName string, req *data.Request) (yierr *constant.YiError) {
	if this.node.IsMe(nodeName) {
		return this.node.AcceptRequest(req)
	}
	distributeReq := &action.RpcRequest{}
	distributeReq.NodeInfo = this.node.GetNodeInfo()
	distributeReq.Req = req
	distributeResp := &action.RpcError{}
	err := this.connMap[nodeName].Call("RpcServer.AcceptRequest", distributeReq, distributeResp)
	if err != nil {
		yierr = constant.NewYiErrorf(constant.ERR_RPC_CALL,
			"Distribute fail. NodeName: %s, req: %v", nodeName, req)
	} else {
		if distributeResp.Yierr != nil {
			yierr = distributeResp.Yierr
		}
	}
	return
}

func (this *RpcClient) DistributeRequests(nodeName string, reqs []*data.Request) (yierr *constant.YiError) {
	if this.node.IsMe(nodeName) {
		this.node.AcceptRequests(reqs)
		return
	}
	distributeReqs := &action.RpcRequestList{}
	distributeReqs.NodeInfo = this.node.GetNodeInfo()
	distributeReqs.Reqs = reqs
	distributeResp := &action.RpcError{}
	err := this.connMap[nodeName].Call("RpcServer.AcceptRequests", distributeReqs, distributeResp)
	if err != nil {
		yierr = constant.NewYiErrorf(constant.ERR_RPC_CALL,
			"Distribute fail. NodeName: %s", nodeName)
	}
	return
}

//call all node to start spider named spiderName
func (this *RpcClient) StartSpider(spiderName string) (yierr *constant.YiError) {
	nodeInfoList := this.cluster.GetAllNode()
	for _, nodeInfo := range nodeInfoList {
		if this.node.IsMe(nodeInfo.Name) {
			yierr := this.node.FirstStartSpider(spiderName)
			if yierr != nil {
				return yierr
			}
		} else {
			go func() {
				req := &action.RpcSpiderName{
					SpiderName: spiderName,
				}
				req.NodeInfo = this.node.GetNodeInfo()
				resp := &action.RpcError{}
				err := this.connMap[nodeInfo.Name].Call("RpcServer.StartSpider", req, resp)
				if err != nil {
					yierr = constant.NewYiErrorf(constant.ERR_RPC_CALL,
						"Start spider fail, Node: %s, SpiderName: %s, ERROR: %s", nodeInfo.Name, spiderName, err)
					log.Error(yierr)
				}
				if resp.Yierr != nil {
					yierr = resp.Yierr
					log.Error("Start spider fail, Node: %s, SpiderName: %s, ERROR: %s", nodeInfo.Name, spiderName, yierr)
				}
			}()
		}
	}
	return nil
}

//call all node to stop spider named spiderName
func (this *RpcClient) StopSpider(spiderName string) (yierr *constant.YiError) {
	nodeInfoList := this.cluster.GetAllNode()
	for _, nodeInfo := range nodeInfoList {
		if this.node.IsMe(nodeInfo.Name) {
			this.node.StopSpider(spiderName)
		} else {
			req := &action.RpcSpiderName{
				SpiderName: spiderName,
			}
			req.NodeInfo = this.node.GetNodeInfo()
			resp := &action.RpcError{}
			err := this.connMap[nodeInfo.Name].Call("RpcServer.StopSpider", req, resp)
			if err != nil {
				yierr = constant.NewYiErrorf(constant.ERR_RPC_CALL,
					"Stop spider fail, Node: %s, SpiderName: %s, ERROR: %s", nodeInfo.Name, spiderName, err)
				log.Error(yierr)
			}
			if resp.Yierr != nil {
				yierr = resp.Yierr
				log.Error("Stop spider fail, Node: %s, SpiderName: %s, ERROR: %s", nodeInfo.Name, spiderName, yierr)
			}
		}
	}
	return nil
}

//call all node to pause spider named spiderName
func (this *RpcClient) PauseSpider(spiderName string) (yierr *constant.YiError) {
	nodeInfoList := this.cluster.GetAllNode()
	for _, nodeInfo := range nodeInfoList {
		if this.node.IsMe(nodeInfo.Name) {
			this.node.PauseSpider(spiderName)
		} else {
			req := &action.RpcSpiderName{
				SpiderName: spiderName,
			}
			req.NodeInfo = this.node.GetNodeInfo()
			resp := &action.RpcError{}
			err := this.connMap[nodeInfo.Name].Call("RpcServer.PauseSpider", req, resp)
			if err != nil {
				yierr = constant.NewYiErrorf(constant.ERR_RPC_CALL,
					"Pause spider fail, Node: %s, SpiderName: %s, ERROR: %s", nodeInfo.Name, spiderName, err)
				log.Error(yierr)
			}
			if resp.Yierr != nil {
				yierr = resp.Yierr
				log.Error("Pause spider fail, Node: %s, SpiderName: %s, ERROR: %s", nodeInfo.Name, spiderName, yierr)
			}
		}
	}
	return nil
}

//call all node to recover spider named spiderName
func (this *RpcClient) RecoverSpider(spiderName string) (yierr *constant.YiError) {
	nodeInfoList := this.cluster.GetAllNode()
	for _, nodeInfo := range nodeInfoList {
		if this.node.IsMe(nodeInfo.Name) {
			this.node.RecoverSpider(spiderName)
		} else {
			req := &action.RpcSpiderName{
				SpiderName: spiderName,
			}
			req.NodeInfo = this.node.GetNodeInfo()
			resp := &action.RpcError{}
			err := this.connMap[nodeInfo.Name].Call("RpcServer.RecoverSpider", req, resp)
			if err != nil {
				yierr = constant.NewYiErrorf(constant.ERR_RPC_CALL,
					"Recover spider fail, Node: %s, SpiderName: %s, ERROR: %s", nodeInfo.Name, spiderName, err)
				log.Error(yierr)
			}
			if resp.Yierr != nil {
				yierr = resp.Yierr
				log.Error("Recover spider fail, Node: %s, SpiderName: %s, ERROR: %s", nodeInfo.Name, spiderName, yierr)
			}
		}
	}
	return nil
}

//call all node to add spider
func (this *RpcClient) AddSpider(spider spider.Spider) (yierr *constant.YiError) {
	nodeInfoList := this.cluster.GetAllNode()
	csp := spider.Copy()
	yierr = this.node.AddSpider(spider)
	if yierr != nil {
		return
	}
	for _, nodeInfo := range nodeInfoList {
		if !this.node.IsMe(nodeInfo.Name) {
			go func() {
				req := &action.RpcSpider{
					Spider: csp,
				}
				req.NodeInfo = this.node.GetNodeInfo()
				resp := &action.RpcError{}
				err := this.connMap[nodeInfo.Name].Call("RpcServer.AddSpider", req, resp)
				if err != nil {
					yierr = constant.NewYiErrorf(constant.ERR_RPC_CALL,
						"Add spider fail, Node: %s, spider: %v, ERROR: %s", nodeInfo.Name, spider, err)
					log.Error(yierr)
				}
				if resp.Yierr != nil {
					yierr = resp.Yierr
					log.Errorf("Add spider fail, Node: %s, spider: %v, ERROR: %s", nodeInfo.Name, spider, yierr)
				}
			}()
		}
	}
	return nil
}

//call all node to sign request
func (this *RpcClient) SignRequest(dreq *data.Request) (yierr *constant.YiError) {
	nodeInfoList := this.cluster.GetAllNode()
	for _, nodeInfo := range nodeInfoList {
		if !this.node.IsMe(nodeInfo.Name) {
			go func() {
				req := &action.RpcRequest{
					Req: dreq,
				}
				req.NodeInfo = this.node.GetNodeInfo()
				resp := &action.RpcError{}
				err := this.connMap[nodeInfo.Name].Call("RpcServer.SignRequest", req, resp)
				if err != nil {
					yierr = constant.NewYiErrorf(constant.ERR_RPC_CALL,
						"Sign request fail, Node: %s, Request: %v, ERROR: %s", nodeInfo.Name, dreq, err)
					log.Error(yierr)
				}
				if resp.Yierr != nil {
					yierr = resp.Yierr
					log.Errorf("Sign request fail, Node: %s, Request: %v, ERROR: %s", nodeInfo.Name, dreq, yierr)
				}
			}()
		} else {
			yierr := this.node.SignRequest(dreq)
			if yierr != nil {
				log.Errorf("Sign request fail, Node: %s, Request: %v, ERROR: %s", nodeInfo.Name, dreq, yierr)
			}
		}
	}
	return nil
}

func (this *RpcClient) InitSpider(spiderName string) (yierr *constant.YiError) {
	yierr = this.CanWeInitSpider(spiderName)
	if yierr != nil {
		return
	}
	nodeInfoList := this.cluster.GetAllNode()
	yierr = this.node.InitSpider(spiderName)
	if yierr != nil {
		return
	}
	for _, nodeInfo := range nodeInfoList {
		if !this.node.IsMe(nodeInfo.Name) {
			//go func() {
				req := &action.RpcSpiderName{
					SpiderName: spiderName,
				}
				req.NodeInfo = this.node.GetNodeInfo()
				resp := &action.RpcError{}
				err := this.connMap[nodeInfo.Name].Call("RpcServer.InitSpider", req, resp)
				if err != nil {
					yierr = constant.NewYiErrorf(constant.ERR_RPC_CALL,
						"Init spider fail, Node: %s, spiderName: %s, ERROR: %s", nodeInfo.Name, spiderName, err)
					return
				}
				if resp.Yierr != nil {
					yierr = resp.Yierr
					log.Errorf("Init spider fail, Node: %s, spiderName: %v, ERROR: %s", nodeInfo.Name, spiderName, yierr)
					return
				}
			//}()
		}
	}
	return nil
}


func (this *RpcClient) ComplileSpider(spiderName string) (yierr *constant.YiError) {
	nodeInfoList := this.cluster.GetAllNode()
	go this.node.ComplileSpider(spiderName)
	for _, nodeInfo := range nodeInfoList {
		if !this.node.IsMe(nodeInfo.Name) {
			go func() {
				req := &action.RpcSpiderName{
					SpiderName: spiderName,
				}
				req.NodeInfo = this.node.GetNodeInfo()
				resp := &action.RpcError{}
				err := this.connMap[nodeInfo.Name].Call("RpcServer.ComplileSpider", req, resp)
				if err != nil {
					yierr = constant.NewYiErrorf(constant.ERR_RPC_CALL,
						"Complile spider fail, Node: %s, spiderName: %s, ERROR: %s", nodeInfo.Name, spiderName, err)
					log.Error(yierr)
				}
				if resp.Yierr != nil {
					yierr = resp.Yierr
					log.Errorf("Complile spider fail, Node: %s, spiderName: %v, ERROR: %s", nodeInfo.Name, spiderName, yierr)
				}
			}()
		}
	}
	return nil
}

func (this *RpcClient) DeleteSpider(spiderName string) (yierr *constant.YiError) {
	nodeInfoList := this.cluster.GetAllNode()
	yierr = this.node.DeleteSpider(spiderName)
	if yierr != nil {
		return
	}
	for _, nodeInfo := range nodeInfoList {
		if !this.node.IsMe(nodeInfo.Name) {
			go func() {
				req := &action.RpcSpiderName{
					SpiderName: spiderName,
				}
				req.NodeInfo = this.node.GetNodeInfo()
				resp := &action.RpcError{}
				err := this.connMap[nodeInfo.Name].Call("RpcServer.DeleteSpider", req, resp)
				if err != nil {
					yierr = constant.NewYiErrorf(constant.ERR_RPC_CALL,
						"Delete spider fail, Node: %s, spiderName: %s, ERROR: %s", nodeInfo.Name, spiderName, err)
					log.Error(yierr)
				}
				if resp.Yierr != nil {
					yierr = resp.Yierr
					log.Errorf("Delete spider fail, Node: %s, spiderName: %v, ERROR: %s", nodeInfo.Name, spiderName, yierr)
				}
			}()
		}
	}
	return nil
}

func (this *RpcClient) SpiderStatusList() (ssl []*spider.SpiderStatus, yierr *constant.YiError) {
	nodeInfoList := this.cluster.GetAllNode()
	mssl := []*spider.SpiderStatus{}
	ssm := make(map[string]*spider.SpiderStatus)
	for _, nodeInfo := range nodeInfoList {
		if this.node.IsMe(nodeInfo.Name) {
			mssl = this.node.GetSpiderStatusList()
		} else {
			req := &action.RpcBase{}
			resp := &action.RpcSpiderStatusList{}
			err := this.connMap[nodeInfo.Name].Call("RpcServer.SpiderStatusList", req, resp)
			if err != nil {
				yierr = constant.NewYiErrorf(constant.ERR_RPC_CALL,
					"get spider status list fail, Node: %s ERROR: %s", nodeInfo.Name, err)
				return
			}
			mssl = resp.SpiderStatusList
		}
		fmt.Println(mssl)
		for _, ss := range mssl {
			ssm[ss.Name] = megerSpiderStatus(ssm[ss.Name], ss)
		}
	}
	ssl = []*spider.SpiderStatus{}
	for _, ss := range ssm {
		ssl = append(ssl, ss)
	}
	return ssl, nil
}

func (this *RpcClient) GetSpiderStatusListByNodeName(nodeName string) (ssl []*spider.SpiderStatus, yierr *constant.YiError) {
	if this.node.IsMe(nodeName) {
		return this.node.GetSpiderStatusList(), nil
	}
	req := &action.RpcBase{}
	resp := &action.RpcSpiderStatusList{}
	client := this.connMap[nodeName]
	if client == nil {
		return nil, constant.NewYiErrorf(constant.ERR_NODE_NOT_FOUND, "Node not found.(NodeName: %s)", nodeName)
	}
	err := client.Call("RpcServer.SpiderStatusList", req, resp)
	if err != nil {
		yierr = constant.NewYiErrorf(constant.ERR_RPC_CALL,
			"get spider status list fail, Node: %s ERROR: %s", nodeName, err)
		return
	}
	ssl = resp.SpiderStatusList
	return
}

func (this *RpcClient) GetDistributeQueueSize(nodeName string) (total uint64, yierr *constant.YiError) {
	if this.node.IsMe(nodeName) {
		return this.node.GetDistributeQueueSize(), nil
	}
	client := this.connMap[nodeName]
	if client == nil {
		yierr = constant.NewYiErrorf(constant.ERR_NODE_NOT_FOUND, "Node not found.(NodeName: %s)", nodeName)
		return
	}
	req := &action.RpcBase{}
	resp := &action.RpcNum{}
	err := client.Call("RpcServer.GetDistributeQueueSize", req, resp)
	if err != nil {
		yierr = constant.NewYiErrorf(constant.ERR_RPC_CALL,
			"get distribute queue size fail, Node: %s ERROR: %s", nodeName, err)
		return
	}
	return resp.Num, nil
}

func (this *RpcClient) GetNodeInfoByNodeName(nodeName string) (nodeInfo *node.NodeInfo, yierr *constant.YiError) {
	if this.node.IsMe(nodeName) {
		return this.node.GetNodeInfo(), nil
	}
	client := this.connMap[nodeName]
	if client == nil {
		yierr = constant.NewYiErrorf(constant.ERR_NODE_NOT_FOUND, "Node not found.(NodeName: %s)", nodeName)
		return
	}
	req := &action.RpcBase{}
	resp := &action.RpcBase{}
	err := client.Call("RpcServer.GetNodeInfo", req, resp)
	if err != nil {
		yierr = constant.NewYiErrorf(constant.ERR_RPC_CALL,
			"get spider status list fail, Node: %s ERROR: %s", nodeName, err)
		return
	}
	return resp.NodeInfo, nil
}

func (this *RpcClient) CanWeInitSpider(spiderName string) (yierr *constant.YiError) {
	nodeInfoList := this.cluster.GetAllNode()
	var ok bool
	for _, nodeInfo := range nodeInfoList {
		if this.node.IsMe(nodeInfo.Name) {
			ok, yierr = this.node.CanStartSpider(spiderName)
		} else {
			req := &action.RpcSpiderName{SpiderName:spiderName}
			resp := &action.RpcError{}
			err := this.connMap[nodeInfo.Name].Call("RpcServer.CanInitSpider", req, resp)
			if err != nil {
				yierr = constant.NewYiErrorf(constant.ERR_RPC_CALL,
					"Query whether we can init spider fail, Node: %s ERROR: %s", nodeInfo.Name, err)
				return
			}
			ok, yierr = resp.Result, resp.Yierr
		}
		if yierr != nil {
			return
		}
		if !ok {
			return constant.NewYiErrorf(constant.ERR_NOT_COMPLILED, "The spdier is not compiled.(SpiderName: %s, NodeName: %s)", spiderName, nodeInfo.Name)
		}
	}
	return nil
}

func (this *RpcClient) GetSpiderStatusMapBySpiderName(spiderName string) (spiderStatusMap map[string]*spider.SpiderStatus, yierr *constant.YiError){
	nodeInfoList := this.cluster.GetAllNode()
	var spiderStatus *spider.SpiderStatus
	spiderStatusMap = make(map[string]*spider.SpiderStatus)
	for _, nodeInfo := range nodeInfoList {
		if this.node.IsMe(nodeInfo.Name) {
			spiderStatus, yierr = this.node.GetSpiderStatus(spiderName)
		} else {
			req := &action.RpcSpiderName{SpiderName:spiderName}
			resp := &action.RpcSpiderStatus{}
			err := this.connMap[nodeInfo.Name].Call("RpcServer.GetSpiderStatusBySpiderName", req, resp)
			if err != nil {
				yierr = constant.NewYiErrorf(constant.ERR_RPC_CALL,
					"Get spider status by spider name fail, Node: %s ERROR: %s", nodeInfo.Name, err)
				return
			}
			yierr = resp.Yierr
			spiderStatus = resp.SpiderStatus
		}
		if yierr != nil {
			return
		}
		spiderStatusMap[nodeInfo.Name] = spiderStatus
	}
	return
}

func (this *RpcClient) GetNodeScore(nodeName string) (score uint64, yierr *constant.YiError) {
	if this.node.IsMe(nodeName) {
		score = this.node.GetScore()
		return
	}
	client := this.connMap[nodeName]
	if client == nil {
		yierr = constant.NewYiErrorf(constant.ERR_NODE_NOT_FOUND, "Node not found.(NodeName: %s)", nodeName)
		return
	}
	req := &action.RpcBase{}
	resp := &action.RpcNum{}
	err := client.Call("RpcServer.GetNodeScore", req, resp)
	if err != nil {
		yierr = constant.NewYiErrorf(constant.ERR_RPC_CALL,
			"get node score fail, Node: %s ERROR: %s", nodeName, err)
		return
	}
	return resp.Num, nil
}

func (this *RpcClient) FilterRequests(reqs []*data.Request) []*data.Request {
	nodeInfoList := this.cluster.GetAllNode()
	for _, nodeInfo := range nodeInfoList {
		if !this.node.IsMe(nodeInfo.Name) {
			req := &action.RpcRequestList{Reqs:reqs}
			resp := &action.RpcRequestList{}
			err := this.connMap[nodeInfo.Name].Call("RpcServer.FilterRequests", req, resp)
			if err != nil {
				yierr := constant.NewYiErrorf(constant.ERR_RPC_CALL,
					"Call RpcServer.FilterRequests Fail, NodeName: %s, ERROR: %s", nodeInfo.Name, err)
				log.Error(yierr)
				continue
			}
			reqs = resp.Reqs
		}
	}
	return reqs
}

//sign requests
func (this *RpcClient) SignRequests(reqs []*data.Request) {
	nodeInfoList := this.cluster.GetAllNode()
	for _, nodeInfo := range nodeInfoList {
		if !this.node.IsMe(nodeInfo.Name) {
			req := &action.RpcRequestList{Reqs:reqs}
			resp := &action.RpcBase{}
			err := this.connMap[nodeInfo.Name].Call("RpcServer.SignRequests", req, resp)
			if err != nil {
				yierr := constant.NewYiErrorf(constant.ERR_RPC_CALL,
					"Call RpcServer.SignRequests Fail, NodeName: %s, ERROR: %s", nodeInfo.Name, err)
				log.Error(yierr)
				continue
			}
		}
	}
}