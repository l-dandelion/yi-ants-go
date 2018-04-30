package main

import (
	"net/http"
	"testing"
	"github.com/l-dandelion/yi-ants-go/core/module/local/downloader"
	"github.com/l-dandelion/yi-ants-go/core/module"
	"github.com/l-dandelion/yi-ants-go/core/module/data"
	"fmt"
)

func TestDownloader(t *testing.T) {
	downloader, yierr := downloader.New(module.MID("D1"), genHTTPClient(), module.CalculateScoreSimple)
	if yierr != nil {
		t.Fatal(yierr)
	}
	httpReq, err := http.NewRequest("GET", "https://500px.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	req := data.NewRequest(httpReq)
	resp, yierr := downloader.Download(req)
	if yierr != nil {
		t.Fatal(yierr)
	}
	text, err := resp.GetText()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(text))
}