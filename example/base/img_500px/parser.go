package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/l-dandelion/yi-ants-go/core/module/data"
	"github.com/l-dandelion/yi-ants-go/lib/constant"
	"github.com/l-dandelion/yi-ants-go/lib/utils"
)

// function for parsing response
func parseATag(resp *data.Response) ([]data.Data, []*constant.YiError) {
	reqURL := resp.HTTPResp().Request.URL
	httpResp := resp.HTTPResp()
	//TODO: 支持更多的HTTP响应状态。
	if httpResp.StatusCode != 200 {
		err := fmt.Errorf(
			fmt.Sprintf("Unsupported status code %d! (httpResponse: %v)",
				httpResp.StatusCode, httpResp))
		return nil, []*constant.YiError{constant.NewYiErrore(constant.ERR_CRAWL_ANALYZER, err)}
	}
	dom, err := resp.GetDom()
	if err != nil {
		return nil, []*constant.YiError{constant.NewYiErrore(constant.ERR_CRAWL_ANALYZER, err)}
	}
	dataList := []data.Data{}
	yierrList := []*constant.YiError{}
	dom.Find("a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists || href == "" || href == "#" || href == "/" {
			return
		}
		trimHref := strings.TrimSpace(href)
		lowHref := strings.ToLower(trimHref)
		fdStart := strings.Index(lowHref, "javascript")
		if fdStart == 0 {
			return
		}
		aURL, err := utils.ParseURL(lowHref)
		if err != nil {
			yierrList = append(yierrList, constant.NewYiErrore(constant.ERR_CRAWL_ANALYZER, err))
			return
		}
		if !aURL.IsAbs() {
			aURL = reqURL.ResolveReference(aURL)
		}
		httpReq, err := http.NewRequest("GET", aURL.String(), nil)
		if err != nil {
			yierrList = append(yierrList, constant.NewYiErrore(constant.ERR_CRAWL_ANALYZER, err))
			return
		}
		req := data.NewRequest(httpReq)
		dataList = append(dataList, req)
	})
	item := data.Item{
		"URL":     reqURL.String(),
		"DirPath": "result/",
		"Reader":  resp.HTTPResp().Body,
		"Etx":     ".html",
	}
	dataList = append(dataList, item)
	return dataList, yierrList
}

func parseATag2(resp *data.Response) ([]data.Data, []*constant.YiError) {
	matchedContentType := false
	httpResp := resp.HTTPResp()
	reqURL := httpResp.Request.URL
	if httpResp.Header != nil {
		contentTypes := httpResp.Header["Content-Type"]
		for _, contentType := range contentTypes {
			if strings.Index(contentType, "text/html") == 0 {
				matchedContentType = true
				break
			}
		}
	}
	dataList := []data.Data{}
	yierrList := []*constant.YiError{}
	if matchedContentType {
		dom, err := resp.GetDom()
		if err != nil {
			return nil, []*constant.YiError{constant.NewYiErrore(constant.ERR_CRAWL_ANALYZER, err)}
		}
		dom.Find("a").Each(func(i int, s *goquery.Selection) {
			href, exists := s.Attr("href")
			if !exists || href == "" || href == "#" || href == "/" {
				return
			}
			trimHref := strings.TrimSpace(href)
			lowHref := strings.ToLower(trimHref)
			fdStart := strings.Index(lowHref, "javascript")
			if fdStart == 0 {
				return
			}
			aURL, err := utils.ParseURL(lowHref)
			if err != nil {
				yierrList = append(yierrList, constant.NewYiErrore(constant.ERR_CRAWL_ANALYZER, err))
				return
			}
			if !aURL.IsAbs() {
				aURL = reqURL.ResolveReference(aURL)
			}
			httpReq, err := http.NewRequest("GET", aURL.String(), nil)
			if err != nil {
				yierrList = append(yierrList, constant.NewYiErrore(constant.ERR_CRAWL_ANALYZER, err))
				return
			}
			req := data.NewRequest(httpReq)
			dataList = append(dataList, req)
		})
		dom.Find("img").Each(func(i int, s *goquery.Selection) {
			href, exists := s.Attr("src")
			if !exists || href == "" || href == "#" || href == "/" {
				return
			}
			trimHref := strings.TrimSpace(href)
			lowHref := strings.ToLower(trimHref)
			fdStart := strings.Index(lowHref, "javascript")
			if fdStart == 0 {
				return
			}
			aURL, err := utils.ParseURL(lowHref)
			if err != nil {
				yierrList = append(yierrList, constant.NewYiErrore(constant.ERR_CRAWL_ANALYZER, err))
				return
			}
			if !aURL.IsAbs() {
				aURL = reqURL.ResolveReference(aURL)
			}
			httpReq, err := http.NewRequest("GET", aURL.String(), nil)
			if err != nil {
				yierrList = append(yierrList, constant.NewYiErrore(constant.ERR_CRAWL_ANALYZER, err))
				return
			}
			req := data.NewRequest(httpReq)
			dataList = append(dataList, req)
		})

		item := data.Item{
			"URL":     reqURL.String(),
			"DirPath": "temp/500px/htmls",
			"Reader":  resp.HTTPResp().Body,
			"Etx":     ".html",
		}
		dataList = append(dataList, item)
	}
	return dataList, yierrList
}

func parseImgTag(resp *data.Response) ([]data.Data, []*constant.YiError) {
	pictureFormat := ""
	httpResp := resp.HTTPResp()
	if httpResp.Header != nil {
		contentTypes := httpResp.Header["Content-Type"]
		contentType := ""
		for _, ct := range contentTypes {
			if strings.Index(ct, "image") == 0 {
				contentType = ct
				break
			}
		}

		index1 := strings.Index(contentType, "/")
		index2 := strings.Index(contentType, ";")
		if index1 > 0 {
			if index2 < 0 {
				pictureFormat = contentType[index1+1:]
			} else if index1 < index2 {
				pictureFormat = contentType[index1+1 : index2]
			}
		}
	}
	dataList := []data.Data{}
	yierrList := []*constant.YiError{}
	if pictureFormat != "" {
		reqURL := resp.HTTPResp().Request.URL
		item := data.Item{
			"URL":     reqURL.String(),
			"DirPath": "temp/500px/imgs/",
			"Reader":  resp.HTTPResp().Body,
			"Etx":     "." + pictureFormat,
		}
		dataList = append(dataList, item)
	}
	return dataList, yierrList
}
