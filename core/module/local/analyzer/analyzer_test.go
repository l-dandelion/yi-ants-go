package analyzer

import (
	"net/http"
	"strings"
	"testing"

	"bufio"
	"fmt"
	"github.com/l-dandelion/yi-ants-go/core/module"
	"github.com/l-dandelion/yi-ants-go/core/module/data"
	"github.com/l-dandelion/yi-ants-go/lib/constant"
	"strconv"
	"github.com/l-dandelion/yi-ants-go/core/module/stub"
)

//implementation of interface io.ReadCloser
type testingReader struct {
	sr *strings.Reader
}

func (r testingReader) Read(b []byte) (n int, err error) {
	return r.sr.Read(b)
}

func (r testingReader) Close() error {
	return nil
}

func TestNew(t *testing.T) {
	mid := module.MID("A1|127.0.0.1:8080")
	parsers := []module.ParseResponse{genTestingRespParser(false)}
	a, err := New(mid, parsers, nil)
	if err != nil {
		t.Fatalf("An error occurs when creating an analyzer: %s(mid: %s)", err, mid)
	}
	if a == nil {
		t.Fatalf("Couldn't create analyzer!")
	}
	if a.ID() != mid {
		t.Fatalf("Inconsistent MID for analyzer: expected: %s, actual: %s", mid, a.ID())
	}
	if len(a.RespParsers()) != len(parsers) {
		t.Fatalf("Inconsistent response parser number for pipeline: expected: %d, actual: %d", len(parsers), len(a.RespParsers()))
	}

	//wrong args
	mid = module.MID("A127.0.0.1")
	a, err = New(mid, parsers, nil)
	if err == nil {
		t.Fatalf("No error when create an analyzer with illegal MID %q", mid)
	}
	mid = module.MID("D1|127.0.0.1:8000")
	parserList := [][]module.ParseResponse{
		nil,
		[]module.ParseResponse{},
		[]module.ParseResponse{genTestingRespParser(false), nil},
	}
	for _, parsers := range parserList {
		a, err = New(mid, parsers, nil)
		if err == nil {
			t.Fatalf("No error when create an analyzer with illegal parsers %#v!", parsers)
		}
	}
}

func TestAnalyze(t *testing.T) {
	number := uint32(10)
	method := "GET"
	expectedURL := "https://github.com/gopcp"
	expectedDepth := uint32(1)
	resps := getTestingResps(number, method, expectedURL, expectedDepth, t)
	mid := module.MID("A1|127.0.0.1:8080")
	parsers := []module.ParseResponse{genTestingRespParser(false)}
	a, yierr := New(mid, parsers, nil)
	if yierr != nil {
		t.Fatalf("An error occurs when creating an analyzer: %s (mid: %s)", yierr, mid)
	}
	mdata := []data.Data{}
	parseErrors := []*constant.YiError{}
	for _, resp := range resps {
		data1, parseErrors1 := a.Analyze(resp)
		mdata = append(mdata, data1...)
		parseErrors = append(parseErrors, parseErrors1...)
	}
	for i, e := range parseErrors {
		t.Errorf("An error occurs when parsing response: %s (index: %d)", e, i)
	}

	var count int
	for i, d := range mdata {
		if d == nil {
			t.Fatalf("nil datum!(index: %d)", i)
		}
		if _, ok := d.(*data.Request); ok {
			continue
		}
		item, ok := d.(data.Item)
		if !ok {
			t.Errorf("Inconsistent datum type: expected: %T, actual: %T (index: %d)", data.Item{}, d, i)
		}
		if item["url"] != expectedURL {
			t.Errorf("Inconsistent URL: expected: %s, actual: %s (index: %d)", expectedURL, item["url"], i)
		}
		index, ok := item["index"].(int)
		if !ok {
			t.Errorf("Inconsistent index type: expected: %T, actual: %T (index: %d)", int(0), item["index"], i)
		}
		if index != count {
			t.Errorf("Inconsistent index: expected: %d, actual: %d (index: %d)", count, index, i)
		}
		depth, ok := item["depth"].(uint32)
		if !ok {
			t.Errorf("Inconsistent depth type: expected: %T, actual: %T (index:%d)", uint32(0), item["depth"], i)
		}
		if depth != expectedDepth {
			t.Errorf("Inconsistent depth: expected: %d, actual: %d (index: %d)", expectedDepth, depth, i)
		}
		count ++
	}

	/*
	 * wrong args
	 */

	//nil response
	_, yierrs := a.Analyze(nil)
	if len(yierrs) == 0 {
		t.Fatal("No error when analyze with nil response!")
	}
	//nil HTTP response
	resp := data.NewResponse(nil, nil)
	_, yierrs = a.Analyze(resp)
	if len(yierrs) == 0 {
		t.Fatal("No error when analyze response with illegal response %#v!", resp)
	}
	//nil HTTP request
	httpResp := &http.Response{
		Request: nil,
		Body: nil,
	}
	resp = data.NewResponse(nil, httpResp)
	_, yierrs = a.Analyze(resp)
	if len(yierrs) == 0 {
		t.Fatal("No error when analyze response with nil request URL!")
	}
	//nil HTTP request URL
	httpReq, _ := http.NewRequest(method, expectedURL, nil)
	httpReq.URL = nil
	httpResp = &http.Response{
		Request: httpReq,
		Body: nil,
	}
	req := data.NewRequest(httpReq)
	resp = data.NewResponse(req, httpResp)
	_, yierrs = a.Analyze(resp)
	if len(yierrs) == 0 {
		t.Fatal("No error when analyze response with nil request URL!")
	}
}

