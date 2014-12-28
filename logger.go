package main

import (
	"net/http"
	"time"

	"github.com/op/go-logging"
)

type HttpLogger struct {
	logger *logging.Logger
	handler http.Handler
}

func NewHttpLogger(handler http.Handler) *HttpLogger {
	logger := logging.MustGetLogger("request")
	return &HttpLogger { logger, handler }
}

func (httpLogger *HttpLogger) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	httpLogger.logger.Info("Request: %s %s", request.Method, request.URL)
	startTime := time.Now()
	httpLogger.handler.ServeHTTP(responseWriter, request)
	endTime := time.Now()
	requestDuration := endTime.Sub(startTime)
	httpLogger.logger.Info("Response : %s", requestDuration)
}
