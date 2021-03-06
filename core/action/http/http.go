package http

import (
	"encoding/json"
	"github.com/l-dandelion/yi-ants-go/core/action"
	"github.com/l-dandelion/yi-ants-go/core/cluster"
	"github.com/l-dandelion/yi-ants-go/core/node"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
	"github.com/l-dandelion/yi-ants-go/lib/constant"
)

type Result struct {
	Yierr *constant.YiError
	Content interface{}
}

type Router struct {
	node        node.Node
	cluster     cluster.Cluster
	mux         map[string]func(http.ResponseWriter, *http.Request)
	reporter    action.Watcher
	distributer action.Watcher
	rpcClient   action.RpcClientAnts
}

func NewRouter(node node.Node, cluster cluster.Cluster, reporter, distributer action.Watcher, rpcClient action.RpcClientAnts) *Router {
	mux := make(map[string]func(http.ResponseWriter, *http.Request))
	router := &Router{
		node:        node,
		cluster:     cluster,
		mux:         mux,
		reporter:    reporter,
		distributer: distributer,
		rpcClient:   rpcClient,
	}
	mux["/"] = router.Welcome
	mux["/cluster"] = router.Cluster
	mux["/spiders"] = router.Spiders
	mux["/crawl"] = router.Crawl
	mux["/spiderstatus"] = router.SpiderStatus
	mux["/stopspider"] = router.StopSpider
	mux["/pausespider"] = router.PauseSpider
	mux["/recoverspider"] = router.RecoverSpider
	mux["/startspider"] = router.Crawl
	return router
}

func (this *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	url := r.URL.String()
	log.Info("Get request:" + url)
	path := r.URL.Path
	if h, ok := this.mux[path]; ok {
		w.Header().Set("Content-Type", "application/json")
		h(w, r)
		return
	}
	this.Welcome(w, r)
}

func (this *Router) Welcome(w http.ResponseWriter, r *http.Request) {
	now := time.Now().Format("2006-01-02 15:04:05")
	welcome := WelcomeInfo{
		"for crawl",
		"do not panic",
		now,
	}
	encoder, err := json.Marshal(welcome)
	if err != nil {
		log.Error(err)
	}
	w.Write(encoder)
}
func (this *Router) Spiders(w http.ResponseWriter, r *http.Request) {
	spiderList := this.node.GetSpidersName()
	encoder, err := json.Marshal(spiderList)
	if err != nil {
		log.Error(err)
	}
	w.Write(encoder)
}

// try to start spider
// if ok
// *		tell other node start spider
// *		start reporter and distribute in this node
func (this *Router) Crawl(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	spiderName := r.Form["spider"][0]
	now := time.Now().Format("2006-01-02 15:04:05")
	startResult := &StartSpiderResult{}
	startResult.Time = now
	startResult.Spider = spiderName
	yierr := this.rpcClient.StartSpider(spiderName)
	if yierr == nil {
		startResult.Success = true
	}
	startResult.Yierr = yierr
	encoder, err := json.Marshal(startResult)
	if err != nil {
		log.Error(err)
	}
	w.Write(encoder)
}

func (this *Router) Cluster(w http.ResponseWriter, r *http.Request) {
	encoder, err := json.Marshal(this.cluster.GetClusterInfo())
	if err != nil {
		log.Error(err)
	}
	w.Write(encoder)
}

func (this *Router) SpiderStatus(w http.ResponseWriter, r *http.Request){
	r.ParseForm()
	spiderName := r.Form["spider"][0]
	spiderStatus, yierr := this.node.GetSpiderStatus(spiderName)
	result := &Result{
		Yierr: yierr,
		Content: spiderStatus,
	}
	encoder, err := json.Marshal(result)
	if err != nil {
		log.Error(err)
	}
	w.Write(encoder)
}

func (this *Router) StopSpider(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	spiderName := r.Form["spider"][0]
	now := time.Now().Format("2006-01-02 15:04:05")
	startResult := &StartSpiderResult{}
	startResult.Time = now
	startResult.Spider = spiderName
	yierr := this.rpcClient.StopSpider(spiderName)
	if yierr == nil {
		startResult.Success = true
	}
	startResult.Yierr = yierr
	encoder, err := json.Marshal(startResult)
	if err != nil {
		log.Error(err)
	}
	w.Write(encoder)
}

func (this *Router) PauseSpider(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	spiderName := r.Form["spider"][0]
	now := time.Now().Format("2006-01-02 15:04:05")
	startResult := &StartSpiderResult{}
	startResult.Time = now
	startResult.Spider = spiderName
	yierr := this.rpcClient.PauseSpider(spiderName)
	if yierr == nil {
		startResult.Success = true
	}
	startResult.Yierr = yierr
	encoder, err := json.Marshal(startResult)
	if err != nil {
		log.Error(err)
	}
	w.Write(encoder)
}

func (this *Router) RecoverSpider(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	spiderName := r.Form["spider"][0]
	now := time.Now().Format("2006-01-02 15:04:05")
	startResult := &StartSpiderResult{}
	startResult.Time = now
	startResult.Spider = spiderName
	yierr := this.rpcClient.RecoverSpider(spiderName)
	if yierr == nil {
		startResult.Success = true
	}
	startResult.Yierr = yierr
	encoder, err := json.Marshal(startResult)
	if err != nil {
		log.Error(err)
	}
	w.Write(encoder)
}