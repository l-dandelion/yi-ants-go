package utils

import "strings"

func GetNoDomainUri(uri string) string {
	if strings.HasPrefix(uri, "http") || strings.HasPrefix(uri, "//") {
		pos := strings.Index(uri, "//")
		if pos > 0 {
			uri = string(uri[pos+2:])
		}
		pos = strings.Index(uri, "/")
		if pos > 0 {
			uri = string(uri[pos:])
		}
	}
	return uri
}

func GetControllerInfoFromUri(uri string) (isApi bool, cName string, aName string) {
	cName = "home"
	aName = "index"

	if uri == "" {
		return
	}
	uri = GetNoDomainUri(uri)
	turi := strings.Split(strings.TrimPrefix(uri, "/"), "?")
	arrUri := strings.Split(turi[0], "/")
	tlen := len(arrUri)
	cur := 0
	if tlen > 1 && arrUri[cur] == "api" {
		isApi = true
		cur = cur + 1
	}
	if tlen > cur && arrUri[cur] != "" {
		cName = strings.ToLower(arrUri[cur])
		cur = cur + 1
	}
	if tlen > cur && arrUri[cur] != "" {
		aName = strings.ToLower(arrUri[cur])
	}
	return
}