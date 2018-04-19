package main

import (
	"crypto/md5"
	"github.com/l-dandelion/yi-ants-go/core/module/data"
	"github.com/l-dandelion/yi-ants-go/lib/constant"
	"io"
	"github.com/l-dandelion/yi-ants-go/lib/utils"
	"fmt"
)

func process(item data.Item) (data.Item, *constant.YiError) {
	dirPath := item["DirPath"].(string)
	murl := item["URL"].(string)
	h := md5.New()
	h.Write([]byte(murl))
	etx := item["Etx"].(string)
	fileName := fmt.Sprintf("%x", h.Sum(nil)) + etx
	reader := item["Reader"].(io.Reader)
	err := utils.SaveFileByReader(dirPath, fileName, reader)
	if err != nil {
		return nil, constant.NewYiErrore(constant.ERR_CRAWL_PIPELINE, err)
	}
	return nil, nil
}