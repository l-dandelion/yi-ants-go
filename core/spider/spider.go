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
	log "github.com/sirupsen/logrus"
	"github.com/l-dandelion/yi-ants-go/lib/library/plugin"
	"strings"
	"sync"
)

type SpiderStatus struct {
	Name            string
	CompilingStatus int8
	Status          int8
	Crawled         int
	Success         int
	Running         int
	Waiting         int
	StartTime       time.Time
	EndTime         time.Time
	ComplilingError *constant.YiError
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
	Complile() *constant.YiError
	CanStart() bool
}

func (spider *mySpider) GetInitReqs() []*data.Request {
	return spider.InitialReqs
}

type mySpider struct {
	scheduler.Scheduler
	Name                string
	RequestArgs         scheduler.RequestArgs
	DataArgs            scheduler.DataArgs
	respParsers         []module.ParseResponse
	itemProccessors     []module.ProcessItem
	StrGenParsers       string
	StrGenProcessors    string
	InitialReqs         []*data.Request
	StartTime           time.Time
	EndTime             time.Time
	compilingError      *constant.YiError
	compilingStatus     int8
	compilingStatusLock sync.Mutex
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
			if strings.Index(urlStr, "http") != 0 {
				urlStr = "http://" + urlStr
			}
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

func (spider *mySpider) Complile() (yierr *constant.YiError) {
	spider.compilingStatusLock.Lock()
	if spider.compilingStatus != constant.COMPLILING_STATUS_UNCOMPLILED {
		yierr = constant.NewYiErrorf(constant.ERR_COMPLILE_FAIL, "The compliling is not uncompliled.(Status: %d)", spider.compilingStatus)
		spider.compilingStatusLock.Unlock()
		return
	}
	spider.compilingStatus = constant.COMPLILING_STATUS_COMPLILING
	spider.compilingStatusLock.Unlock()
	defer func() {
		spider.compilingStatusLock.Lock()
		if yierr != nil {
			spider.compilingStatus = constant.COMPLILING_STATUS_FAIL
			spider.compilingError = yierr
		} else {
			spider.compilingStatus = constant.COMPLILING_STATUS_COMPLILED
		}
		spider.compilingStatusLock.Unlock()
	}()
	f, err := plugin.GenFuncFromStr(spider.StrGenParsers, "GenParsers")
	if err != nil {
		yierr = constant.NewYiErrorf(constant.ERR_FUNC_GEN, "Gen parsers fail.Err: %s", err)
		return
	}
	spider.respParsers = f.(func() []module.ParseResponse)()
	f, err = plugin.GenFuncFromStr(spider.StrGenProcessors, "GenProcessors")
	if err != nil {
		yierr = constant.NewYiErrorf(constant.ERR_FUNC_GEN, "Gen processors fail.Err: %s", err)
		return
	}
	spider.itemProccessors = f.(func() []module.ProcessItem)()
	return
}

/*
 * initialize scheduler
 */
func (spider *mySpider) InitSchduler() (yierr *constant.YiError) {
	spider.compilingStatusLock.Lock()
	if spider.compilingStatus != constant.COMPLILING_STATUS_COMPLILED {
		defer spider.compilingStatusLock.Unlock()
		return constant.NewYiErrorf(constant.ERR_NOT_COMPLILED, "Spider is not complized.(Status: %d)", spider.compilingStatus)
	}
	spider.compilingStatusLock.Unlock()

	sched := scheduler.New(spider.Name)
	spider.Scheduler = sched
	downloader, yierr := downloader.New("D1", genHTTPClient(), module.CalculateScoreSimple)
	if yierr != nil {
		return yierr
	}
	parsers := spider.respParsers
	analyzer, yierr := analyzer.New("A1", parsers, module.CalculateScoreSimple)
	if yierr != nil {
		return yierr
	}
	processors := spider.itemProccessors
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
	spider.compilingStatusLock.Lock()
	if spider.compilingStatus != constant.COMPLILING_STATUS_COMPLILED {
		defer spider.compilingStatusLock.Unlock()
		return constant.NewYiErrorf(constant.ERR_NOT_COMPLILED, "Spider is not complized.(Status: %d)", spider.compilingStatus)
	}
	spider.compilingStatusLock.Unlock()

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
	spider.compilingStatusLock.Lock()
	if spider.compilingStatus != constant.COMPLILING_STATUS_COMPLILED {
		defer spider.compilingStatusLock.Unlock()
		return constant.NewYiErrorf(constant.ERR_NOT_COMPLILED, "Spider is not complized.(Status: %d)", spider.compilingStatus)
	}
	spider.compilingStatusLock.Unlock()
	if spider.Scheduler == nil {
		return constant.NewYiErrorf(constant.ERR_SCHEDULER_NOT_INITILATED, "Spider is not initilated.")
	}

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
	spider.compilingStatusLock.Lock()
	if spider.compilingStatus != constant.COMPLILING_STATUS_COMPLILED {
		defer spider.compilingStatusLock.Unlock()
		return constant.NewYiErrorf(constant.ERR_NOT_COMPLILED, "Spider is not complized.(Status: %d)", spider.compilingStatus)
	}
	spider.compilingStatusLock.Unlock()

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
	spider.compilingStatusLock.Lock()
	defer spider.compilingStatusLock.Unlock()

	if spider.Scheduler == nil || spider.Summary() == nil {
		return &SpiderStatus{
			Name:            spider.Name,
			CompilingStatus: spider.compilingStatus,
			ComplilingError: spider.compilingError,
			Status:          0,
			Crawled:         0,
			Success:         0,
			Running:         0,
			Waiting:         0,
			StartTime:       spider.StartTime,
			EndTime:         spider.EndTime,
		}
	}

	summary := spider.Summary().Struct()
	return &SpiderStatus{
		Name:            spider.Name,
		CompilingStatus: spider.compilingStatus,
		ComplilingError: spider.compilingError,
		Status:          spider.Status(),
		Crawled:         int(summary.Downloader.Called),
		Success:         int(summary.Downloader.Completed),
		Running:         int(summary.Downloader.Handling),
		Waiting:         int(summary.ReqBufferPool.Total),
		StartTime:       spider.StartTime,
		EndTime:         spider.EndTime,
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

// can start
func (spider *mySpider) CanStart() bool {
	spider.compilingStatusLock.Lock()
	defer spider.compilingStatusLock.Unlock()
	return spider.compilingStatus == constant.COMPLILING_STATUS_COMPLILED
}