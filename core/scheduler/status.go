package scheduler

import (
	"github.com/l-dandelion/yi-ants-go/lib/constant"
)

/*
 * check whether status can be changed to wantedStatus from currentStatus
 */
func checkStatus(currentStatus, wantedStatus int8) (yierr *constant.YiError) {
	switch currentStatus {
	case constant.RUNNING_STATUS_PREPARING:
		yierr = constant.NewYiErrorf(constant.ERR_CRAWL_SCHEDULER, "The scheduler is being initializd!")
	case constant.RUNNING_STATUS_STARTING:
		yierr = constant.NewYiErrorf(constant.ERR_CRAWL_SCHEDULER, "The scheduler is being started!")
	case constant.RUNNING_STATUS_STOPPING:
		yierr = constant.NewYiErrorf(constant.ERR_CRAWL_SCHEDULER, "The scheduler is being stopped!")
	case constant.RUNNING_STATUS_PAUSING:
		yierr = constant.NewYiErrorf(constant.ERR_CRAWL_SCHEDULER, "The scheduler is being paused!")
	}
	if yierr != nil {
		return
	}

	if currentStatus == constant.RUNNING_STATUS_UNPREPARED &&
		wantedStatus != constant.RUNNING_STATUS_PREPARING {
		yierr = constant.NewYiErrorf(constant.ERR_CRAWL_SCHEDULER, "the scheduler has not yet been initialized!")
	}

	switch wantedStatus {
	case constant.RUNNING_STATUS_PREPARING:
		switch currentStatus {
		case constant.RUNNING_STATUS_STARTED:
			yierr = constant.NewYiErrorf(constant.ERR_CRAWL_SCHEDULER, "the scheduler has been started!")
		case constant.RUNNING_STATUS_PAUSED:
			yierr = constant.NewYiErrorf(constant.ERR_CRAWL_SCHEDULER, "the scheduler has not been stopped!")
		}
	case constant.RUNNING_STATUS_STARTING:
		switch currentStatus {
		case constant.RUNNING_STATUS_UNPREPARED:
			yierr = constant.NewYiErrorf(constant.ERR_CRAWL_SCHEDULER, "the scheduler has not been initialized!")
		case constant.RUNNING_STATUS_STARTED:
			yierr = constant.NewYiErrorf(constant.ERR_CRAWL_SCHEDULER, "the scheduler has been started!")
		}
	case constant.RUNNING_STATUS_PAUSING:
		if currentStatus != constant.RUNNING_STATUS_STARTED {
			yierr = constant.NewYiErrorf(constant.ERR_CRAWL_SCHEDULER, "the scheduler has not been started!")
		}
	case constant.RUNNING_STATUS_STOPPING:
		if currentStatus != constant.RUNNING_STATUS_STARTED &&
			currentStatus != constant.RUNNING_STATUS_PAUSED {
			yierr = constant.NewYiErrorf(constant.ERR_CRAWL_SCHEDULER, "the scheduler has not been started!")
		}
	default:
		yierr = constant.NewYiErrorf(constant.ERR_CRAWL_SCHEDULER,
			"unsupported wanted status for check! (wantedStatus: %d)", wantedStatus)
	}
	return
}

/*
 * get status description
 */
func GetStatusDescription(status int8) string {
	desc, ok := constant.GlobalArrRunningStatusDesc[status]
	if !ok {
		return "Unknow"
	}
	return desc
}