package main

import (
	"crypto/md5"
	"fmt"
	"github.com/l-dandelion/yi-ants-go/core/module"
	"github.com/l-dandelion/yi-ants-go/core/module/data"
	"github.com/l-dandelion/yi-ants-go/lib/constant"
	"github.com/l-dandelion/yi-ants-go/lib/utils"
	"io"
)

func process(item data.Item) (data.Item, *constant.YiError) {
	dirPath := item["DirPath"].(string)
	murl := item["URL"].(string)
	h := md5.New()
	h.Write([]byte(murl))
	etx := item["Etx"].(string)
	sum := h.Sum(nil)
	var fileName string
	if etx == ".html" {
		dirPath = dirPath + "/" + fmt.Sprintf("%x", sum)[0:3]
	} else {
		dirPath = dirPath + "/" + fmt.Sprintf("%x", sum)[0:2]
	}
	fileName = fmt.Sprintf("%x", sum) + etx
	fileName = fmt.Sprintf("%x", sum) + etx
	reader := item["Reader"].(io.Reader)
	err := utils.SaveFileByReader(dirPath, fileName, reader)
	if err != nil {
		return nil, constant.NewYiErrore(constant.ERR_CRAWL_PIPELINE, err)
	}
	return nil, nil
}

func GenProcessors() []module.ProcessItem {
	return []module.ProcessItem{process}
}
