package base

import (
	"github.com/l-dandelion/yi-ants-go/lib/common"
	"sync"
)

var (
	inputDataPool sync.Pool
)

func init() {
	inputDataPool.New = func() interface{} {
		return &common.InputData{}
	}
}
