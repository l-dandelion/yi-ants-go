package rpc

import (
	"github.com/l-dandelion/yi-ants-go/core/action"
	"github.com/l-dandelion/yi-ants-go/core/cluster"
	"github.com/l-dandelion/yi-ants-go/core/node"
	"github.com/l-dandelion/yi-ants-go/lib/library/log"
	"net"
	"net/rpc"
	"strconv"
	"github.com/l-dandelion/yi-ants-go/lib/constant"
)

const (
	RPC_TYPE = "tcp"
)

type RpcServer struct {
	node        node.Node
	cluster     cluster.Cluster
	port        int
	rpcClient   action.RpcClientAnts
	distributer action.Watcher
}

func NewRpcServer(node node.Node, cluster cluster.Cluster, port int, rpcClient action.RpcClientAnts, distributer action.Watcher) *RpcServer {
	rpcServer := &RpcServer{
		node, cluster, port, rpcClient, distributer,
	}
	rpcServer.start()
	return rpcServer
}

//listen
func (this *RpcServer) server() {
	rpc.Register(this)
	listener, err := net.Listen(RPC_TYPE, ":"+strconv.Itoa(this.port))
	if err != nil {
		log.Errorf("Server listen fail: %s", err)
		return
	}
	log.Infof("Listen...")
	for {
		if conn, err := listener.Accept(); err != nil {
			log.Errorf("Server accept fail: %s", err)
			return
		} else {
			log.Infof("New connection")
			go rpc.ServeConn(conn)
		}
	}
}

//start listen
func (this *RpcServer) start() {
	go this.server()
}

func (this *RpcServer) IsAlive(request *action.RpcBase, response *action.RpcBase) error {
	log.Infof("Local:%s Call Is Alive", this.node.GetNodeInfo().Name)
	response.NodeInfo = this.node.GetNodeInfo()
	response.Result = true
	return nil
}

//accept a request by myself
func (this *RpcServer) AcceptRequest(req *action.RpcRequest, resp *action.RpcError) error {
	err := this.node.AcceptRequest(req.Req)
	resp.Yierr = err
	resp.Result = err == nil
	resp.NodeInfo = this.node.GetNodeInfo()
	return nil
}

//start a spider named req.SpiderName
func (this *RpcServer) StartSpider(req *action.RpcSpiderName, resp *action.RpcError) error {
	if constant.RunMode == "dubug" {
		log.Infof("Local:%s Start Spider(%s)", this.node.GetNodeInfo().Name, req.SpiderName)
	}
	err := this.node.StartSpider(req.SpiderName)
	resp.Yierr = err
	resp.Result = err == nil
	resp.NodeInfo = this.node.GetNodeInfo()
	return nil
}

//pause a spider named req.SpiderName
func (this *RpcServer) PauseSpider(req *action.RpcSpiderName, resp *action.RpcError) error {
	err := this.node.PauseSpider(req.SpiderName)
	resp.Yierr = err
	resp.Result = err == nil
	resp.NodeInfo = this.node.GetNodeInfo()
	return nil
}

//Recover a spider named req.SpiderName
func (this *RpcServer) RecoverSpider(req *action.RpcSpiderName, resp *action.RpcError) error {
	err := this.node.RecoverSpider(req.SpiderName)
	resp.Yierr = err
	resp.Result = err == nil
	resp.NodeInfo = this.node.GetNodeInfo()
	return nil
}

//stop a spider named req.SpiderName
func (this *RpcServer) StopSpider(req *action.RpcSpiderName, resp *action.RpcError) error {
	err := this.node.StopSpider(req.SpiderName)
	resp.Yierr = err
	resp.Result = err == nil
	resp.NodeInfo = this.node.GetNodeInfo()
	return nil
}

//stop a spider named req.SpiderName
func (this *RpcServer) AddSpider(req *action.RpcSpider, resp *action.RpcError) error {
	resp.Yierr = this.node.AddSpider(req.Spider)
	resp.Result = resp.Yierr == nil
	resp.NodeInfo = this.node.GetNodeInfo()
	return nil
}

//first add a spider
func (this *RpcServer) FirstAddSpider(req *action.RpcSpider, resp *action.RpcError) error {
	resp.Yierr = this.rpcClient.AddSpider(req.Spider)
	resp.Result = resp.Yierr == nil
	resp.NodeInfo = this.node.GetNodeInfo()
	return nil
}

//add req.nodeINfo in cluster by myself and connect it
func (this *RpcServer) LetMeIn(req *action.RpcBase, resp *action.RpcError) error {
	err := this.rpcClient.Connect(req.NodeInfo.Ip, req.NodeInfo.Port)
	resp.Yierr = err
	if err != nil {
		log.Warnf("Let Me In Fail: %s", err)
	}
	resp.Result = err == nil
	resp.NodeInfo = this.node.GetNodeInfo()
	return nil
}

//get all node information
func (this *RpcServer) GetAllNode(req *action.RpcBase, resp *action.RpcNodeInfoList) error {
	resp.NodeInfo = this.node.GetNodeInfo()
	resp.Result = true
	resp.NodeInfoList = this.cluster.GetAllNode()
	return nil
}

// sign a request
func (this *RpcServer) SignRequest(req *action.RpcRequest, resp *action.RpcError) error {
	resp.NodeInfo = this.node.GetNodeInfo()
	resp.Yierr = this.node.SignRequest(req.Req)
	resp.Result = resp.Yierr == nil
	return nil
}