package constant

import (
	"fmt"
	"testing"
)

func TestNewYiError(t *testing.T) {
	var (
		errno    int
		errmsg   string
		errdesc  string
		myerr    *YiError
		expected string
		err      error
	)
	errno = ERR_CRAWL_DOWNLOADER
	errmsg = "downloader error"
	errdesc = "download fail"
	myerr = NewYiError(errno, errmsg, errdesc)
	expected = fmt.Sprintf(ERROR_TEMPLATE, errno, errmsg, errdesc)
	if myerr.Error() != expected {
		t.Fatalf("ERROR: NewYiError. expected: %s, actual: %s", expected, myerr)
	}

	errno = ERR_CRAWL_ANALYZER
	errdesc = "download fail"
	myerr = NewYiErrorf(errno, errdesc)
	expected = fmt.Sprintf(ERROR_TEMPLATE, errno, GetErrMsg(errno), errdesc)
	if myerr.Error() != expected {
		t.Fatalf("ERROR: NewYiErrorf. expected: %s, actual: %s", expected, myerr)
	}

	errno = ERR_CRAWL_ANALYZER
	err = fmt.Errorf("download fail")
	myerr = NewYiErrore(errno, err)
	expected = fmt.Sprintf(ERROR_TEMPLATE, errno, GetErrMsg(errno), errdesc)
	if myerr.Error() != expected {
		t.Fatalf("ERROR: NewYiErrore. expected: %s, actual: %s", expected, myerr)
	}
}
