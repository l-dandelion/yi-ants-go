package scheduler

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/l-dandelion/yi-ants-go/core/module"
	"github.com/l-dandelion/yi-ants-go/core/module/data"
	"github.com/l-dandelion/yi-ants-go/core/module/local/analyzer"
	"github.com/l-dandelion/yi-ants-go/core/module/local/downloader"
	"github.com/l-dandelion/yi-ants-go/core/module/local/pipeline"
	"github.com/l-dandelion/yi-ants-go/lib/constant"
	"github.com/l-dandelion/yi-ants-go/lib/library/log"
)

func TestArgsRequest(t *testing.T) {
	requestArgs := genRequestArgs([]string{}, 0)
	if err := requestArgs.Check(); err != nil {
		t.Fatalf("Inconsistent check result: expected: %v, actual: %v",
			nil, err)
	}
	requestArgs = genRequestArgs(nil, 0)
	if err := requestArgs.Check(); err == nil {
		t.Fatalf("Inconsistent check result: expected: %v, actual: %v",
			nil, err)
	}
	// 测试Same方法的正确性。
	one := genRequestArgs([]string{
		"bing.com",
	}, 0)
	another := genRequestArgs([]string{
		"bing.com",
	}, 0)
	same := one.Same(&another)
	if !same {
		t.Fatalf("Inconsistent request arguments sameness: expected: %v, actual: %v",
			true, same)
	}
	same = one.Same(nil)
	if same {
		t.Fatalf("Inconsistent request arguments sameness with nil parameter: expected: %v, actual: %v",
			false, same)
	}
	another = genRequestArgs([]string{
		"bing.com",
	}, 1)
	same = one.Same(&another)
	if same {
		t.Fatalf("Inconsistent request arguments sameness with different max depth: expected: %v, actual: %v",
			false, same)
	}
	another = genRequestArgs(nil, 0)
	same = one.Same(&another)
	if same {
		t.Fatalf("Inconsistent request arguments sameness with different accepted domains: expected: %v, actual: %v",
			false, same)
	}
	another = genRequestArgs([]string{
		"bing.net",
		"bing.com",
	}, 0)
	same = one.Same(&another)
	if same {
		t.Fatalf("Inconsistent request arguments sameness with different accepted domains: expected: %v, actual: %v",
			false, same)
	}
	one = genRequestArgs([]string{
		"sogou.com",
		"bing.com",
	}, 0)
	same = one.Same(&another)
	if same {
		t.Fatalf("Inconsistent request arguments sameness with different accepted domains: expected: %v, actual: %v",
			false, same)
	}
}

func TestArgsData(t *testing.T) {
	dataArgs := genDataArgs(10, 2, 1)
	if err := dataArgs.Check(); err != nil {
		t.Fatalf("Inconsistent check result: expected: %v, actual: %v",
			nil, err)
	}
	dataArgsList := []DataArgs{}
	for i := 0; i < 8; i++ {
		values := [8]uint32{2, 2, 2, 2, 2, 2, 2, 2}
		values[i] = 0
		dataArgsList = append(
			dataArgsList, genDataArgsByDetail(values))
	}
	for _, dataArgs := range dataArgsList {
		if err := dataArgs.Check(); err == nil {
			t.Fatalf("No error when check data arguments! (dataArgs: %#v)",
				dataArgs)
		}
	}
}

// create an instance of request args
func genRequestArgs(acceptedDomains []string, maxDepth uint32) RequestArgs {
	return RequestArgs{
		AcceptedDomains: acceptedDomains,
		MaxDepth:        maxDepth,
	}
}

// create an instance of data args
func genDataArgs(
	bufferCap uint32, maxBufferNumber uint32, stepLen uint32) DataArgs {
	values := [8]uint32{}
	var bufferCapStep uint32
	var maxBufferNumberStep uint32
	for i := uint32(0); i < 8; i++ {
		if i%2 == 0 {
			values[i] = bufferCap + bufferCapStep*stepLen
			bufferCapStep++
		} else {
			values[i] = maxBufferNumber + maxBufferNumberStep*stepLen
			maxBufferNumberStep++
		}
	}
	return genDataArgsByDetail(values)
}

