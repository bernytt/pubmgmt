package http

import (
	log "github.com/Sirupsen/logrus"
	"github.com/fengxsong/pubmgmt/api"
	"gopkg.in/gin-gonic/gin.v1"
)

const (
	// ErrInvalidJSON defines an error raised the app is unable to parse request data
	ErrInvalidJSON = pub.Error("Invalid JSON")
	// ErrInvalidRequestFormat defines an error raised when the format of the data sent in a request is not valid
	ErrInvalidRequestFormat = pub.Error("Invalid request data format")
	// ErrInvalidQueryFormat defines an error raised when the data sent in the query or the URL is invalid
	ErrInvalidQueryFormat = pub.Error("Invalid query format")
	// ErrEmptyResponseBody defines an error raised when portainer excepts to parse the body of a HTTP response and there is nothing to parse
	ErrEmptyResponseBody = pub.Error("Empty response body")
)

type logger *log.Logger

// Error writes an API error message to the response and logger.
func Error(ctx *gin.Context, err error, code int, logger *log.Logger) {
	if logger != nil {
		logger.Errorf("http error: %s (code=%d)\n", err, code)
	}
	ctx.IndentedJSON(code, errorResponse{Err: err.Error()})
}

// writes errors and abort.
func ErrorAbort(ctx *gin.Context, err error, code int, logger *log.Logger) {
	Error(ctx, err, code, logger)
	ctx.Abort()
}

// print info level message to logger
func Infof(logger *log.Logger, format string, args ...interface{}) {
	if logger != nil {
		logger.Infof(format, args...)
	}
}

// print error level message to logger
func Errorf(logger *log.Logger, format string, args ...interface{}) {
	if logger != nil {
		logger.Errorf(format, args...)
	}
}

// errorResponse is a generic response for sending a error.
type errorResponse struct {
	Err string `json:"err,omitempty"`
}

// msgResponse is a generic response for sending a message when action is complete.
type msgResponse struct {
	Msg string `json:"msg,omitempty"`
}
