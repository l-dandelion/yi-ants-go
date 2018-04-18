package scheduler

import (
	"encoding/json"

	"github.com/l-dandelion/yi-ants-go/core/module"
	"github.com/l-dandelion/yi-ants-go/lib/constant"
	"github.com/l-dandelion/yi-ants-go/lib/library/buffer"
)

type SchedSummary interface {
	Struct() SummaryStruct
	String() (string, *constant.YiError)
}

/*
 * implementation of interface SchedSummary
 */
type mySchedSummary struct {
	requestArgs RequestArgs
	dataArgs    DataArgs
	moduleArgs  ModuleArgs
	maxDepth    uint32
	sched       *myScheduler
}

/*
 * get scheduler summary struct
 */
func (ss *mySchedSummary) Struct() SummaryStruct {
	return SummaryStruct{
		RequestArgs:     ss.requestArgs,
		DataArgs:        ss.dataArgs,
		Status:          GetStatusDescription(ss.sched.Status()),
		Downloader:      ss.sched.downloader.Summary(),
		Analyzer:        ss.sched.analyzer.Summary(),
		Pipeline:        ss.sched.pipeline.Summary(),
		ReqBufferPool:   getBufferPoolSummary(ss.sched.reqBufferPool),
		RespBufferPool:  getBufferPoolSummary(ss.sched.respBufferPool),
		ItemBufferPool:  getBufferPoolSummary(ss.sched.itemBufferPool),
		ErrorBufferPool: getBufferPoolSummary(ss.sched.errorBufferPool),
		NumURL:          ss.sched.urlMap.Len(),
	}
}

/*
 * get scheduler summary string (json)
 */
func (ss *mySchedSummary) String() (string, *constant.YiError) {
	b, err := json.MarshalIndent(ss.Struct(), "", "    ")
	if err != nil {
		return "", constant.NewYiErrore(constant.ERR_GET_SCHEDULER_SUMMARY, err)
	}
	return string(b), nil
}

/*
 * create an instance of SchedSummary
 */
func newSchedSummary(requestArgs RequestArgs, dataArgs DataArgs, moduleArgs ModuleArgs, sched *myScheduler) SchedSummary {
	if sched == nil {
		return nil
	}
	return &mySchedSummary{
		requestArgs: requestArgs,
		dataArgs:    dataArgs,
		moduleArgs:  moduleArgs,
		maxDepth:    requestArgs.MaxDepth,
		sched:       sched,
	}
}

/*
 * scheduler summary struct
 */
type SummaryStruct struct {
	RequestArgs     RequestArgs             `json:"request_args"`
	DataArgs        DataArgs                `json:"data_args"`
	Status          string                  `json:"status"`
	Downloader      module.SummaryStruct    `json:"downloader"`
	Analyzer        module.SummaryStruct    `json:"analyzer"`
	Pipeline        module.SummaryStruct    `json:"pipeline"`
	ReqBufferPool   BufferPoolSummaryStruct `json:"request_buffer_pool"`
	RespBufferPool  BufferPoolSummaryStruct `json:"response_buffer_pool"`
	ItemBufferPool  BufferPoolSummaryStruct `json:"item_buffer_pool"`
	ErrorBufferPool BufferPoolSummaryStruct `json:"error_buffer_pool"`
	NumURL          uint64                  `json:"url_number"`
}

/*
 * compare two summary struct
 */
func (one *SummaryStruct) Same(anthor SummaryStruct) bool {
	if !one.RequestArgs.Same(&anthor.RequestArgs) {
		return false
	}
	if one.DataArgs != anthor.DataArgs {
		return false
	}
	if one.Status != anthor.Status {
		return false
	}

	if anthor.Downloader != one.Downloader ||
		anthor.Pipeline != one.Pipeline ||
		anthor.Analyzer != one.Analyzer {
		return false
	}

	if one.ReqBufferPool != anthor.ReqBufferPool ||
		one.RespBufferPool != anthor.RespBufferPool ||
		one.ItemBufferPool != anthor.ItemBufferPool ||
		one.ErrorBufferPool != anthor.ErrorBufferPool {
		return false
	}

	if one.NumURL != anthor.NumURL {
		return false
	}

	return true
}

/*
 * buffer pool summary struct
 */
type BufferPoolSummaryStruct struct {
	BufferCap       uint32 `json:"buffer_cap"`
	MaxBufferNumber uint32 `json:"max_buffer_number"`
	BufferNumber    uint32 `json:"buffer_number"`
	Total           uint64 `json:"total"`
}

/*
 * get buffer pool summary
 */
func getBufferPoolSummary(bufferPool buffer.Pool) BufferPoolSummaryStruct {
	return BufferPoolSummaryStruct{
		BufferCap:       bufferPool.BufferCap(),
		MaxBufferNumber: bufferPool.MaxBufferNumber(),
		BufferNumber:    bufferPool.BufferNumber(),
		Total:           bufferPool.Total(),
	}
}
