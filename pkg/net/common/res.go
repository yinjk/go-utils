/*
 @Desc

 @Date 2020-04-17 15:17
 @Author yinjk
*/
package common

import (
	"fmt"
)

const SUCCESS = 0

//Result encapsulation of all return message, code is the status code, if success,
// the status code is 0, if fail, may be other
type Result struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"response"`
}

//IsSuccess is this result.code equals 0
func (r *Result) IsSuccess() bool {
	return r.Code == SUCCESS
}

/**
 * 返回一个成功的结果
 * return a success result to solace
 *
 * @param data: the data who is into Result struct
 * @return: The pointer for Result struct
 */
func NewSuccessResult(data interface{}, msg ...interface{}) *Result {
	message := "Done"
	if len(msg) > 0 {
		message = fmt.Sprint(msg...)
	}
	return &Result{
		Code:    SUCCESS,
		Data:    data,
		Message: message,
	}
}

/**
 * 返回一个失败的结果
 * returns a failed result to solace
 *
 * @param errorCode: the code of err
 * @param msg: the description for this failed call
 * @return: The pointer for Result struct
 */
func NewFailResult(errorCode int, msg string) *Result {
	return &Result{
		Code:    errorCode,
		Data:    nil,
		Message: msg,
	}
}
