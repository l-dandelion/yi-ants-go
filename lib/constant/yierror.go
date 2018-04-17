package constant

import (
	"errors"
	"fmt"
)

const ERROR_TEMPLATE = "errNo: %d errMsg: %s errDesc: %s"

/*
 * Implement error type
 * 实现错误类型
 */
type YiError struct {
	ErrNo   int    //error number(错误号)
	ErrMsg  string //error message(错误信息)
	ErrDesc string //error description(错误描述)
	Err     error  //new by error message(由错误信息产生)
}

/*
 * Implement error type
 */
func (myerr *YiError) Error() string {
	return fmt.Sprintf(ERROR_TEMPLATE, myerr.ErrNo, myerr.ErrMsg, myerr.ErrDesc)
}

/*
 * New YiError by error number and error
 */
func NewYiErrore(errno int, err error, args ...interface{}) (e *YiError) {
	e = &YiError{}
	e.ErrNo = errno
	if len(args) > 0 {
		e.ErrMsg = fmt.Sprintf(GetErrMsg(errno), args...)
	} else {
		e.ErrMsg = GetErrMsg(errno)
	}
	e.ErrDesc = err.Error()
	e.Err = err
	return
}

/*
 * New YiError by error number and error description
 */
func NewYiErrorf(errno int, errdesc string, args ...interface{}) (e *YiError) {
	e = &YiError{}
	e.ErrNo = errno
	e.ErrMsg = GetErrMsg(errno)
	e.ErrDesc = fmt.Sprintf(errdesc, args...)
	e.Err = errors.New(e.ErrDesc)
	return
}

/*
 * New YiError by error number, error message and error description
 */
func NewYiError(errno int, errmsg string, errdesc string) (e *YiError) {
	e = &YiError{}
	e.ErrNo = errno
	e.ErrMsg = errmsg
	e.ErrDesc = errdesc
	e.Err = errors.New(errmsg)
	return
}
