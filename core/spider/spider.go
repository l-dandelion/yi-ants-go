package spider

import (
	"net/http"
	"time"

	"github.com/l-dandelion/yi-ants-go/core/module"
	"github.com/l-dandelion/yi-ants-go/core/module/data"
	"github.com/l-dandelion/yi-ants-go/core/module/local/analyzer"
	"github.com/l-dandelion/yi-ants-go/core/module/local/downloader"
	"github.com/l-dandelion/yi-ants-go/core/module/local/pipeline"
	"github.com/l-dandelion/yi-ants-go/core/scheduler"
	"github.com/l-dandelion/yi-ants-go/lib/constant"
)

type SpiderStatus struct {
	Name      string
	Status    int8
	Crawled   int
	Success   int
	Running   int
	Waiting   int
	StartTime time.Time
	EndTime   time.Time
}

type Spider interface {
	scheduler.Scheduler
	SpiderName() string
	NotFirstStart() *constant.YiError
	FirstStart() *constant.YiError
	AcceptedRequest(req *data.Request) bool
	SpiderStatus() *SpiderStatus
}

type mySpider struct {
	scheduler.Scheduler
	name            string
	RequestArgs     scheduler.RequestArgs
	DataArgs        scheduler.DataArgs
	RespParsers     []module.ParseResponse
	ItemProccessors []module.ProcessItem
	InitialReqs     []*data.Request
	StartTime       time.Time
	EndTime         time.Time
}

/*
 * create an instance of spider
 */
func New(name string,
	requestArgs scheduler.RequestArgs,
	dataArgs scheduler.DataArgs,
	initialUrls []string,
	initialReqs []*data.Request,
	parsers []module.ParseResponse,
	processors []module.ProcessItem) (Spider, *constant.YiError) {
	spider := &mySpider{
		name:            name,
		RespParsers:     parsers,
		ItemProccessors: processors,
		DataArgs:        dataArgs,
		RequestArgs:     requestArgs,
	}
	//yierr := spider.initSchduler()
	//if yierr != nil {
	//	return nil, yierr
	//}
	spider.InitialReqs = []*data.Request{}
	if initialUrls != nil {
		for _, urlStr := range initialUrls {
			httpReq, err := http.NewRequest("GET", urlStr, nil)
			if err != nil {
				return nil, constant.NewYiErrore(constant.ERR_SPIDER_NEW, err)
			}
			req := data.NewRequest(httpReq)
			spider.InitialReqs = append(spider.InitialReqs, req)
		}
	}
	if initialReqs != nil {
		for _, req := range initialReqs {
			spider.InitialReqs = append(spider.InitialReqs, req)
		}
	}
	return spider, nil
}

func (spider *mySpider) SpiderName() string {
	return spider.name
}

/*
 * initialize scheduler
 */
func (spider *mySpider) initSchduler() *constant.YiError {
	sched := scheduler.New(spider.name)
	downloader, yierr := downloader.New("D1", genHTTPClient(), module.CalculateScoreSimple)
	if yierr != nil {
		return yierr
	}
	analyzer, yierr := analyzer.New("A1", spider.RespParsers, module.CalculateScoreSimple)
	if yierr != nil {
		return yierr
	}
	pipeline, yierr := pipeline.New("P1", spider.ItemProccessors, module.CalculateScoreSimple)
	moduleArgs := scheduler.ModuleArgs{
		Downloader: downloader,
		Analyzer:   analyzer,
		Pipeline:   pipeline,
	}
	yierr = sched.Init(spider.RequestArgs, spider.DataArgs, moduleArgs)
	if yierr != nil {
		return yierr
	}
	spider.Scheduler = sched
	return nil
}

/*
 * start a spider
 */
func (spider *mySpider) NotFirstStart() *constant.YiError {
	yierr := spider.Scheduler.Start(nil)
	if yierr == nil {
		spider.StartTime = time.Now()
	}
	return yierr
}

/*
 * first start a spider
 */
func (spider *mySpider) FirstStart() *constant.YiError {
	yierr := spider.Scheduler.Start(spider.InitialReqs)
	if yierr == nil {
		spider.StartTime = time.Now()
	}
	return yierr
}

/*
 * accepted a request
 */
func (spider *mySpider) AcceptedRequest(req *data.Request) bool {
	return spider.SendReq(req)
}

/*
 * stop a spider
 */
func (spider *mySpider) Stop() *constant.YiError {
	yierr := spider.Scheduler.Stop()
	if yierr == nil {
		spider.EndTime = time.Now()
	}
	return yierr
}

/*
 * get spider status
 */
func (spider *mySpider) SpiderStatus() *SpiderStatus {
	summary := spider.Summary().Struct()
	return &SpiderStatus{
		Name: spider.name,
		Status: spider.Status(),
		Crawled: int(summary.Downloader.Called),
		Success: int(summary.Downloader.Completed),
		Running: int(summary.Downloader.Handling),
		Waiting: int(summary.ReqBufferPool.Total),
		StartTime: spider.StartTime,
		EndTime: spider.EndTime,
	}
}