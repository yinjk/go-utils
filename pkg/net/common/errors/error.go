//@Desc
//@Date 2019-11-21 14:00
//@Author yinjk
package errors

import "errors"

var (
	DBError        = errors.New("database is error")
	UnSupportError = errors.New("this feature is not supported temporarily")
)
