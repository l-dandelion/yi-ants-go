package utils

import (
	"net/url"
)

func ParseURL(href string) (*url.URL, error) {
	return url.Parse(href)
}

func GetComplateUrl(reqURL *url.URL, href string) (string, error) {
	aURL, err := url.Parse(href)
	if err != nil {
		return "", err
	}
	if !aURL.IsAbs() {
		aURL = reqURL.ResolveReference(aURL)
	}
	return aURL.String(), nil
}
