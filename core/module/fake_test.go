package module

import (
	"sync/atomic"

	"github.com/l-dandelion/yi-ants-go/core/module/data"
	"github.com/l-dandelion/yi-ants-go/lib/constant"
)

var (
	// default fake downloader
	defaultFakeDownloader = NewFakeDownloader(MID("D0"), CalculateScoreSimple)

	// default fake analyzer
	defaultFakeAnalyzer = NewFakeAnalyzer(MID("A1"), CalculateScoreSimple)

	// default fake pipeline
	defaultFakePipeline = NewFakePipeline(MID("P2"), CalculateScoreSimple)

	// fake modules
	fakeModules = []Module{
		defaultFakeDownloader,
		defaultFakeAnalyzer,
		defaultFakePipeline,
	}

	// default fake module map
	defaultFakeModuleMap = map[int8]Module{
		TYPE_DOWNLOADER: defaultFakeDownloader,
		TYPE_ANALYZER:   defaultFakeAnalyzer,
		TYPE_PIPELINE:   defaultFakePipeline,
	}

	//fake module function map
	fakeModuleFuncMap = map[int8]func(mid MID) Module{
		TYPE_DOWNLOADER: func(mid MID) Module {
			return NewFakeDownloader(mid, CalculateScoreSimple)
		},
		TYPE_ANALYZER: func(mid MID) Module {
			return NewFakeAnalyzer(mid, CalculateScoreSimple)
		},
		TYPE_PIPELINE: func(mid MID) Module {
			return NewFakePipeline(mid, CalculateScoreSimple)
		},
	}
)

// fake module
type fakeModule struct {
	mid             MID            // the id of module
	score           uint64         // the score of module
	count           uint64         // the count of module
	scoreCalculator CalculateScore // score calculator
}

/*
 * get MID of module
 */
func (fm *fakeModule) ID() MID {
	return fm.mid
}

/*
 * get network address of module
 */
func (fm *fakeModule) Addr() string {
	parts, err := SplitMID(fm.mid)
	if err == nil {
		return parts[2]
	}
	return ""
}

/*
 * get score of module
 */
func (fm *fakeModule) Score() uint64 {
	return atomic.LoadUint64(&fm.score)
}

/*
 * set score for module
 */
func (fm *fakeModule) SetScore(score uint64) {
	atomic.StoreUint64(&fm.score, score)
}

/*
 * get the score calculator of module
 */
func (fm *fakeModule) ScoreCalculator() CalculateScore {
	return fm.scoreCalculator
}

/*
 * get called count
 */
func (fm *fakeModule) CalledCount() uint64 {
	return fm.count + 10
}

/*
 * get accepted count
 */
func (fm *fakeModule) AcceptedCount() uint64 {
	return fm.count + 8
}

/*
 * get completed count
 */
func (fm *fakeModule) CompletedCount() uint64 {
	return fm.count + 6
}

/*
 * get handling number
 */
func (fm *fakeModule) HandlingNumber() uint64 {
	return fm.count + 2
}

/*
 * get counts of module
 */
func (fm *fakeModule) Counts() Counts {
	return Counts{
		fm.CalledCount(),
		fm.AcceptedCount(),
		fm.CompletedCount(),
		fm.HandlingNumber(),
	}
}

/*
 * get summary of module
 */
func (fm *fakeModule) Summary() SummaryStruct {
	return SummaryStruct{}
}

/*
 * create an instance for fake analyzer
 */
func NewFakeAnalyzer(mid MID, scoreCalculator CalculateScore) Analyzer {
	return &fakeAnalyzer{
		fakeModule: fakeModule{
			mid:             mid,
			scoreCalculator: scoreCalculator,
		},
	}
}

/*
 * fake analyzer
 */
type fakeAnalyzer struct {
	fakeModule //fake module
}

/*
 * (fake)the function to generate response parsers
 */
func (analyzer *fakeAnalyzer) RespParsers() []ParseResponse {
	return nil
}

/*
 * (fake)the function to analyze
 */
func (analyzer *fakeAnalyzer) Analyze(resp *data.Response) (dataList []data.Data, errorList []*constant.YiError) {
	return
}

/*
 * create an instance for fake downloader
 */
func NewFakeDownloader(mid MID, scoreCalculator CalculateScore) Downloader {
	return &fakeDownloader{
		fakeModule: fakeModule{
			mid:             mid,
			scoreCalculator: scoreCalculator,
		},
	}
}

// fake downloader
type fakeDownloader struct {
	fakeModule // fake module
}

/*
 * (fake)the function to download
 */
func (downloader *fakeDownloader) Download(req *data.Request) (*data.Response, *constant.YiError) {
	return nil, nil
}

/*
 * create an instance for pipeline
 */
func NewFakePipeline(mid MID, scoreCalculator CalculateScore) Pipeline {
	return &fakePipeline{
		fakeModule: fakeModule{
			mid:             mid,
			scoreCalculator: scoreCalculator,
		},
	}
}

// fake pipeline
type fakePipeline struct {
	fakeModule      //fake module
	failFast   bool // fail fast
}

/*
 * (fake)the function to generate item processors
 */
func (pipeline *fakePipeline) ItemProcessors() []ProcessItem {
	return nil
}

/*
 * (fake)the function to process item
 */
func (pipeline *fakePipeline) Send(item data.Item) []*constant.YiError {
	return nil
}

/*
 * the function to check whether the pipeline is fast fail
 */
func (pipeline *fakePipeline) FailFast() bool {
	return pipeline.failFast
}

/*
 * the function to set whether the pipeline is fast fail
 */
func (pipeline *fakePipeline) SetFailFast(failFast bool) {
	pipeline.failFast = failFast
}
