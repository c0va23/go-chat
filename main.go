package main

import (
	"net/http"
	"fmt"
	"os"
	"log"
	"encoding/json"
	"html/template"
	
	"github.com/gorilla/mux"
	"github.com/yosssi/ace"
	"code.google.com/p/go-uuid/uuid"
)

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/api/messages", getMessages).Methods("GET")
	router.HandleFunc("/api/messages", createMessage).Methods("POST")
	router.HandleFunc("/api/messages", clean).Methods("DELETE")

	router.HandleFunc("/{user}", index).Methods("GET")

	serveMux := http.NewServeMux()
	serveMux.Handle("/assets/", http.FileServer(http.Dir(".")))
	serveMux.Handle("/", router)

	address := fmt.Sprintf("%s:%s", os.Getenv("HOST"), os.Getenv("PORT"))
	log.Fatal(http.ListenAndServe(address, serveMux))
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

	for message := range messageList.Iterator(lastEventId) {
		fmt.Fprintf(responseWriter, "id: %s\n", message.Id)
		fmt.Fprint(responseWriter, "data: ")
		if encodeErr := json.NewEncoder(responseWriter).Encode(message); nil != encodeErr {
			log.Println(encodeErr.Error())
			return
		}
		fmt.Fprint(responseWriter, "\n\n")
		flusher.Flush()
	}
}

func clean(responseWriter http.ResponseWriter, request *http.Request) {
	messageList.Clean()
	responseWriter.WriteHeader(http.StatusOK)
}
