package crawler

import (
	"sync"

	"github.com/l-dandelion/yi-ants-go/core/module/data"
	"github.com/l-dandelion/yi-ants-go/core/spider"
	"github.com/l-dandelion/yi-ants-go/lib/constant"
	"github.com/l-dandelion/yi-ants-go/lib/library/buffer"
	log "github.com/sirupsen/logrus"
)

// crawler
type Crawler interface {
	AddSpider(sp spider.Spider) *constant.YiError
	StartSpider(spiderName string) *constant.YiError
	FirstStartSpider(spiderName string) *constant.YiError
	StopSpider(spiderName string) *constant.YiError
	PauseSpider(spiderName string) *constant.YiError
	RecoverSpider(spiderName string) *constant.YiError
	GetSpiderStatus(spiderName string) (*spider.SpiderStatus, *constant.YiError)
	GetSpidersName() []string
	CanWeStopSpider(spiderName string) (bool, *constant.YiError)
	PopRequest() (*data.Request, *constant.YiError)
	AcceptRequest(*data.Request) *constant.YiError
	SignRequest(req *data.Request) *constant.YiError
	HasRequest(req *data.Request) (bool, *constant.YiError)
	GetSpiders() []spider.Spider
	InitSpider(spiderName string) *constant.YiError
	ComplileSpider(spiderName string) *constant.YiError
	DeleteSpider(spiderName string) *constant.YiError
	GetSpiderStatusList() []*spider.SpiderStatus
	GetDistributeQueueSize() uint64
	CanStartSpider(string) (bool, *constant.YiError)
	GetScore() uint64
	PopRequests(max uint64) ([]*data.Request, *constant.YiError)
	AcceptRequests(reqs []*data.Request)
	FilterRequests(reqs []*data.Request) []*data.Request
	SignRequests(reqs []*data.Request)
}

type myCrawler struct {
	distributeQueue     buffer.Pool
	distributeQueueLock sync.Mutex
	spiderMapLock       sync.RWMutex
	SpiderMap           map[string]spider.Spider //contains all spiders
}

/*
 * create an instance of Crawler
 */
func NewCrawler() (Crawler, *constant.YiError) {
	pool, err := buffer.NewPool(50, 20000)
	if err != nil {
		return nil, constant.NewYiErrore(constant.ERR_CRAWLER_NEW, err)
	}
	return &myCrawler{
		distributeQueue: pool,
		SpiderMap:       map[string]spider.Spider{},
	}, nil
}

/*
 * add a spider
 */
func (crawler *myCrawler) AddSpider(sp spider.Spider) *constant.YiError {
	crawler.spiderMapLock.Lock()
	defer crawler.spiderMapLock.Unlock()
	if sp == nil {
		return constant.NewYiErrorf(constant.ERR_ADD_SPIDER, "Nil spider.")
	}
	if _, ok := crawler.SpiderMap[sp.SpiderName()]; ok {
		return constant.NewYiErrorf(constant.ERR_ADD_SPIDER, "Exists spider name.")
	}
	crawler.SpiderMap[sp.SpiderName()] = sp
	return nil
}

/*
 * complile a spider
 */
func (crawler *myCrawler) ComplileSpider(spiderName string) *constant.YiError {
	sp, yierr := crawler.GetSpider(spiderName)
	if yierr != nil {
		return yierr
	}
	return sp.Complile()
}

/*
 * init a spider
 */
func (crawler *myCrawler) InitSpider(spiderName string) *constant.YiError {
	crawler.spiderMapLock.RLock()
	defer crawler.spiderMapLock.RUnlock()
	sp, ok := crawler.SpiderMap[spiderName]
	if !ok {
		return constant.NewYiErrorf(constant.ERR_SPIDER_NOT_FOUND,
			"Spider not found.(spiderName: %s)", spiderName)
	}
	yierr := sp.InitSchduler()
	return yierr
}

/*
 * start a spider
 */
func (crawler *myCrawler) StartSpider(spiderName string) *constant.YiError {
	crawler.spiderMapLock.RLock()
	defer crawler.spiderMapLock.RUnlock()
	sp, ok := crawler.SpiderMap[spiderName]
	if !ok {
		return constant.NewYiErrorf(constant.ERR_SPIDER_NOT_FOUND,
			"Spider not found.(spiderName: %s)", spiderName)
	}
	return sp.NotFirstStart(crawler.distributeQueue)
}

/*
 * first start a spider
 */
