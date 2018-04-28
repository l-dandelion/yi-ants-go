package main

import (
	"net/rpc"
	"strconv"
	"github.com/l-dandelion/yi-ants-go/lib/library/log"
	"github.com/l-dandelion/yi-ants-go/core/action"
	"fmt"
	"encoding/gob"
	"github.com/l-dandelion/yi-ants-go/lib/constant"
)


func main() {
	constant.RunMode = "debug"
	ip := "127.0.0.1"
	port := 8200
	client, err := rpc.Dial("tcp", ip+":"+strconv.Itoa(port))
	if err != nil {
		log.Panic(err)
	}
	req := &action.RpcSpider{}
	sp, yierr := genSpider()
	if yierr != nil {
		log.Panic(yierr)
	}
	gob.Register(sp)

	req.Spider = sp
	resp := &action.RpcError{}
	err = client.Call("RpcServer.FirstAddSpider", req, resp)
	if err != nil {
		log.Panic(err)
	}
	fmt.Println(resp.Result)
}
