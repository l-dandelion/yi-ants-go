package scheduler

import (
	"github.com/l-dandelion/yi-ants-go/lib/constant"
	"github.com/l-dandelion/yi-ants-go/lib/library/pool"
)

var (
	downloaderPool = pool.NewPool(constant.MaxThread)
	analyzerPool   = pool.NewPool(constant.MaxThread)
	pipelinePool   = pool.NewPool(constant.MaxThread)
)