func (crawler *myCrawler) FirstStartSpider(spiderName string) *constant.YiError {
	sp, yierr := crawler.GetSpider(spiderName)
	if yierr != nil {
		return yierr
	}
	return sp.FirstStart(crawler.distributeQueue)
}

/*
 * stop a spider
 */
func (crawler *myCrawler) StopSpider(spiderName string) *constant.YiError {
	sp, yierr := crawler.GetSpider(spiderName)
	if yierr != nil {
		return yierr
	}
	return sp.Stop()
}

/*
 * pause a spider
 */
func (crawler *myCrawler) PauseSpider(spiderName string) *constant.YiError {
	sp, yierr := crawler.GetSpider(spiderName)
	if yierr != nil {
		return yierr
	}
	return sp.Pause()
}

/*
 * recover a spider
 */
func (crawler *myCrawler) RecoverSpider(spiderName string) *constant.YiError {
	sp, yierr := crawler.GetSpider(spiderName)
	if yierr != nil {
		return yierr
	}
	return sp.Recover()
}

/*
 * delete a spider
 */
func (crawler *myCrawler) DeleteSpider(spiderName string) *constant.YiError {
	sp, yierr := crawler.GetSpider(spiderName)
	if yierr != nil {
		return yierr
	}
	if sp.GetSched() != nil {
		go sp.Stop()
	}
	crawler.spiderMapLock.Lock()
	defer crawler.spiderMapLock.Unlock()
	delete(crawler.SpiderMap, spiderName)
	return nil
}

/*
 * accepted a request
 */
func (crawler *myCrawler) AcceptedRequest(req *data.Request) *constant.YiError {
	spiderName := req.SpiderName()
	sp, yierr := crawler.GetSpider(spiderName)
	if yierr != nil {
		return yierr
	}
	sp.AcceptedRequest(req)
	return nil
}

/*
 * get spider by spider name
 */
func (crawler *myCrawler) GetSpider(spiderName string) (spider.Spider, *constant.YiError) {
	crawler.spiderMapLock.RLock()
	defer crawler.spiderMapLock.RUnlock()
	sp, ok := crawler.SpiderMap[spiderName]
	if !ok {
		return nil, constant.NewYiErrorf(constant.ERR_SPIDER_NOT_FOUND,
			"Spider not found.(spiderName: %s)", spiderName)
	}
	return sp, nil
}

/*
 * get all spider name
 */
func (crawler *myCrawler) GetSpidersName() []string {
	names := []string{}
	crawler.spiderMapLock.RLock()
	defer crawler.spiderMapLock.RUnlock()
	for name, _ := range crawler.SpiderMap {
		names = append(names, name)
	}
	return names
}

/*
 * check whether whether we can stop the spider
 */
func (crawler *myCrawler) CanWeStopSpider(spiderName string) (bool, *constant.YiError) {
	sp, yierr := crawler.GetSpider(spiderName)
	if yierr != nil {
		return false, yierr
	}
	return sp.Idle(), nil
}

/*
 * get spider status
 */
func (crawler *myCrawler) GetSpiderStatus(spiderName string) (*spider.SpiderStatus, *constant.YiError) {
	sp, yierr := crawler.GetSpider(spiderName)
	if yierr != nil {
		return nil, yierr
	}
	return sp.SpiderStatus(), nil
}

/*
 * get all spider status
 */
func (crawler *myCrawler) GetSpidersStatus() []*spider.SpiderStatus {
	spiderStatusList := []*spider.SpiderStatus{}
	crawler.spiderMapLock.RLock()
	defer crawler.spiderMapLock.RUnlock()
	for _, spider := range crawler.SpiderMap {
		spiderStatusList = append(spiderStatusList, spider.SpiderStatus())
	}
	return spiderStatusList
}

/*
 * pop a request
 */
func (crawler *myCrawler) PopRequest() (*data.Request, *constant.YiError) {
	crawler.distributeQueueLock.Lock()
	defer crawler.distributeQueueLock.Unlock()
	req, err := crawler.distributeQueue.Get()
	if err != nil {
		return nil, constant.NewYiErrore(constant.ERR_REQUEST_POP, err)
	}
	return req.(*data.Request), nil
}

/*
 * accept a request
 */
func (crawler *myCrawler) AcceptRequest(req *data.Request) *constant.YiError {
	sp, yierr := crawler.GetSpider(req.SpiderName())
	if yierr != nil {
		return yierr
	}
	sp.AcceptedRequest(req)
	return nil
}