// create an instance of data args by detail
func genDataArgsByDetail(values [8]uint32) DataArgs {
	return DataArgs{
		ReqBufferCap:         values[0],
		ReqMaxBufferNumber:   values[1],
		RespBufferCap:        values[2],
		RespMaxBufferNumber:  values[3],
		ItemBufferCap:        values[4],
		ItemMaxBufferNumber:  values[5],
		ErrorBufferCap:       values[6],
		ErrorMaxBufferNumber: values[7],
	}
}

// create an simple instance of module args
func genSimpleModuleArgs(t *testing.T) ModuleArgs {
	snGen := module.NewSNGenerator(1, 0)
	return ModuleArgs{
		Downloader: genSimpleDownloaders(1, false, snGen, t)[0],
		Analyzer:   genSimpleAnalyzers(1, false, snGen, t)[0],
		Pipeline:   genSimplePipelines(1, false, snGen, t)[0],
	}
}

// create simple instaces of downloader
func genSimpleDownloaders(number int8, reuseMID bool, snGen module.SNGenerator, t *testing.T) []module.Downloader {
	if number < -1 {
		return []module.Downloader{nil}
	} else if number == -1 { // 不合规的MID。
		mid := module.MID(fmt.Sprintf("A%d", snGen.Get()))
		httpClient := &http.Client{}
		d, err := downloader.New(mid, httpClient, nil)
		if err != nil {
			t.Fatalf("An error occurs when creating a downloader: %s (mid: %s, httpClient: %#v)",
				err, mid, httpClient)
		}
		return []module.Downloader{d}
	}
	results := make([]module.Downloader, number)
	var mid module.MID
	for i := int8(0); i < number; i++ {
		if i == 0 || !reuseMID {
			mid = module.MID(fmt.Sprintf("D%d", snGen.Get()))
		}
		httpClient := &http.Client{}
		d, err := downloader.New(mid, httpClient, nil)
		if err != nil {
			t.Fatalf("An error occurs when creating a downloader: %s (mid: %s, httpClient: %#v)",
				err, mid, httpClient)
		}
		results[i] = d
	}
	return results
}

// create simple instances of analyzer
func genSimpleAnalyzers(number int8, reuseMID bool, snGen module.SNGenerator, t *testing.T) []module.Analyzer {
	respParsers := []module.ParseResponse{parseATag}
	if number < -1 {
		return []module.Analyzer{nil}
	} else if number == -1 { // 不合规的MID。
		mid := module.MID(fmt.Sprintf("P%d", snGen.Get()))
		a, err := analyzer.New(mid, respParsers, nil)
		if err != nil {
			t.Fatalf("An error occurs when creating an analyzer: %s (mid: %s, respParsers: %#v)",
				err, mid, respParsers)
		}
		return []module.Analyzer{a}
	}
	results := make([]module.Analyzer, number)
	var mid module.MID
	for i := int8(0); i < number; i++ {
		if i == 0 || !reuseMID {
			mid = module.MID(fmt.Sprintf("A%d", snGen.Get()))
		}
		a, err := analyzer.New(mid, respParsers, nil)
		if err != nil {
			t.Fatalf("An error occurs when creating an analyzer: %s (mid: %s, respParsers: %#v)",
				err, mid, respParsers)
		}
		results[i] = a
	}
	return results
}

