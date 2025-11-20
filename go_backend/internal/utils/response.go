package utils

// 统一响应格式

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// SuccessResponse 返回成功响应
func SuccessResponse(data interface{}) Response {
	return Response{
		Code:    200,
		Message: "success",
		Data:    data,
	}
}

// ErrorResponse 返回错误响应
func ErrorResponse(message string) Response {
	return Response{
		Code:    400,
		Message: message,
		Data:    nil,
	}
}