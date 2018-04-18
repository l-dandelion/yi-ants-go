package scheduler

import (
	"github.com/l-dandelion/yi-ants-go/core/module"
	"github.com/l-dandelion/yi-ants-go/lib/constant"
)

/*
 * args for init scheduler
 */
type Args interface {
	Check() *constant.YiError //check whether it is vaild
}

/*
 * implementation of interface Args
 */
type RequestArgs struct {
	AcceptedDomains []string `json:"accepted_primary_domains"` //accepted domains
	MaxDepth        uint32   `json:"max_depth"`                //max crawl depth
}

/*
 * check whether the request args is vaild
 */
func (args *RequestArgs) Check() *constant.YiError {
	if args.AcceptedDomains == nil {
		return constant.NewYiErrorf(constant.ERR_ARGS, "Nil accepted domains")
	}
	return nil
}

/*
 * check whether it is same as anthor
 */
func (args *RequestArgs) Same(anthor *RequestArgs) bool {
	if anthor == nil {
		return false
	}
	if args.MaxDepth != anthor.MaxDepth {
		return false
	}
	if len(args.AcceptedDomains) != len(anthor.AcceptedDomains) {
		return false
	}
	if anthor.AcceptedDomains != nil {
		for i, acceptedDomain := range anthor.AcceptedDomains {
			if args.AcceptedDomains[i] != acceptedDomain {
				return false
			}
		}
	}
	return true
}

/*
 * implementation of interface Args
 */
type DataArgs struct {
	ReqBufferCap         uint32 `json:"req_buffer_cap"`          // request buffer capacity
	ReqMaxBufferNumber   uint32 `json:"req_max_buffer_number"`   // max request buffer number
	RespBufferCap        uint32 `json:"resp_buffer_cap"`         // response buffer capacity
	RespMaxBufferNumber  uint32 `json:"resp_max_buffer_number"`  // max response buffer number
	ItemBufferCap        uint32 `json:"item_buffer_cap"`         // item buffer capacity
	ItemMaxBufferNumber  uint32 `json:"item_max_buffer_number"`  // max item buffer number
	ErrorBufferCap       uint32 `json:"error_buffer_cap"`        // error buffer capacity
	ErrorMaxBufferNumber uint32 `json:"error_max_buffer_number"` // max error buffer number
}

/*
 * check whether the data args is vaild
 */
func (args *DataArgs) Check() *constant.YiError {
	if args.ReqBufferCap == 0 {
		return constant.NewYiErrorf(constant.ERR_ARGS, "Zero request buffer capacity.")
	}
	if args.ReqMaxBufferNumber == 0 {
		return constant.NewYiErrorf(constant.ERR_ARGS, "Zero max request buffer number.")
	}
	if args.RespBufferCap == 0 {
		return constant.NewYiErrorf(constant.ERR_ARGS, "Zero response buffer capacity.")
	}
	if args.RespMaxBufferNumber == 0 {
		return constant.NewYiErrorf(constant.ERR_ARGS, "Zero max response buffer capacity.")
	}
	if args.ItemBufferCap == 0 {
		return constant.NewYiErrorf(constant.ERR_ARGS, "Zero item buffer capacity.")
	}
	if args.ItemMaxBufferNumber == 0 {
		return constant.NewYiErrorf(constant.ERR_ARGS, "Zero max item buffer capatity.")
	}
	if args.ErrorBufferCap == 0 {
		return constant.NewYiErrorf(constant.ERR_ARGS, "Zero error buffer capacity.")
	}
	if args.ErrorMaxBufferNumber == 0 {
		return constant.NewYiErrorf(constant.ERR_ARGS, "Zero max error buffer number.")
	}
	return nil
}

/*
 * implementation of interface Args
 */
type ModuleArgs struct {
	Downloader module.Downloader //downloader
	Analyzer   module.Analyzer   //analyzer
	Pipeline   module.Pipeline   //pipeline
}

/*
 * check whether the module args is vaild.
 */
func (args *ModuleArgs) Check() *constant.YiError {
	if args.Downloader == nil {
		return constant.NewYiErrorf(constant.ERR_SCHEDULER_ARGS, "Nil downloader.")
	}
	if args.Analyzer == nil {
		return constant.NewYiErrorf(constant.ERR_SCHEDULER_ARGS, "Nil analyzer.")
	}
	if args.Pipeline == nil {
		return constant.NewYiErrorf(constant.ERR_SCHEDULER_ARGS, "Nil pipeline.")
	}
	return nil
}