func TestCount(t *testing.T) {
	mid := module.MID("A1|127.0.0.1:8080")
	//counts after initialted
	parsers := []module.ParseResponse{genTestingRespParser(false)}
	a, _ := New(mid, parsers, nil)
	ai := a.(stub.ModuleInternal)
	if ai.CalledCount() != 0 {
		t.Fatalf("Inconsistent called count for internal module: expected: %d, actual: %d", 0, ai.CalledCount())
	}
	if ai.AcceptedCount() != 0 {
		t.Fatalf("Inconsistent accepted count for internal module: expected: %d, actual: %d", 0, ai.AcceptedCount())
	}
	if ai.CompletedCount() != 0 {
		t.Fatalf("Inconsistent completed count for internal module: expected: %d, actual: %d", 0, ai.CompletedCount())
	}
	if ai.HandlingNumber() != 0 {
		t.Fatalf("Inconsistent handling number for internal module: expected: %d, actual: %d", 0, ai.HandlingNumber())
	}

	//counts after fail
	parsers = []module.ParseResponse{genTestingRespParser(true)}
	a, _ = New(mid, parsers, nil)
	ai = a.(stub.ModuleInternal)
	resp := getTestingResps(1, "GET", "https://github.com/gopcp", 0, t)[0]
	a.Analyze(resp)
	if ai.CalledCount() != 1 {
		t.Fatalf("Inconsistent called count for internal module: expected: %d, actual: %d", 1, ai.CalledCount())
	}
	if ai.AcceptedCount() != 1 {
		t.Fatalf("Inconsistent accepted count for internal module: expected: %d, actual: %d", 1, ai.AcceptedCount())
	}
	if ai.CompletedCount() != 0 {
		t.Fatalf("Inconsistent completed count for internal module: expected: %d, actual: %d", 0, ai.CompletedCount())
	}
	if ai.HandlingNumber() != 0 {
		t.Fatalf("Inconsistent handling number for internal module: expected: %d, actual: %d", 0, ai.HandlingNumber())
	}

	//counts with wrong args
	parsers = []module.ParseResponse{genTestingRespParser(false)}
	a, _ = New(mid, parsers, nil)
	ai = a.(stub.ModuleInternal)
	resp = data.NewResponse(nil, nil)
	a.Analyze(resp)
	if ai.CalledCount() != 1 {
		t.Fatalf("Inconsistent called count for internal module: expected: %d, actual: %d", 1, ai.CalledCount())
	}
	if ai.AcceptedCount() != 0 {
		t.Fatalf("Inconsistent accepted count for internal module: expected: %d, actual: %d", 0, ai.AcceptedCount())
	}
	if ai.CompletedCount() != 0 {
		t.Fatalf("Inconsistent completed count for internal module: expected: %d, actual: %d", 0, ai.CompletedCount())
	}
	if ai.HandlingNumber() != 0 {
		t.Fatalf("Inconsistent handling number for internal module: expected: %d, actual: %d", 0, ai.HandlingNumber())
	}

	//counts after success
	parsers = []module.ParseResponse{genTestingRespParser(false)}
	a, _ = New(mid, parsers, nil)
	ai = a.(stub.ModuleInternal)
	resp = getTestingResps(1, "GET", "https://github.com/gopcp", 0, t)[0]
	a.Analyze(resp)
	if ai.CalledCount() != 1 {
		t.Fatalf("Inconsistent called count for internal module: expected: %d, actual: %d",
			1, ai.CalledCount())
	}
	if ai.AcceptedCount() != 1 {
		t.Fatalf("Inconsistent accepted count for internal module: expected: %d, actual: %d",
			1, ai.AcceptedCount())
	}
	if ai.CompletedCount() != 1 {
		t.Fatalf("Inconsistent completed count for internal module: expected: %d, actual: %d",
			1, ai.CompletedCount())
	}
	if ai.HandlingNumber() != 0 {
		t.Fatalf("Inconsistent handling number for internal module: expected: %d, actual: %d",
			0, ai.HandlingNumber())
	}
}

