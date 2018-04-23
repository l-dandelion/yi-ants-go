package http

import (
	"github.com/l-dandelion/yi-ants-go/lib/constant"
)

// welcome struct
type WelcomeInfo struct {
	Message  string
	Greeting string
	Time     string
}

// result of start spider
type StartSpiderResult struct {
	Success    bool
	Yierr      *constant.YiError
	Spider     string
	Time       string
}
