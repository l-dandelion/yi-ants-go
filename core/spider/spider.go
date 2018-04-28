package spider

import (
	"net/http"
	"time"

	"encoding/gob"
	"github.com/l-dandelion/yi-ants-go/core/module"
	"github.com/l-dandelion/yi-ants-go/core/module/data"
	"github.com/l-dandelion/yi-ants-go/core/module/local/analyzer"
	"github.com/l-dandelion/yi-ants-go/core/module/local/downloader"
	"github.com/l-dandelion/yi-ants-go/core/module/local/pipeline"
	"github.com/l-dandelion/yi-ants-go/core/scheduler"
	"github.com/l-dandelion/yi-ants-go/lib/constant"
	"github.com/l-dandelion/yi-ants-go/lib/library/buffer"
	"github.com/l-dandelion/yi-ants-go/lib/library/log"
	"github.com/l-dandelion/yi-ants-go/lib/library/plugin"
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
	NotFirstStart(distributeQueue buffer.Pool) *constant.YiError
	FirstStart(distributeQueue buffer.Pool) *constant.YiError
	AcceptedRequest(req *data.Request) bool
	SpiderStatus() *SpiderStatus
	GetInitReqs() []*data.Request
	InitSchduler() *constant.YiError
	SetSched(scheduler.Scheduler)
	GetSched() scheduler.Scheduler
	Copy() Spider
}

func (spider *mySpider) GetInitReqs() []*data.Request {
	return spider.InitialReqs
}

type mySpider struct {
	scheduler.Scheduler
	Name        string
	RequestArgs scheduler.RequestArgs
	DataArgs    scheduler.DataArgs
	//RespParsers     []module.ParseResponse
	//ItemProccessors []module.ProcessItem
	StrGenParsers    string
	StrGenProcessors string
	InitialReqs      []*data.Request
	StartTime        time.Time
	EndTime          time.Time
}

/*
 * create an instance of spider
 */
func New(name string,
	requestArgs scheduler.RequestArgs,
	dataArgs scheduler.DataArgs,
	initialUrls []string,
	initialReqs []*data.Request,
	strGenParsers string,
	strGenProcessors string) (Spider, *constant.YiError) {
	spider := &mySpider{
		Name:             name,
		StrGenParsers:    strGenParsers,
		StrGenProcessors: strGenProcessors,
		//RespParsers:     parsers,
		//ItemProccessors: processors,
		DataArgs:    dataArgs,
		RequestArgs: requestArgs,
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
			if constant.RunMode == "debug" {
				log.Infof("%v", req)
			}
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
	return spider.Name
}

/*
 * initialize scheduler
 */
func (spider *mySpider) InitSchduler() *constant.YiError {
	sched := scheduler.New(spider.Name)
	spider.Scheduler = sched
	downloader, yierr := downloader.New("D1", genHTTPClient(), module.CalculateScoreSimple)
	if yierr != nil {
		return yierr
	}
	f, err := plugin.GenFuncFromStr(spider.StrGenParsers, "GenParsers")
	if err != nil {
		return constant.NewYiErrore(constant.ERR_SPIDER_NEW, err)
	}
	parsers := f.(func() []module.ParseResponse)()
	analyzer, yierr := analyzer.New("A1", parsers, module.CalculateScoreSimple)
	if yierr != nil {
		return yierr
	}
	f, err = plugin.GenFuncFromStr(spider.StrGenProcessors, "GenProcessors")
	if err != nil {
		return constant.NewYiErrore(constant.ERR_SPIDER_NEW, err)
	}
	processors := f.(func() []module.ProcessItem)()
	pipeline, yierr := pipeline.New("P1", processors, module.CalculateScoreSimple)
	moduleArgs := scheduler.ModuleArgs{
		Downloader: downloader,
		Analyzer:   analyzer,
		Pipeline:   pipeline,
	}
	yierr = sched.Init(spider.RequestArgs, spider.DataArgs, moduleArgs)
	if yierr != nil {
		return yierr
	}
	return nil
}

func (spider *mySpider) InitDistributeQueue(distributerQueue buffer.Pool) {
	spider.Scheduler.SetDistributeQueue(distributerQueue)
}

/*
 * start a spider
 */
func (spider *mySpider) NotFirstStart(distributeQueue buffer.Pool) *constant.YiError {
	spider.InitDistributeQueue(distributeQueue)
	yierr := spider.Scheduler.Start(nil)
	if yierr == nil {
		spider.StartTime = time.Now()
	}
	return yierr
}

/*
 * first start a spider
 */
func (spider *mySpider) FirstStart(distributeQueue buffer.Pool) *constant.YiError {
	spider.InitDistributeQueue(distributeQueue)
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
		Name:      spider.Name,
		Status:    spider.Status(),
		Crawled:   int(summary.Downloader.Called),
		Success:   int(summary.Downloader.Completed),
		Running:   int(summary.Downloader.Handling),
		Waiting:   int(summary.ReqBufferPool.Total),
		StartTime: spider.StartTime,
		EndTime:   spider.EndTime,
	}
}

/*
 * get scheduler
 */
func (spider *mySpider) GetSched() scheduler.Scheduler {
	return spider.Scheduler
}

/*
 * set scheduler
 */
func (spider *mySpider) SetSched(sched scheduler.Scheduler) {
	spider.Scheduler = sched
}

func init() {
	gob.Register(&mySpider{})
}

func (spider *mySpider) Copy() Spider {
	return &mySpider{
		Name:        spider.Name,
		RequestArgs: spider.RequestArgs,
		DataArgs:    spider.DataArgs,
		//RespParsers     []module.ParseResponse
		//ItemProccessors []module.ProcessItem
		StrGenParsers:    spider.StrGenParsers,
		StrGenProcessors: spider.StrGenProcessors,
		InitialReqs:      spider.InitialReqs,
		StartTime:        spider.StartTime,
		EndTime:          spider.EndTime,
	}
}

//func (spider *mySpider) SignRequest(req *data.Request) *constant.YiError {
//	if spider.Status() == constant.RUNNING_STATUS_UNPREPARED {
//		return constant.NewYiErrorf(constant.ERR_SCHEDULER_NOT_INITILATED, "Scheduler has not been initilated.")
//	}
//	spider.Scheduler.SignRequest(req)
//	return nil
//}
