package scheduler

import (
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/l-dandelion/yi-ants-go/core/module/data"
	"github.com/l-dandelion/yi-ants-go/lib/constant"
)

/*
 * start download
 */
func (sched *myScheduler) download() {
	go func() {
		for {
			// stopped
			if sched.canceled() {
				break
			}
			// paused
			if sched.Status() == constant.RUNNING_STATUS_PAUSED {
				time.Sleep(100*time.Millisecond)
				continue
			}
			datum, err := sched.reqBufferPool.Get()
			if err != nil {
				log.Warnln("The request buffer pool was closed. Break request reception.")
				return
			}

			downloaderPool.Add()
			go func(datum interface{}) {
				defer downloaderPool.Done()
				req, ok := datum.(*data.Request)
				if !ok {
					yierr := constant.NewYiErrorf(constant.ERR_CRAWL_SCHEDULER,
						"Incorrect request type: %T", datum)
					sched.sendError(yierr)
					return
				}
				sched.downloadOne(req)
			}(datum)
		}
		downloaderPool.Wait()
	}()
}

/*
 * download one
 */
func (sched *myScheduler) downloadOne(req *data.Request) {
	if req == nil {
		return
	}
	if sched.canceled() {
		return
	}
	downloader := sched.downloader
	resp, yierr := downloader.Download(req)
	if resp != nil {
		sched.sendResp(resp)
	}
	if yierr != nil {
		sched.sendError(yierr)
	}
}
