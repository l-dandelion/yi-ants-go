package crawler

import (
	"sync"

	"github.com/l-dandelion/yi-ants-go/core/module/data"
	"github.com/l-dandelion/yi-ants-go/core/spider"
	"github.com/l-dandelion/yi-ants-go/lib/constant"
	"github.com/l-dandelion/yi-ants-go/lib/library/buffer"
)

// crawler
type Crawler interface {
	AddSpider(sp spider.Spider) bool
	StartSpider(spiderName string) *constant.YiError
	FirstStartSpider(spiderName string) *constant.YiError
	StopSpider(spiderName string) *constant.YiError
	PauseSpider(spiderName string) *constant.YiError
	RecoverSpider(spiderName string) *constant.YiError
	GetSpiderStatus(spiderName string) (*spider.SpiderStatus, *constant.YiError)
	GetSpidersName() []string
	GetSpidersStatus() []*spider.SpiderStatus
	CanWeStopSpider(spiderName string) (bool, *constant.YiError)
	PopRequest() (*data.Request, *constant.YiError)
	AcceptRequest(*data.Request) *constant.YiError
}

type myCrawler struct {
	distributeQueue buffer.Pool
	spiderMapLock   sync.RWMutex
	SpiderMap       map[string]spider.Spider //contains all spiders
}

/*
 * create an instance of Crawler
 */
func NewCrawler() (Crawler, *constant.YiError) {
	pool, err := buffer.NewPool(50, 1000)
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
func (crawler *myCrawler) AddSpider(sp spider.Spider) bool {
	crawler.spiderMapLock.Lock()
	defer crawler.spiderMapLock.Unlock()
	if sp == nil {
		return false
	}
	if _, ok := crawler.SpiderMap[sp.SpiderName()]; ok {
		return false
	}
	crawler.SpiderMap[sp.SpiderName()] = sp
	return true
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
	return sp.NotFirstStart()
}

/*
 * first start a spider
 */
func (crawler *myCrawler) FirstStartSpider(spiderName string) *constant.YiError {
	sp, yierr := crawler.GetSpider(spiderName)
	if yierr != nil {
		return yierr
	}
	return sp.FirstStart()
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
func (Crawler *myCrawler) PopRequest() (*data.Request, *constant.YiError) {
	req, err := Crawler.distributeQueue.Get()
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