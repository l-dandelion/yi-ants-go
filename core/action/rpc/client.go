package rpc

import (
	"github.com/l-dandelion/yi-ants-go/core/action"
	"github.com/l-dandelion/yi-ants-go/core/cluster"
	"github.com/l-dandelion/yi-ants-go/core/module/data"
	"github.com/l-dandelion/yi-ants-go/core/node"
	"github.com/l-dandelion/yi-ants-go/core/spider"
	"github.com/l-dandelion/yi-ants-go/lib/constant"
	"github.com/l-dandelion/yi-ants-go/lib/library/log"
	"net/rpc"
	"strconv"
	"time"
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
	log.Infof("Local:%s Call Dial(%s, %d)", this.node.GetNodeInfo().Name, ip, port)
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
	log.Infof("Local: %s Connect to node: %s:%d\n", this.node.GetNodeInfo().Name, ip, port)
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
	if constant.RunMode == "debug" {
		log.Infof("Distribute NodoName: %s, req: %v", nodeName, req)
	}
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

//call all node to start spider named spiderName
func (this *RpcClient) StartSpider(spiderName string) (yierr *constant.YiError) {
	if constant.RunMode == "debug" {
		log.Infof("LocalNode: %s, StartSpider(%s)", this.node.GetNodeInfo().Name, spiderName)
	}
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
	if constant.RunMode == "debug" {
		log.Infof("%v", spider.GetInitReqs()[0])
	}
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