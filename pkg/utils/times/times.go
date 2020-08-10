package times

import (
	"github.com/pkg/errors"
	"time"
)

var (
	ErrorTimeOut   = errors.New("time out with out result!")
	DateTimeFormat = "2006-01-02 15:04:05"
	DateFormat     = "2006-01-02"
	TimeFormat     = "15:04:05"
)

//TimeOutWithResult 超时机制，如果defaultResult为空，超时会返回ErrorTimeOut错误，如果defaultResult不为空，超时会返回defaultResult
func TimeOutWithResult(provide func() (result interface{}, err error), timeout time.Duration, defaultResult interface{}) (result interface{}, err error) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	resultCh := make(chan interface{}, 1)
	errCh := make(chan error, 1)
	go func(resultCh chan interface{}, errCh chan error) {
		if result, err = provide(); err != nil {
			errCh <- err
		} else {
			resultCh <- result
		}
		close(resultCh)
		close(errCh)
	}(resultCh, errCh)

	select {
	case <-timer.C:
		//time out, return the defaultResult
		if defaultResult != nil {
			return defaultResult, nil
		}
		return nil, ErrorTimeOut
	case result = <-resultCh:
		//the normal exec, and return the real result
		return
	case err = <-errCh:
		//the normal exec, and return the errors
		return
	}
}

//TimeOutFunc 超时机制，通过timeout表示多少时间超时，超时会返回一个ErrorTimeOut错误
func TimeOutFunc(provide func() (err error), timeout time.Duration) (err error) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	errCh := make(chan error, 1)
	go func(errCh chan error) {
		errCh <- provide()
		close(errCh)
	}(errCh)

	select {
	case <-timer.C:
		//time out, return the defaultResult
		return ErrorTimeOut
	case err = <-errCh:
		//the normal exec and return
		return
	}
}
