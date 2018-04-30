package cluster

import (
	"github.com/l-dandelion/yi-ants-go/core/node"
	"github.com/l-dandelion/yi-ants-go/lib/utils"
	"sync"
)

type ClusterInfo struct {
	Name      string
	NodeList  []*node.NodeInfo
	LocalNode *node.NodeInfo
}

type Cluster interface {
	DeleteDeadNode(nodeName string)
	AddNode(nodeInfo *node.NodeInfo)
	GetClusterInfo() *ClusterInfo
	GetAllNode() []*node.NodeInfo
}

type myCluster struct {
	lock      sync.RWMutex
	Name      string
	NodeList  []*node.NodeInfo
	LocalNode *node.NodeInfo
	settings  *utils.Settings
}

func New(settings *utils.Settings, mnode *node.NodeInfo) Cluster {
	return &myCluster{
		Name:      settings.Name,
		NodeList:  []*node.NodeInfo{mnode},
		LocalNode: mnode,
		settings:  settings,
	}
}

func (cluster *myCluster) GetClusterInfo() *ClusterInfo {
	cluster.lock.RLock()
	defer cluster.lock.RUnlock()
	return &ClusterInfo{
		Name:      cluster.Name,
		NodeList:  cluster.NodeList,
		LocalNode: cluster.LocalNode,
	}
}

func (cluster *myCluster) DeleteDeadNode(nodeName string) {
	cluster.lock.Lock()
	defer cluster.lock.Unlock()
	deletedIndex := -1
	for i, node := range cluster.NodeList {
		if node.Name == nodeName {
			deletedIndex = i
			break
		}
	}
	if deletedIndex > 0 {
		cluster.NodeList = append(cluster.NodeList[0:deletedIndex],
			cluster.NodeList[deletedIndex+1:]...)
	}
}

func (cluster *myCluster) AddNode(nodeInfo *node.NodeInfo) {
	cluster.lock.Lock()
	defer cluster.lock.Unlock()
	for _, node := range cluster.NodeList {
		if node.Name == nodeInfo.Name {
			return
		}
	}
	cluster.NodeList = append(cluster.NodeList, nodeInfo)
}

func (cluster *myCluster) GetAllNode() []*node.NodeInfo {
	return cluster.NodeList
}
