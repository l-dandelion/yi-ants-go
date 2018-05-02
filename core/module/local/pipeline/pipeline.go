package pipeline

import (
	"github.com/l-dandelion/yi-ants-go/core/module"
	"github.com/l-dandelion/yi-ants-go/core/module/data"
	"github.com/l-dandelion/yi-ants-go/core/module/stub"
	"github.com/l-dandelion/yi-ants-go/lib/constant"
	log "github.com/sirupsen/logrus"
)

/*
 * implementation of interface module.Pipeline
 */
type myPipeline struct {
	stub.ModuleInternal
	itemProcessors []module.ProcessItem
	failFast       bool
}

/*
 * create an instance for module.Pipeline
 */
func New(mid module.MID,
	itemProcessors []module.ProcessItem,
	scoreCalculator module.CalculateScore) (pipeline module.Pipeline, yierr *constant.YiError) {
	moduleBase, yierr := stub.NewModuleInternal(mid, scoreCalculator)
	if yierr != nil {
		return
	}
	if itemProcessors == nil {
		return nil, constant.NewYiErrorf(constant.ERR_NEW_PIPELINE, "Nil item processor list")
	}
	if len(itemProcessors) == 0 {
		return nil, constant.NewYiErrorf(constant.ERR_NEW_PIPELINE, "Empty item processor list")
	}
	var processors []module.ProcessItem
	for i, processor := range itemProcessors {
		if processor == nil {
			return nil, constant.NewYiErrorf(constant.ERR_NEW_PIPELINE, "Nil item processor[%d]", i)
		}
		processors = append(processors, processor)
	}
	return &myPipeline{
		ModuleInternal: moduleBase,
		itemProcessors: processors,
	}, nil
}

/*
 * get item processors
 */
func (pipeline *myPipeline) ItemProcessors() []module.ProcessItem {
	processors := make([]module.ProcessItem, len(pipeline.itemProcessors))
	copy(processors, pipeline.itemProcessors)
	return processors
}

/*
 * process item
 */
func (pipeline *myPipeline) Send(item data.Item) (yierrs []*constant.YiError) {
	pipeline.IncrHandlingNumber()
	defer pipeline.DecrHandlingNumber()
	pipeline.IncrCalledCount()
	if item == nil {
		yierrs = append(yierrs, constant.NewYiErrorf(constant.ERR_CRAWL_PIPELINE, "Nil item"))
		return
	}
	pipeline.IncrAcceptedCount()
	log.Infof("Process item %+v... \n", item)
	currentItem := item
	for _, processor := range pipeline.itemProcessors {
		processedItem, yierr := processor(currentItem)
		if yierr != nil {
			yierrs = append(yierrs, yierr)
			if pipeline.failFast {
				break
			}
		}
		if processedItem != nil {
			currentItem = processedItem
		}
	}
	if len(yierrs) == 0 {
		pipeline.IncrCompletedCount()
	}
	return
}

/*
 * get failFast
 */
func (pipeline *myPipeline) FailFast() bool {
	return pipeline.failFast
}

/*
 * set failFast
 */
func (pipeline *myPipeline) SetFailFast(failFast bool) {
	pipeline.failFast = failFast
}

/*
 * used in module.SummaryStruct.Extra
 */
type extraSummaryStruct struct {
	FailFast        bool `json:"fail_fast"`
	ProcessorNumber int  `json:"process_number"`
}

/*
 * rewrite Summary()
 */
func (pipeline *myPipeline) Summary() module.SummaryStruct {
	summary := pipeline.ModuleInternal.Summary()
	summary.Extra = extraSummaryStruct{
		FailFast: pipeline.failFast,
		ProcessorNumber: len(pipeline.itemProcessors),
	}
	return summary
}
