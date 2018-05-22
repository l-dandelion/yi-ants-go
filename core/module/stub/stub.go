package stub

import (
	"github.com/l-dandelion/yi-ants-go/core/module"
	"github.com/l-dandelion/yi-ants-go/lib/constant"
	"sync/atomic"
)

/*
 * implementation of interface ModuleInternal
 */
type myModule struct {
	score           uint64                // score of module
	calledCount     uint64                // called count
	acceptedCount   uint64                // accepted count
	completedCount  uint64                // completed count
	handlingNumber  uint64                // handling number
	mid             module.MID            // id of module
	addr            string                // network address of module
	scoreCalculator module.CalculateScore // score calculator
}

/*
 * create an instance of ModuleInternal
 */
func NewModuleInternal(
	mid module.MID,
	scoreCalculator module.CalculateScore) (mi ModuleInternal, yierr *constant.YiError) {
	parts, yierr := module.SplitMID(mid)
	if yierr != nil {
		return
	}
	return &myModule{
		mid:             mid,
		addr:            parts[2],
		scoreCalculator: scoreCalculator,
	}, nil
}

/*
 * get mid of module
 */
func (m *myModule) ID() module.MID {
	return m.mid
}

/*
 * get network address of module(ip:port)
 */
func (m *myModule) Addr() string {
	return m.addr
}

/*
 * get score of module, used for load balancing
 */
func (m *myModule) Score() uint64 {
	return m.score
}

/*
* set score of module, used for load balancing
 */
func (m *myModule) SetScore(score uint64) {
	atomic.StoreUint64(&m.score, score)
}

/*
 * get score calculator of module, used for load balancing
 */
func (m *myModule) ScoreCalculator() module.CalculateScore {
	return m.scoreCalculator
}

/*
 * get the called count of module (concurrent security)
 */
func (m *myModule) CalledCount() uint64 {
	return atomic.LoadUint64(&m.calledCount)
}

/*
 * get the accepted count of module (concurrent security)
 */
func (m *myModule) AcceptedCount() uint64 {
	return atomic.LoadUint64(&m.acceptedCount)
}

/*
 * get the completed count of module (concurrent security)
 */
func (m *myModule) CompletedCount() uint64 {
	return atomic.LoadUint64(&m.completedCount)
}

/*
 * get the handling number of module (concurrent security)
 */
func (m *myModule) HandlingNumber() uint64 {
	return atomic.LoadUint64(&m.handlingNumber)
}

/*
 * get the counts of module (concurrent security)
 */
func (m *myModule) Counts() module.Counts {
	return module.Counts{
		CalledCount:    m.CalledCount(),
		AcceptedCount:  m.AcceptedCount(),
		CompletedCount: m.CompletedCount(),
		HandlingNumber: m.HandlingNumber(),
	}
}

/*
 * get the summary of module (concurrent security)
 */
func (m *myModule) Summary() module.SummaryStruct {
	counts := m.Counts()
	return module.SummaryStruct{
		ID:        m.ID(),
		Called:    counts.CalledCount,
		Accepted:  counts.AcceptedCount,
		Completed: counts.CompletedCount,
		Handling:  counts.HandlingNumber,
		Extra:     nil,
	}
}

/*
 * increase the called count by one (concurrent security)
 */
func (m *myModule) IncrCalledCount() {
	atomic.AddUint64(&m.calledCount, 1)
}

/*
 * increase the accepted count by one (concurrent security)
 */
func (m *myModule) IncrAcceptedCount() {
	atomic.AddUint64(&m.acceptedCount, 1)
}

/*
 * increase the completed count by one (concurrent security)
 */
func (m *myModule) IncrCompletedCount() {
	atomic.AddUint64(&m.completedCount, 1)
}

/*
 * increase the handling number by one (concurrent security)
 */
func (m *myModule) IncrHandlingNumber() {
	atomic.AddUint64(&m.handlingNumber, 1)
}

/*
 * increase the handling number by one (concurrent security)
 */
func (m *myModule) DecrHandlingNumber() {
	atomic.AddUint64(&m.handlingNumber, ^uint64(0))
}

/*
 * clear all counts (concurrent security)
 */
func (m *myModule) Clear() {
	atomic.StoreUint64(&m.calledCount, 0)
	atomic.StoreUint64(&m.acceptedCount, 0)
	atomic.StoreUint64(&m.completedCount, 0)
	atomic.StoreUint64(&m.handlingNumber, 0)
}
