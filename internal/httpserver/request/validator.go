package request

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/salivare/subscriptions-service/internal/httpserver/render"
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

var validate = validator.New()

func ValidateStruct(w http.ResponseWriter, r *http.Request, req interface{}) bool {
	if err := validate.Struct(req); err != nil {
		var vErr validator.ValidationErrors
		if errors.As(err, &vErr) {
			resp := ValidationError(vErr)
			render.JSON(w, r, resp)
			return false
		}

		render.JSON(w, r, response.Error("invalid request"))
		return false
	}

	return true
}
