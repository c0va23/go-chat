package main

import (
	"net/http"
	"fmt"
	"os"
	"encoding/json"
	"html/template"
	
	"github.com/gorilla/mux"
	"github.com/yosssi/ace"
	"github.com/op/go-logging"
	"code.google.com/p/go-uuid/uuid"
)

var logger *logging.Logger

func init() {
	logFormater := logging.MustStringFormatter(`%{color}[%{level:8s}] %{time} > %{message} %{color:reset}`)
	logging.SetFormatter(logFormater)
	logBackend := logging.NewLogBackend(os.Stdout, "", 0)
	logging.SetBackend(logBackend)

	logger = logging.MustGetLogger("sever")
}

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/api/messages", getMessages).Methods("GET")
	router.HandleFunc("/api/messages", createMessage).Methods("POST")
	router.HandleFunc("/api/messages", deleteMessages).Methods("DELETE")

	router.HandleFunc("/{user}", index).Methods("GET")

	serveMux := http.NewServeMux()
	serveMux.Handle("/assets/", http.FileServer(http.Dir(".")))
	serveMux.Handle("/", router)

	httpLogger := NewHttpLogger(serveMux)

	address := fmt.Sprintf("%s:%s", os.Getenv("HOST"), os.Getenv("PORT"))

	logger.Info("Listen on %s", address)

	logger.Fatal(http.ListenAndServe(address, httpLogger))
}

var indexTemplae = template.Must(ace.Load("templates/index", "", &ace.Options{
	DelimLeft: "<<",
	DelimRight: ">>",
	DynamicReload: true,
}))

func index(responseWriter http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	indexTemplae.Execute(responseWriter, vars)
}

func createMessage(responseWriter http.ResponseWriter, request *http.Request) {
	var messageData MessageData

	json.NewDecoder(request.Body).Decode(&messageData)

	message := NewMessage(messageData)

	messageList.Push(message)

	responseWriter.WriteHeader(http.StatusAccepted)
}

const LAST_EVENT_ID_HEADER = "Last-Event-ID"
const CONTENT_TYPE_HEADER = "Content-Type"
const EVENT_STREAM_TYPE = "text/event-stream"

func getMessages(responseWriter http.ResponseWriter, request *http.Request) {
	lastEventId := uuid.Parse(request.Header.Get(LAST_EVENT_ID_HEADER))

	flusher := responseWriter.(http.Flusher)

	responseWriter.Header().Add(CONTENT_TYPE_HEADER, EVENT_STREAM_TYPE)
	responseWriter.WriteHeader(http.StatusOK)

	iterator := messageList.Iterator(lastEventId)
	go iterator.Iterate()
	defer iterator.Close()

	for message := range iterator.Messages {
		fmt.Fprintf(responseWriter, "id: %s\n", message.Id)
		fmt.Fprint(responseWriter, "data: ")
		if encodeErr := json.NewEncoder(responseWriter).Encode(message); nil != encodeErr {
			logger.Error(encodeErr.Error())
			break
		}
		fmt.Fprint(responseWriter, "\n\n")
		flusher.Flush()
	}
}

func deleteMessages(responseWriter http.ResponseWriter, request *http.Request) {
	messageList.Clean()
	responseWriter.WriteHeader(http.StatusOK)
}
