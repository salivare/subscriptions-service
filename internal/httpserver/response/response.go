package response

import "net/http"

const (
	StatusOK    = "OK"
	StatusError = "Error"
)

type Response struct {
	Status string      `json:"status"`
	Error  interface{} `json:"error,omitempty"`
	Code   int         `json:"-"`
}

func (r Response) StatusCode() int {
	if r.Code == 0 {
		return 200
	}

	return r.Code
}

func OK() Response {
	return Response{
		Status: StatusOK,
		Code:   http.StatusOK,
	}
}

func Error(msg string) Response {
	return Response{
		Status: StatusError,
		Error:  msg,
		Code:   http.StatusBadRequest,
	}
}

func Conflict(msg string) Response {
	return Response{
		Status: StatusError,
		Error:  msg,
		Code:   http.StatusConflict,
	}
}

func Internal(msg string) Response {
	return Response{
		Status: StatusError,
		Error:  msg,
		Code:   http.StatusInternalServerError,
	}
}
