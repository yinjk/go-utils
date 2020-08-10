/*
@Desc:
	All status code definitions in this file for this system.

@Date 2019/10/14
@Author yinjk
*/
package code

var (
	Success                = 0
	ErrorCodeOffset        = 250000
	UnCaughtExceptionError = ErrorCodeOffset + 1 // 未捕获异常错误

	CommonError = ErrorCodeOffset // 一般错误起始编码，比如参数错误等

	// 一般错误起始编码，比如参数错误等
	InvalidParameter    = CommonError + 1       // 无效的参数
	BadRequest          = ErrorCodeOffset + 400 // 提交参数解析错误
	Unauthorized        = ErrorCodeOffset + 401 // 未授权
	InternalServerError = ErrorCodeOffset + 500 // 服务器内部错误
)
