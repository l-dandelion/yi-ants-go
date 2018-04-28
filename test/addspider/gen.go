package main

import (
	"github.com/l-dandelion/yi-ants-go/core/scheduler"
	"github.com/l-dandelion/yi-ants-go/core/spider"
	"github.com/l-dandelion/yi-ants-go/lib/constant"
	"io/ioutil"
	"os"
)

func ReadAll(filePth string) ([]byte, error) {
	f, err := os.Open(filePth)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(f)
}

func genSpider() (spider.Spider, *constant.YiError) {
	requestArgs := scheduler.RequestArgs{
		MaxDepth:        10,
		AcceptedDomains: []string{"pixabay.com"},
	}
	dataArgs := scheduler.DataArgs{
		ReqBufferCap:         1000,
		ReqMaxBufferNumber:   10000,
		RespBufferCap:        50,
		RespMaxBufferNumber:  100,
		ItemBufferCap:        50,
		ItemMaxBufferNumber:  1000,
		ErrorBufferCap:       50,
		ErrorMaxBufferNumber: 1,
	}
	byteGenParsers, err := ReadAll("./parser.go")
	if err != nil {
		return nil, constant.NewYiErrore(constant.ERR_SPIDER_NEW, err)
	}
	byteGenProcessors, err := ReadAll("./processor.go")
	if err != nil {
		return nil, constant.NewYiErrore(constant.ERR_SPIDER_NEW, err)
	}

	sp, yierr := spider.New("test",
		requestArgs,
		dataArgs,
		[]string{"https://pixabay.com"},
		nil,
		string(byteGenParsers),
		string(byteGenProcessors))
	return sp, yierr
}
