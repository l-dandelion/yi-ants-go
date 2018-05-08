package watcher

import (
	"github.com/l-dandelion/yi-ants-go/core/action"
	"github.com/l-dandelion/yi-ants-go/core/cluster"
	"github.com/l-dandelion/yi-ants-go/core/module/data"
	"github.com/l-dandelion/yi-ants-go/core/node"
	"github.com/l-dandelion/yi-ants-go/lib/constant"
	"github.com/l-dandelion/yi-ants-go/lib/library/pool"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

var (
	InternalTime     = 3 * time.Second
	MaxDistributeNum = uint64(100)
)

type Distributer struct {
	sync.RWMutex
	Status         int8
	LastIndex      int
	Cluster        cluster.Cluster
	RpcClient      action.RpcClientAnts
	Node           node.Node
	MaxThread      int
	pool           *pool.Pool
	distributeLock sync.RWMutex
	scoreMap       map[string]uint64
	scoreMapLock   sync.Mutex
}

func (this *Distributer) UpdateScore() {
	nodeInfos := this.Cluster.GetAllNode()
	for _, ni := range nodeInfos {
		score, yierr := this.RpcClient.GetNodeScore(ni.Name)
		if yierr != nil {
			log.Infof("Update distribute score error: %s", yierr)
			continue
		}
		log.Infof("NodeName: %s Score: %d", ni.Name, score)
		this.scoreMapLock.Lock()
		this.scoreMap[ni.Name] = score
		this.scoreMapLock.Unlock()
	}
}

func (this *Distributer) UpdateScoreRun() {
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
		this.UpdateScore()
		time.Sleep(InternalTime)
	}
}

func (this *Distributer) GetBestIndex() int {
	this.scoreMapLock.Lock()
	defer this.scoreMapLock.Unlock()
	index := 0
	score := this.scoreMap[this.Node.GetNodeInfo().Name]
	nodeInfos := this.Cluster.GetAllNode()
	for i, nodeInfo := range nodeInfos {
		tmp := this.scoreMap[nodeInfo.Name]
		if index == 0 && tmp < score/2 || index != 0 && tmp < score {
			index = i
			score = tmp
		}
	}
	log.Infof("GetBestIndex: %d", index)
	return index
}

func NewDistributer(mnode node.Node, cluster cluster.Cluster, rpcClient action.RpcClientAnts) *Distributer {
	return &Distributer{
		Status:    constant.RUNNING_STATUS_STOPPED,
		Cluster:   cluster,
		RpcClient: rpcClient,
		Node:      mnode,
		MaxThread: 10,
		pool:      pool.NewPool(10),
		scoreMap:  make(map[string]uint64),
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
	go this.UpdateScoreRun()
	go this.DistributeRun()
}

func (this *Distributer) DistributeRun() {
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
		requests, err := this.Node.PopRequests(MaxDistributeNum)
		if err != nil {
			log.Errorf("Distribute Error: %s", err)
			continue
		}
		this.pool.Add()
		go func(requests []*data.Request) {
			defer this.pool.Done()
			nodeName := this.Distribute()
			this.scoreMapLock.Lock()
			this.scoreMap[nodeName] += uint64(len(requests))
			this.scoreMapLock.Unlock()
			this.RpcClient.SignRequests(requests)
			if(len(requests) != 0) {
				log.Infof("Sign request success. Num: %d", len(requests))
			}
			yierr := this.RpcClient.DistributeRequests(nodeName, requests)
			if yierr != nil {
				log.Error(yierr)
				return
			}
			if len(requests) != 0 {
				log.Infof("Distribute request success. Num: %d", len(requests))
			}
		}(requests)

		if uint64(len(requests)) < MaxDistributeNum {
			time.Sleep(InternalTime)
		}
	}
}

func (this *Distributer) Distribute() string {
	index := this.GetBestIndex()
	nodeList := this.Cluster.GetAllNode()
	if index > len(nodeList) {
		index = 0
	}
	nodeName := nodeList[index].Name
	return nodeName
}

func (this *Distributer) FilterReqests(reqs []*data.Request) []*data.Request {
	return this.RpcClient.FilterRequests(reqs)
}