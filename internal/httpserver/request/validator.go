package request

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/salivare/subscriptions-service/internal/httpserver/response"
)

func ValidationError(errs validator.ValidationErrors) response.Response {
	var errMsgs []string

	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is a required field", err.Field()))
		default:
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is not valid", err.Field()))
		}
	}

	return response.Response{
		Status: response.StatusError,
		Error:  strings.Join(errMsgs, ", "),
		Code:   http.StatusBadRequest,
	}
}
