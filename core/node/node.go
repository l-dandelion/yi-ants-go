package node

import (
	"github.com/l-dandelion/yi-ants-go/lib/utils"
	"github.com/l-dandelion/yi-ants-go/core/crawler"
	"strconv"
	"github.com/l-dandelion/yi-ants-go/lib/constant"
)

type NodeInfo struct {
	Name     string
	Ip       string
	Port     int
	Settings *utils.Settings
}

type Node interface {
	crawler.Crawler
	GetNodeInfo() *NodeInfo
	IsMe(nodeName string) bool
}

type myNode struct {
	crawler.Crawler
	NodeInfo *NodeInfo
}

/*
 * create an instance of Node
 */
func New(settings *utils.Settings) (Node, *constant.YiError){
	ip := utils.GetLocalIp()
	name := ip + ":" + strconv.Itoa(settings.TcpPort)
	nodeInfo := &NodeInfo{name, ip, settings.TcpPort, settings}
	crawler, yierr := crawler.NewCrawler()
	if yierr != nil {
		return nil, yierr
	}
	return &myNode{
		NodeInfo: nodeInfo,
		Crawler: crawler,
	}, nil
}

/*
 * get node info
 */
func (node *myNode) GetNodeInfo() *NodeInfo {
	return node.NodeInfo
}

/*
 * check whether it is myself
 */
func (node *myNode) IsMe(nodeName string) bool {
	return node.NodeInfo.Name == nodeName
}

