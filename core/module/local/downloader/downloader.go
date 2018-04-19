package downloader

import (
	"net/http"

	"github.com/l-dandelion/yi-ants-go/core/module"
	"github.com/l-dandelion/yi-ants-go/core/module/data"
	"github.com/l-dandelion/yi-ants-go/core/module/stub"
	"github.com/l-dandelion/yi-ants-go/lib/constant"
	"github.com/l-dandelion/yi-ants-go/lib/library/log"
)

const RETRY_TIMES = 3

/*
 * create an instance for downloader
 */
func New(
	mid module.MID,
	client *http.Client,
	scoreCalculator module.CalculateScore) (downloader module.Downloader, yierr *constant.YiError) {
	moduleBase, yierr := stub.NewModuleInternal(mid, scoreCalculator)
	//check whether the args are vaild
	if yierr != nil {
		return
	}
	if client == nil {
		yierr = constant.NewYiErrorf(constant.ERR_NEW_DOWNLOADER_FAIL, "Client is nil.")
		return
	}

	return &myDownloader{
		ModuleInternal: moduleBase,
		httpClient:     client,
	}, nil
}

/*
 * implementation of interface module.Downloader
 */
type myDownloader struct {
	stub.ModuleInternal              //module internal instance
	httpClient          *http.Client //http client for downloading
}

/*
 * download according to request, return a response if success, or an error return
 */
func (downloader *myDownloader) Download(req *data.Request) (*data.Response, *constant.YiError) {
	downloader.IncrHandlingNumber()
	defer downloader.DecrHandlingNumber()
	downloader.IncrCalledCount()

	//check whether request is vaild.
	if req == nil {
		return nil, constant.NewYiErrorf(constant.ERR_CRAWL_DOWNLOADER, "Request is nil.")
	}
	if req.HTTPReq() == nil {
		return nil, constant.NewYiErrorf(constant.ERR_CRAWL_DOWNLOADER, "HTTP request is nil.")
	}
	downloader.IncrAcceptedCount()

	var (
		httpResp *http.Response
		err      error
	)

	log.Infof("Do the request (URL: %s, depth: %d)... \n",
		req.HTTPReq().URL, req.Depth())
	// try to download RETRY_TIMES times
	for i := 0; i < RETRY_TIMES; i++ {
		httpResp, err = downloader.httpClient.Do(req.HTTPReq())
		if err == nil {
			break
		}
	}

	if err != nil {
		return nil, constant.NewYiErrore(constant.ERR_CRAWL_DOWNLOADER, err)
	}
	resp := data.NewResponse(req, httpResp)
	downloader.IncrCompletedCount()
	return resp, nil
}