func genTestingRespParser(fail bool) module.ParseResponse {
	if fail {
		return func(resp *data.Response) (data []data.Data, parseErrors []*constant.YiError) {
			errs := []*constant.YiError{constant.NewYiErrorf(constant.ERR_CRAWL_ANALYZER,
				"Fail!(httpResp:%#v, respDepth:%#v)", resp.HTTPResp(), resp.Depth())}
			return nil, errs
		}
	}
	return func(resp *data.Response) (mdata []data.Data, parseErrors []*constant.YiError) {
		httpResp := resp.HTTPResp()
		respDepth := resp.Depth()
		mdata = []data.Data{}
		parseErrors = []*constant.YiError{}
		item := data.Item(map[string]interface{}{})
		item["url"] = httpResp.Request.URL.String()
		bufReader := bufio.NewReader(httpResp.Body)
		line, _, err := bufReader.ReadLine()
		if err != nil {
			parseErrors = append(parseErrors, constant.NewYiErrore(constant.ERR_CRAWL_ANALYZER, err))
			return
		}
		lineStr := string(line)
		begin := strings.LastIndex(lineStr, "[")
		end := strings.LastIndex(lineStr, "]")
		if begin < 0 ||
			end < 0 ||
			begin > end {
			err := fmt.Errorf("Wrong index for index: %d, %d", begin, end)
			parseErrors = append(parseErrors, constant.NewYiErrore(constant.ERR_CRAWL_ANALYZER, err))
			return
		}
		index, err := strconv.Atoi(lineStr[begin+1 : end])
		if err != nil {
			parseErrors = append(parseErrors, constant.NewYiErrore(constant.ERR_CRAWL_ANALYZER, err))
			return
		}
		item["index"] = index
		item["depth"] = respDepth
		mdata = append(mdata, item)
		req := data.NewRequest(nil)
		mdata = append(mdata, req)
		return
	}
}

var fakeHTTPRespBody = "Fake HTTP Response [%d]"

/*
 * create instances of response for testing
 */
func getTestingResps(
	number uint32,
	method string,
	url string,
	depth uint32,
	t *testing.T) []*data.Response {
	httpReq, err := http.NewRequest(method, url, nil)
	if err != nil {
		t.Fatalf("An error occurs when creating a HTTP request: %s (method: %s, url: %s)",
			err, method, url)
	}
	resps := []*data.Response{}
	for i := uint32(0); i < number; i++ {
		httpResp := &http.Response{
			Request: httpReq,
			Body: testingReader{
				strings.NewReader(
					fmt.Sprintf(fakeHTTPRespBody, i))},
		}
		req := data.NewRequest(httpReq)
		req.SetDepth(depth)
		resp := data.NewResponse(req, httpResp)
		resps = append(resps, resp)
	}
	return resps
}
