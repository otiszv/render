package common

import (
	"fmt"
)

type Error struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
	Err     error                  `json::"OriginalError"`
}

const (
	codeInfoLossingError    = "InfoLossingError"
	codeValidateError       = "ValidateError"
	codeTemplateError       = "TemplateDefinitionError"
	codeTemplateRenderError = "TemplateRenderError"
)

func (err Error) Error() string {
	return fmt.Sprintf("%s:%s, data=%v", err.Code, err.Message, err.Data)
}

func NewInfoLossingError(message string, data map[string]interface{}) error {

	if message == "" {
		message = "info lossing error"
	}

	return Error{
		Code:    codeInfoLossingError,
		Message: message,
		Data:    data,
	}
}

func NewValidateError(message string, data map[string]interface{}) error {

	if message == "" {
		message = "validate error"
	}

	return Error{
		Code:    codeValidateError,
		Message: message,
		Data:    data,
	}
}

func NewTemplateDefinitionError(message string, data map[string]interface{}) error {

	if message == "" {
		message = "template format error"
	}

	return Error{
		Code:    codeTemplateError,
		Message: message,
		Data:    data,
	}
}

func NewTemplateRenderError(message string, oriErr error, data map[string]interface{}) error {

	if message == "" {
		message = "template render error"
	}

	return Error{
		Code:    codeTemplateRenderError,
		Message: message,
		Data:    data,
		Err:     oriErr,
	}
}

// multi errors
type Errors []error

func (errs Errors) Error() string {
	ct := ""
	for _, err := range errs {
		ct += fmt.Sprintf("%s\n", err)
	}
	return fmt.Sprintf("%s, errors=%s", "multi errors", ct)
}

//IsTemplateDefinitionError help you judge err type
func IsTemplateDefinitionError(err error) bool {
	return isThatError(err, codeTemplateError)
}

//IsTemplateRenderError help you judge err type
func IsTemplateRenderError(err error) bool {
	return isThatError(err, codeTemplateRenderError)
}

//IsValidateError help you judge err type
func IsValidateError(err error) bool {
	return isThatError(err, codeValidateError)
}

func isThatError(err error, code string) bool {
	_errs, ok := err.(Errors)

	if ok {
		for _, errItem := range _errs {
			if isThatError(errItem, code) {
				return true
			}
		}

		return false
	}

	_err, ok := err.(Error)
	if !ok {
		return false
	}

	//it was a single Error
	return _err.Code == code
}
