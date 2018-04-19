package main

import (
	"time"
	"net/http"
	"fmt"

	"github.com/l-dandelion/yi-ants-go/core/module"
	"github.com/l-dandelion/yi-ants-go/core/module/local/analyzer"
	"github.com/l-dandelion/yi-ants-go/core/module/local/downloader"
	"github.com/l-dandelion/yi-ants-go/core/module/local/pipeline"
	"github.com/l-dandelion/yi-ants-go/core/scheduler"
	"github.com/l-dandelion/yi-ants-go/lib/library/log"
	"github.com/l-dandelion/yi-ants-go/core/module/data"
	"sync/atomic"
)

var (
	sched       scheduler.Scheduler
	snGenerator = module.NewSNGenerator(1, 0)
	errNum uint32 = 0
)

func Monite() {
	// 观察错误。
	go func() {
		errChan := sched.ErrorChan()
		for {
			yierr, ok := <-errChan
			if !ok {
				break
			}
			atomic.AddUint32(&errNum, 1)
			log.Errorf("An error occurs when running scheduler: %s", yierr)
		}
	}()
	//打印有变化的摘要信息。
	go func() {
		var count int
		var prevSummary string
		for {
			summary, yierr := sched.Summary().String()
			if yierr != nil {
				log.Error(yierr)
			}

			log.Infof("ErrNum: %d", atomic.LoadUint32(&errNum))
			if prevSummary == "" || summary != prevSummary {
				log.Infof("-- Summary[%d]:\n%s",
					count, summary)
				prevSummary = summary
				count++
			}
			time.Sleep(time.Millisecond * 500)
		}
	}()
	//若连续5秒空闲，则停止调度器。
	var count int
	max := 5
	tickCh := time.Tick(time.Second)
	for _ = range tickCh {
		if sched.Idle() {
			count++
			log.Infof("Increase idle count, and value is %d.", count)
		}
		if count >= max {
			log.Infof("The idle count is equal or greater than %d.", max)
			break
		}
	}
	if yierr := sched.Stop(); yierr != nil {
		log.Fatalf("An error occurs when stopping scheduler: %s", yierr)
	}
}

func main() {
	sched = scheduler.NewScheduler()
	requestArgs := scheduler.RequestArgs{
		MaxDepth:        3,
		AcceptedDomains: []string{"pixabay.com"},
	}
	dataArgs := scheduler.DataArgs{
		ReqBufferCap:         50,
		ReqMaxBufferNumber:   1000,
		RespBufferCap:        50,
		RespMaxBufferNumber:  100,
		ItemBufferCap:        50,
		ItemMaxBufferNumber:  1000,
		ErrorBufferCap:       50,
		ErrorMaxBufferNumber: 1,
	}
	mid := fmt.Sprintf("D%d", snGenerator.Get())
	client := genHTTPClient()
	downloader, yierr := downloader.New(module.MID(mid), client, module.CalculateScoreSimple)
	if yierr != nil {
		log.Fatalf("An error occurs when new downloader: %s", yierr)
	}
	mid = fmt.Sprintf("A%d", snGenerator.Get())
	analyzer, yierr := analyzer.New(module.MID(mid), []module.ParseResponse{parseImgTag, parseATag2}, module.CalculateScoreSimple)
	if yierr != nil {
		log.Fatalf("An error occurs when new analyzer: %s", yierr)
	}
	mid = fmt.Sprintf("P%d", snGenerator.Get())
	pipeline, yierr := pipeline.New(module.MID(mid), []module.ProcessItem{process}, module.CalculateScoreSimple)
	moduleArgs := scheduler.ModuleArgs{
		Downloader: downloader,
		Analyzer: analyzer,
		Pipeline: pipeline,
	}
	yierr = sched.Init(requestArgs, dataArgs, moduleArgs)
	if yierr != nil {
		log.Fatalf("An error occurs when initializing scheduler: %s", yierr)
	}
	httpReq, err := http.NewRequest("GET", "https://pixabay.com", nil)
	if err != nil {
		log.Fatalf("An error occurs when new HTTP request: %s", err)
	}
	req := data.NewRequest(httpReq)
	req.SetUserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/65.0.3325.181 Safari/537.36")
	fmt.Printf("%#v", req.HTTPReq().Host)
	initialRequest := []*data.Request{req}
	yierr = sched.Start(initialRequest)
	if yierr != nil {
		log.Fatalf("An error occurs when starting scheduler: %s", yierr)
	}
	Monite()
}
