package module

import (
	"github.com/l-dandelion/yi-ants-go/core/module/data"
	"github.com/l-dandelion/yi-ants-go/lib/constant"
)


/*
 * the type of internal counts of module.
 * include: called count, accepted count, completed count, handling number
 */
type Counts struct {
	CalledCount    uint64 // called count
	AcceptedCount  uint64 // accepted count
	CompletedCount uint64 // completed count
	HandlingNumber uint64 // handling number
}

/*
 * the struct of summary of module
 */
type SummaryStruct struct {
	ID        MID         `json:"id"`              // unique id
	Called    uint64      `json:"called"`          // called count
	Accepted  uint64      `json:"accepted"`        // accepted count
	Completed uint64      `json:"completed"`       // completed count
	Handling  uint64      `json:"handling"`        // handling number
	Extra     interface{} `json:"extra,omitempty"` // extra information
}

/*
 * basic interface representing module
 * the implementation type of the interface must be concurrent and secure.
 */
type Module interface {
	ID() MID                         // get the unique id of the module
	Addr() string                    // get the network address of the module (addr:port)
	Score() uint64                   // get the score of the module
	SetScore(score uint64)           // set the score of the module
	ScoreCalculator() CalculateScore // get the score calculator
	CalledCount() uint64             // get the called count of the module
	AcceptedCount() uint64           // get the accepted count of the module
	CompletedCount() uint64          // get the completed count of the module
	HandlingNumber() uint64          // get the handling number of the module
	Counts() Counts                  // get the counts of the module
	Summary() SummaryStruct          // get the struct of summary of the module
}

/*
 * interface for the downloader
 * inherit from module
 * the implementation type of the interface must be concurrent and secure.
 */
type Downloader interface {
	Module                                              // inherit from module
	Download(req *data.Request) (*data.Response, *constant.YiError) // download according to the request and return the response
	Add()
	Done()
}

/*
 * interface for the analyzer
 * inherit from module
 * the implementation type of the interface must be concurrent and secure.
 */
type Analyzer interface {
	Module                                              // inherit from module
	RespParsers() []ParseResponse                       // get responser parsers
	Analyze(resp *data.Response) ([]data.Data, []*constant.YiError) // analyze according to the rule and return the data
}

/*
 * function for parese response
 * httpResp: http response
 * extra: It is same as the extra of request. has extra["proxy"], extra["depth"], and etc.
 */
type ParseResponse func(resp *data.Response) ([]data.Data, []*constant.YiError)

/*
 * interface for the pipeline
 * inherit from module
 * the implementation type of the interface must be concurrent and secure.
 */
type Pipeline interface {
	Module
	ItemProcessors() []ProcessItem // Get item processors
	Send(item data.Item) []*constant.YiError   // Send item to the pipeline, and process by item processors one by one

	/*
	 * This value indicates whether the pipeline is fast failure
	 * if true, it will immediately stop parsing this item and report the errors when errors occur
	 */
	FailFast() bool
	SetFailFast(failFast bool) // Set fail fast
}

/*
 * func for process the item
 */
type ProcessItem func(item data.Item) (result data.Item, yierr *constant.YiError)
