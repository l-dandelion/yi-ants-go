package scheduler

import (
	"github.com/l-dandelion/yi-ants-go/lib/library/log"
	"github.com/l-dandelion/yi-ants-go/core/module/data"
	"github.com/l-dandelion/yi-ants-go/lib/constant"
	"time"
)

//start pick
func (sched *myScheduler) pick() {
	go func() {
		for {
			//stopped
			if sched.canceled() {
				break
			}
			//paused
			if sched.Status() == constant.RUNNING_STATUS_PAUSING {
				time.Sleep(100*time.Millisecond)
				continue
			}
			datum, err := sched.itemBufferPool.Get()
			if err != nil {
				log.Warnln("The item buffer pool was closed. Break item reception.")
				break
			}
			pipelinePool.Add()
			go func(datum interface{}){
				defer pipelinePool.Done()
				item, ok := datum.(data.Item)
				if !ok {
					yierr := constant.NewYiErrorf(constant.ERR_CRAWL_SCHEDULER,
						"Incorrect item type: %T, item=%+v", item, item)
					sched.sendError(yierr)
					return
				}
				sched.pickOne(item)
			}(datum)

		}
	}()
}

//pick one
func (sched *myScheduler) pickOne(item data.Item) {
	if item == nil {
		return
	}
	if sched.canceled() {
		return
	}
	pipeline := sched.pipeline
	errs := pipeline.Send(item)
	if errs != nil {
		for _, err := range errs {
			sched.sendError(constant.NewYiErrore(constant.ERR_CRAWL_PIPELINE, err))
		}
	}
}