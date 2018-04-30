package crawler

import (
	"sync"

	"github.com/l-dandelion/yi-ants-go/core/module/data"
	"github.com/l-dandelion/yi-ants-go/core/spider"
	"github.com/l-dandelion/yi-ants-go/lib/constant"
	"github.com/l-dandelion/yi-ants-go/lib/library/buffer"
	"github.com/l-dandelion/yi-ants-go/lib/library/log"
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
	GetSpidersStatus() []*spider.SpiderStatus
	CanWeStopSpider(spiderName string) (bool, *constant.YiError)
	PopRequest() (*data.Request, *constant.YiError)
	AcceptRequest(*data.Request) *constant.YiError
	SignRequest(req *data.Request) *constant.YiError
	HasRequest(req *data.Request) (bool, *constant.YiError)
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
	yierr := sp.InitSchduler()
	if yierr != nil {
		return yierr
	}
	crawler.SpiderMap[sp.SpiderName()] = sp
	return nil
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
	if constant.RunMode == "debug" {
		log.Infof("Accept request: %v SpiderName: %s", req, req.SpiderName())
	}
	sp, yierr := crawler.GetSpider(req.SpiderName())
	if yierr != nil {
		return yierr
	}
	sp.AcceptedRequest(req)
	return nil
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
