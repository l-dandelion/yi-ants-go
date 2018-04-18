package scheduler

import (
	"github.com/l-dandelion/yi-ants-go/core/module/data"
	"github.com/l-dandelion/yi-ants-go/lib/constant"
	"github.com/l-dandelion/yi-ants-go/lib/library/log"
	"time"
)

/*
 * start analyze
 */
func (sched *myScheduler) analyze() {
	go func() {
		for {
			//stopped
			if sched.canceled() {
				break
			}
			//paused
			if sched.Status() == constant.RUNNING_STATUS_PAUSED {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			datum, err := sched.respBufferPool.Get()
			if err != nil {
				log.Warnln("The response buffer pool was closed. Break response reception.")
				break
			}
			analyzerPool.Add()
			go func(datum interface{}) {
				defer analyzerPool.Done()
				resp, ok := datum.(*data.Response)
				if !ok {
					yierr := constant.NewYiErrorf(constant.ERR_CRAWL_SCHEDULER,
						"Incorrect response type: %T", datum)
					sched.sendError(yierr)
					return
				}
				sched.analyzeOne(resp)
			}(datum)
		}
		analyzerPool.Wait()
	}()
}

/*
 * analyze one
 */
func (sched *myScheduler) analyzeOne(resp *data.Response) {
	if resp == nil {
		return
	}
	if sched.canceled() {
		return
	}
	analyzer := sched.analyzer
	dataList, yierrs := analyzer.Analyze(resp)
	if dataList != nil {
		for _, mdata := range dataList {
			if mdata == nil {
				continue
			}
			switch d := mdata.(type) {
			case *data.Request:
				d.SetDepth(resp.Depth() + 1)
				sched.sendReq(d)
			case data.Item:
				sched.sendItem(d)
			default:
				yierr := constant.NewYiErrorf(constant.ERR_CRAWL_ANALYZER,
					"Unsupported data type: %T (data: %#v)", mdata, mdata)
				sched.sendError(yierr)
			}
		}
	}
	if yierrs != nil {
		for _, yierr := range yierrs {
			sched.sendError(yierr)
		}
	}
}