/*
 * accept requests
 */
func (crawler *myCrawler) AcceptRequests(reqs []*data.Request) {
	if reqs == nil {
		return
	}
	for _, req := range reqs {
		go func(req *data.Request) {
			yierr := crawler.AcceptRequest(req)
			if yierr != nil {
				log.Infof("Accept request Error: %s, Req: %v", yierr, req)
			}
		}(req)
	}
}

/*
 * sign a request
 */
func (crawler *myCrawler) SignRequest(req *data.Request) *constant.YiError {
	sp, yierr := crawler.GetSpider(req.SpiderName())
	if yierr != nil {
		return yierr
	}
	sp.SignRequest(req)
	return nil
}

/*
 * check whether it has this request
 */
func (crawler *myCrawler) HasRequest(req *data.Request) (bool, *constant.YiError) {
	sp, yierr := crawler.GetSpider(req.SpiderName())
	if yierr != nil {
		return false, yierr
	}
	return sp.HasRequest(req), nil
}

/*
 * get the status of a spider named spiderName
 */
func (crawler *myCrawler) SpiderStatus(spiderName string) (*spider.SpiderStatus, *constant.YiError) {
	sp, yierr := crawler.GetSpider(spiderName)
	if yierr != nil {
		return nil, yierr
	}
	return sp.SpiderStatus(), nil
}

/*
 * get spiders
 */
func (crawler *myCrawler) GetSpiders() []spider.Spider {
	crawler.spiderMapLock.RLock()
	defer crawler.spiderMapLock.RUnlock()
	sps := []spider.Spider{}
	for _, sp := range crawler.SpiderMap {
		sps = append(sps, sp)
	}
	return sps
}

/*
 * get spiders status
 */
func (crawler *myCrawler) GetSpiderStatusList() []*spider.SpiderStatus {
	crawler.spiderMapLock.RLock()
	defer crawler.spiderMapLock.RUnlock()
	ssl := []*spider.SpiderStatus{}
	for _, sp := range crawler.SpiderMap {
		ssl = append(ssl, sp.SpiderStatus())
	}
	return ssl
}

/*
 * get distributer queue size
 */
func (crawler *myCrawler) GetDistributeQueueSize() uint64 {
	return crawler.distributeQueue.Total()
}

/*
 * can start spider
 */
func (crawler *myCrawler) CanStartSpider(spiderName string) (bool, *constant.YiError) {
	sp, yierr := crawler.GetSpider(spiderName)
	if yierr != nil {
		return false, yierr
	}
	return sp.CanStart(), nil
}

/*
 * get score
 */
func (crawler *myCrawler) GetScore() uint64 {
	total := crawler.distributeQueue.Total()
	spiderStatusList := crawler.GetSpiderStatusList()
	for _, ss := range spiderStatusList {
		if ss.Status != constant.RUNNING_STATUS_STOPPED {
			total += uint64(ss.Waiting)
		}
	}
	return total
}

/*
 * pop requests
 */
func (crawler *myCrawler) PopRequests(max uint64) ([]*data.Request, *constant.YiError) {
	crawler.distributeQueueLock.Lock()
	defer crawler.distributeQueueLock.Unlock()
	total := crawler.distributeQueue.Total()
	if total > max {
		total = max
	}
	reqs := []*data.Request{}
	for i := uint64(0); i < total; i++ {
		req, err := crawler.distributeQueue.Get()
		if err != nil {
			return nil, constant.NewYiErrore(constant.ERR_REQUEST_POP, err)
		}
		reqs = append(reqs, req.(*data.Request))
	}
	return reqs, nil
}

/*
 * filter requests
 */
func (crawler *myCrawler) FilterRequests(reqs []*data.Request) []*data.Request {
	result := []*data.Request{}
	for _, req := range reqs {
		sp, yierr := crawler.GetSpider(req.SpiderName())
		if yierr != nil {
			log.Error("Filter Request: %s", yierr)
			continue
		}
		if !sp.HasRequest(req) {
			result = append(result, req)
		}
	}
	return result
}

/*
 * Sign requests
 */
func (crawler *myCrawler) SignRequests(reqs []*data.Request) {
	if len(reqs) == 0 {
		return
	}
	for _, req := range reqs {
		sp, yierr := crawler.GetSpider(req.SpiderName())
		if yierr != nil {
			log.Error("Sign Request: %s", yierr)
			continue
		}
		sp.SignRequest(req)
	}
}