// create simple instances of pipeline
func genSimplePipelines(number int8, reuseMID bool, snGen module.SNGenerator, t *testing.T) []module.Pipeline {
	processors := []module.ProcessItem{processItem}
	if number < -1 {
		return []module.Pipeline{nil}
	} else if number == -1 { // 不合规的MID。
		mid := module.MID(fmt.Sprintf("D%d", snGen.Get()))
		p, err := pipeline.New(mid, processors, nil)
		if err != nil {
			t.Fatalf("An error occurs when creating a pipeline: %s (mid: %s, processors: %#v)",
				err, mid, processors)
		}
		return []module.Pipeline{p}
	}
	results := make([]module.Pipeline, number)
	var mid module.MID
	for i := int8(0); i < number; i++ {
		if i == 0 || !reuseMID {
			mid = module.MID(fmt.Sprintf("P%d", snGen.Get()))
		}
		p, err := pipeline.New(mid, processors, nil)
		if err != nil {
			t.Fatalf("An error occurs when creating a pipeline: %s (mid: %s, processors: %#v)",
				err, mid, processors)
		}
		results[i] = p
	}
	return results
}

// function for parsing response
func parseATag(resp *data.Response) ([]data.Data, []*constant.YiError) {
	httpResp := resp.HTTPResp()
	//TODO: 支持更多的HTTP响应状态。
	if httpResp.StatusCode != 200 {
		err := fmt.Errorf(
			fmt.Sprintf("Unsupported status code %d! (httpResponse: %v)",
				httpResp.StatusCode, httpResp))
		return nil, []*constant.YiError{constant.NewYiErrore(constant.ERR_CRAWL_ANALYZER, err)}
	}
	reqURL := httpResp.Request.URL
	httpRespBody := httpResp.Body
	defer func() {
		if httpRespBody != nil {
			httpRespBody.Close()
		}
	}()
	var dataList []data.Data
	var yierrs []*constant.YiError
	// begin
	doc, err := goquery.NewDocumentFromReader(httpRespBody)
	if err != nil {
		yierrs = append(yierrs, constant.NewYiErrore(constant.ERR_CRAWL_ANALYZER, err))
		return dataList, yierrs
	}
	defer httpRespBody.Close()
	// find tag "a" and get url
	doc.Find("a").Each(func(index int, sel *goquery.Selection) {
		href, exists := sel.Attr("href")
		// filter
		if !exists || href == "" || href == "#" || href == "/" {
			return
		}
		href = strings.TrimSpace(href)
		lowerHref := strings.ToLower(href)
		if href != "" && !strings.HasPrefix(lowerHref, "javascript") {
			aURL, err := url.Parse(href)
			if err != nil {
				log.Warnf("An error occurs when parsing attribute %q in tag %q : %s (href: %s)",
					err, "href", "a", href)
				return
			}
			if !aURL.IsAbs() {
				aURL = reqURL.ResolveReference(aURL)
			}
			httpReq, err := http.NewRequest("GET", aURL.String(), nil)
			if err != nil {
				yierrs = append(yierrs, constant.NewYiErrore(constant.ERR_CRAWL_ANALYZER, err))
			} else {
				req := data.NewRequest(httpReq)
				dataList = append(dataList, req)
			}
		}
		text := strings.TrimSpace(sel.Text())
		var id, name string
		if v, ok := sel.Attr("id"); ok {
			id = strings.TrimSpace(v)
		}
		if v, ok := sel.Attr("name"); ok {
			name = strings.TrimSpace(v)
		}
		m := make(map[string]interface{})
		m["a.parent"] = reqURL
		m["a.id"] = id
		m["a.name"] = name
		m["a.text"] = text
		m["a.index"] = index
		item := data.Item(m)
		dataList = append(dataList, item)
		log.Infof("Processed item: %v", m)
	})
	return dataList, yierrs
}

// function for processing item
func processItem(item data.Item) (result data.Item, yierr *constant.YiError) {
	if item == nil {
		return nil, constant.NewYiErrorf(constant.ERR_CRAWL_PIPELINE, "Invalid item!")
	}
	// generate result
	result = make(map[string]interface{})
	for k, v := range item {
		result[k] = v
	}
	if _, ok := result["number"]; !ok {
		result["number"] = len(result)
	}
	time.Sleep(10 * time.Millisecond)
	return result, nil
}
