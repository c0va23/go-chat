package main

import (
	"net/http"
	"time"

	"github.com/op/go-logging"
	"code.google.com/p/go-uuid/uuid"
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
	requestId := uuid.NewRandom()
	httpLogger.logger.Info("(%s) Request: %s %s", requestId, request.Method, request.URL)
	startTime := time.Now()
	httpLogger.handler.ServeHTTP(responseWriter, request)
	endTime := time.Now()
	requestDuration := endTime.Sub(startTime)
	httpLogger.logger.Info("(%s) Response : %s", requestId, requestDuration)
}
