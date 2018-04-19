package utils

import "net/url"

func ParseURL(href string) (*url.URL, error) {
	return url.Parse(href)
}