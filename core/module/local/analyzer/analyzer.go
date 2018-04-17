package analyzer

import (
	"github.com/l-dandelion/yi-ants-go/core/module"
	"github.com/l-dandelion/yi-ants-go/core/module/stub"
	"github.com/l-dandelion/yi-ants-go/core/module/data"
	"github.com/l-dandelion/yi-ants-go/lib/constant"
	"github.com/l-dandelion/yi-ants-go/lib/library/log"
	"github.com/l-dandelion/yi-ants-go/lib/library/reader"
)

/*
 * create an instance for analyzer
 */
func New(
	mid module.MID,
	respParsers []module.ParseResponse,
	scoreCalculator module.CalculateScore) (analyzer module.Analyzer, yierr *constant.YiError) {
	moduleBase, yierr := stub.NewModuleInternal(mid, scoreCalculator)
	if yierr != nil {
		return
	}
	if respParsers == nil {
		yierr = constant.NewYiErrorf(constant.ERR_NEW_ANALYZER_FAIL,
			"Response parsers is nil")
		return
	}
	if len(respParsers) == 0 {
		yierr = constant.NewYiErrorf(constant.ERR_NEW_ANALYZER_FAIL,
			"Empty response parser list")
		return
	}
	var innerParsers []module.ParseResponse
	for i, parser := range respParsers {
		if parser == nil {
			yierr = constant.NewYiErrorf(constant.ERR_NEW_ANALYZER_FAIL,
				"Nil response parser[%d]", i)
			return
		}
		innerParsers = append(innerParsers, parser)
	}
	return &myAnalyzer{
		ModuleInternal: moduleBase,
		respParsers:    innerParsers,
	}, nil
}


/*
 * implementation of interface module.Analyzer
 */
type myAnalyzer struct {
	stub.ModuleInternal                        // module internal instance
	respParsers         []module.ParseResponse //response parser list
}

/*
 * copy response parser list and return
 */
func (analyzer *myAnalyzer)RespParsers() []module.ParseResponse {
	parsers := make([]module.ParseResponse, len(analyzer.respParsers))
	copy(parsers, analyzer.respParsers)
	return parsers
}

/*
 * analyze the response and return the data list, or error list
 */
func (analyzer *myAnalyzer) Analyze(
	resp *data.Response) (dataList []data.Data, yierrList []*constant.YiError) {
	analyzer.IncrHandlingNumber()
	defer analyzer.DecrHandlingNumber()
	analyzer.IncrCalledCount()
	yierrList = []*constant.YiError{}
	// check args
	if resp == nil {
		yierr := constant.NewYiErrorf(constant.ERR_CRAWL_ANALYZER, "Response is nil")
		yierrList = append(yierrList, yierr)
		return
	}
	httpResp := resp.HTTPResp()
	if httpResp == nil {
		yierr := constant.NewYiErrorf(constant.ERR_CRAWL_ANALYZER, "HTTP response is nil")
		yierrList = append(yierrList, yierr)
		return
	}
	req := resp.Request()
	if req == nil {
		yierr := constant.NewYiErrorf(constant.ERR_CRAWL_ANALYZER, "Request is nil")
		yierrList = append(yierrList, yierr)
		return
	}
	httpReq := req.HTTPReq()
	if httpReq == nil {
		yierr := constant.NewYiErrorf(constant.ERR_CRAWL_ANALYZER, "HTTP request is nil")
		yierrList = append(yierrList, yierr)
		return
	}
	reqURL := httpReq.URL
	if reqURL == nil {
		yierr := constant.NewYiErrorf(constant.ERR_CRAWL_ANALYZER, "Request URL is nil")
		yierrList = append(yierrList, yierr)
		return
	}
	analyzer.IncrAcceptedCount()

	respDepth := resp.Depth()
	log.Infof("Parse the response (URL: %s, depth: %d)... \n", reqURL, respDepth)
	if httpResp.Body != nil {
		defer httpResp.Body.Close()
	}
	multiReader, err := reader.NewMultipleReader(httpResp.Body)
	if err != nil {
		yierr := constant.NewYiErrore(constant.ERR_CRAWL_ANALYZER, err)
		yierrList = append(yierrList, yierr)
		return
	}
	dataList = []data.Data{}
	for _, respParser := range analyzer.respParsers {
		if httpResp.Body != nil {
			httpResp.Body.Close()
		}
		httpResp.Body = multiReader.Reader()
		pDataList, pYierrList := respParser(resp)
		if pDataList != nil {
			for _, mdata := range pDataList {
				dataList = appendDataList(dataList, mdata, respDepth)
			}
		}
		if pYierrList != nil {
			for _, yierr := range pYierrList {
				yierrList = append(yierrList, yierr)
			}
		}
	}
	if len(yierrList) == 0 {
		analyzer.IncrCompletedCount()
	}
	return
}

/*
 * add data(request or item) to data list
 */
func appendDataList(dataList []data.Data, mdata data.Data, respDepth uint32) []data.Data {
	if mdata == nil {
		return dataList
	}
	req, ok := mdata.(*data.Request)
	if ok {
		req.SetDepth(respDepth + 1)
	}
	return append(dataList, mdata)
}
