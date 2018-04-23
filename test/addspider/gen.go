package main

import (
	"github.com/l-dandelion/yi-ants-go/core/module"
	"github.com/l-dandelion/yi-ants-go/core/scheduler"
	"github.com/l-dandelion/yi-ants-go/core/spider"
	"github.com/l-dandelion/yi-ants-go/lib/constant"
)

func genSpider() (spider.Spider, *constant.YiError){
	requestArgs := scheduler.RequestArgs{
		MaxDepth:        3,
		AcceptedDomains: []string{"pixabay.com"},
	}
	dataArgs := scheduler.DataArgs{
		ReqBufferCap:         50,
		ReqMaxBufferNumber:   1000,
		RespBufferCap:        50,
		RespMaxBufferNumber:  100,
		ItemBufferCap:        50,
		ItemMaxBufferNumber:  1000,
		ErrorBufferCap:       50,
		ErrorMaxBufferNumber: 1,
	}
	parsers := []module.ParseResponse{parseATag, parseImgTag}
	processors := []module.ProcessItem{process}

	sp, yierr := spider.New("test",
		requestArgs,
		dataArgs,
		[]string{"https://pixabay.com"},
		nil,
		parsers,
		processors)
	return sp, yierr
}
