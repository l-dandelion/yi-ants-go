package main

import (
	"github.com/l-dandelion/yi-ants-go/core/node"
	"github.com/l-dandelion/yi-ants-go/lib/utils"
	"fmt"
	"github.com/l-dandelion/yi-ants-go/core/cluster"
	"github.com/l-dandelion/yi-ants-go/core/action/rpc"
	"github.com/l-dandelion/yi-ants-go/core/action/watcher"
	"time"
	"strconv"
	"net/http"
	http2 "github.com/l-dandelion/yi-ants-go/core/action/http"
	"github.com/l-dandelion/yi-ants-go/lib/library/log"
	"github.com/l-dandelion/yi-ants-go/lib/constant"
)

func solve(isFirst bool, port int, httpPort int) {
	settings := &utils.Settings{
		Name: "test",
		TcpPort: port,
		HttpPort: httpPort,
	}
	mnode, yierr := node.New(settings)
	if yierr != nil {
		fmt.Println(yierr)
	}
	mcluster := cluster.New(settings, mnode.GetNodeInfo())
	rpcClient := rpc.NewRpcClient(mnode, mcluster)
	distributer := watcher.NewDistributer(mnode, mcluster, rpcClient)
	rpcClient.Start()
	rpc.NewRpcServer(mnode, mcluster, port, rpcClient, distributer)
	distributer.Start()

	router := http2.NewRouter(mnode, mcluster, nil, distributer, rpcClient)

	httpPortStr := strconv.Itoa(httpPort)
	httpServer := http.Server{
		Addr:    ":" + httpPortStr,
		Handler: router,
	}
	go func() {
		err := httpServer.ListenAndServe()
		if err != nil {
			log.Panic(err)
		}
	}()

	if !isFirst {
		yierr := rpcClient.LetMeIn("127.0.0.1", 8200)
		if yierr != nil {
			fmt.Printf("%#v", yierr)
		}
		time.Sleep(5*time.Second)
		nodes := mcluster.GetAllNode()
		fmt.Println(nodes)
	}

}

func main() {
	constant.RunMode = "debug"
	solve(false, 8300, 9300)
	time.Sleep(100000*time.Second)
}