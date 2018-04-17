package module

import (
	"math"
	"sync"
)

// default serial number generator
var DefaultSNGen = NewSNGenerator(1, 0)

type SNGenerator interface {
	Start() uint64      // get the minimum serial number
	Max() uint64        // get the maximum serial number
	Next() uint64       // get next serial number
	CycleCount() uint64 // get the cycle count
	Get() uint64        // get a serial number and prepare the next one
}

/*
 * implementation of interface SNGenerator
 */
type mySNGenerator struct {
	start      uint64       // the minimum serial number
	max        uint64       // the maximum serial number
	next       uint64       // next setial number
	cycleCount uint64       // cycle count
	lock       sync.RWMutex // read/write lock
}

/*
 * create a sn generator according to the minimum serial number and the maximum serial number
 */
func NewSNGenerator(start, max uint64) SNGenerator {
	if max == 0 {
		max = math.MaxUint64
	}
	if start > max {
		start = max
	}
	return &mySNGenerator{
		start: start,
		max:   max,
		next:  start,
	}
}

/*
 * get the minimum serial number
 */
func (gen *mySNGenerator) Start() uint64 {
	return gen.start
}

/*
 * get the maximum serial number
 */
func (gen *mySNGenerator) Max() uint64 {
	return gen.max
}

/*
 * get next serial number
 */
func (gen *mySNGenerator) Next() uint64 {
	gen.lock.RLock()
	defer gen.lock.RUnlock()
	return gen.next
}

/*
 * get the cycle count
 */
func (gen *mySNGenerator) CycleCount() uint64 {
	gen.lock.RLock()
	defer gen.lock.RUnlock()
	return gen.cycleCount
}

/*
 * get a serial number and prepare the next one
 */
func (gen *mySNGenerator) Get() uint64 {
	gen.lock.Lock()
	defer gen.lock.Unlock()
	id := gen.next
	//cycle
	if id == gen.max {
		gen.next = gen.start
		gen.cycleCount++
	} else {
		gen.next++
	}
	return id
}
