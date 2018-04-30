package watcher

import (
	"github.com/l-dandelion/yi-ants-go/core/action"
	"github.com/l-dandelion/yi-ants-go/core/cluster"
	"github.com/l-dandelion/yi-ants-go/core/module/data"
	"github.com/l-dandelion/yi-ants-go/core/node"
	"github.com/l-dandelion/yi-ants-go/lib/constant"
	"github.com/l-dandelion/yi-ants-go/lib/library/log"
	"github.com/l-dandelion/yi-ants-go/lib/library/pool"
	"sync"
	"time"
)

type Distributer struct {
	sync.RWMutex
	Status    int8
	LastIndex int
	Cluster   cluster.Cluster
	RpcClient action.RpcClientAnts
	Node      node.Node
	MaxThread int
	pool      *pool.Pool
}

func NewDistributer(mnode node.Node, cluster cluster.Cluster, rpcClient action.RpcClientAnts) *Distributer {
	return &Distributer{
		Status:    constant.RUNNING_STATUS_STOPPED,
		Cluster:   cluster,
		RpcClient: rpcClient,
		Node:      mnode,
		MaxThread: 10,
		pool:      pool.NewPool(10),
	}
}

func (this *Distributer) IsStop() bool {
	this.RLock()
	defer this.RUnlock()
	return this.Status == constant.RUNNING_STATUS_STOPPED
}

func (this *Distributer) IsRunning() bool {
	this.RLock()
	defer this.RUnlock()
	return this.Status == constant.RUNNING_STATUS_STARTING
}

func (this *Distributer) IsPause() bool {
	this.RLock()
	defer this.RUnlock()
	return this.Status == constant.RUNNING_STATUS_PAUSED
}

func (this *Distributer) IsStopping() bool {
	this.RLock()
	defer this.RUnlock()
	return this.Status == constant.RUNNING_STATUS_STOPPING
}

func (this *Distributer) Pause() {
	this.Lock()
	defer this.Unlock()
	if this.Status == constant.RUNNING_STATUS_STARTING {
		this.Status = constant.RUNNING_STATUS_PAUSED
	}
}

func (this *Distributer) UnPause() {
	this.Lock()
	defer this.Unlock()
	if this.Status == constant.RUNNING_STATUS_PAUSED {
		this.Status = constant.RUNNING_STATUS_STARTED
	}
}

func (this *Distributer) Stop() {
	this.Lock()
	defer this.Unlock()
	if this.Status != constant.RUNNING_STATUS_STOPPED {
		this.Status = constant.RUNNING_STATUS_STOPPING
	}
}

func (this *Distributer) Start() {
	if this.IsRunning() {
		return
	}
	for {
		if this.IsStop() {
			break
		}
		time.Sleep(1 * time.Second)
	}
	this.Lock()
	defer this.Unlock()
	this.Status = constant.RUNNING_STATUS_STARTED
	go this.Run()
}

func (this *Distributer) Run() {
	log.Info("Start distributer:")
	for {
		if this.IsStopping() {
			this.Lock()
			defer this.Unlock()
			this.Status = constant.RUNNING_STATUS_STOPPED
			break
		}
		if this.IsStop() {
			break
		}
		if this.IsPause() {
			time.Sleep(1 * time.Second)
			continue
		}
		this.pool.Add()
		go func() {
			defer this.pool.Done()
			request, err := this.Node.PopRequest()
			if err != nil {
				log.Errorf("Distribute Error: %s", err)
				return
			}
			if constant.RunMode == "debug" {
				log.Infof("distribute request: %v thread: %d", request, this.pool.Num())
			}
			//ok, err := this.Node.HasRequest(request)
			//if err != nil {
			//	log.Errorf("Distribute Error: %s", err)
			//	return
			//}
			//if ok {
			//	log.Warnf("Distribute Warn: Request is reapted. (Url: %s)", request.HTTPReq().URL)
			//	return
			//}
			//this.RpcClient.SignRequest(request)
			this.Distribute(request)
			this.RpcClient.Distribute(request.NodeName(), request)
		}()
	}
	log.Info("Stop distributer.")
}

func (this *Distributer) Distribute(request *data.Request) {
	nodeList := this.Cluster.GetAllNode()
	if this.LastIndex >= len(nodeList) {
		this.LastIndex = 0
	}
	nodeName := nodeList[this.LastIndex].Name
	request.SetNodeName(nodeName)
	this.LastIndex++
}
