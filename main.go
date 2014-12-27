package main

import (
	"net/http"
	"fmt"
	"os"
	"log"
	"encoding/json"
	"html/template"
	
	"github.com/toonketels/router"
	"github.com/yosssi/ace"
	"code.google.com/p/go-uuid/uuid"
)

func main() {
	router := router.NewRouter()

	router.Get("/api/messages", http.HandlerFunc(getMessages))
	router.Post("/api/messages", http.HandlerFunc(createMessage))
	router.Get("/api/clean", http.HandlerFunc(clean))

	router.Get("/:user", index)

	mux := http.NewServeMux()
	mux.Handle("/assets/", http.FileServer(http.Dir(".")))
	mux.Handle("/", router)

	address := fmt.Sprintf("%s:%s", os.Getenv("HOST"), os.Getenv("PORT"))
	log.Fatal(http.ListenAndServe(address, mux))
}

var indexTemplae = template.Must(ace.Load("templates/index", "", &ace.Options{
	DelimLeft: "<<",
	DelimRight: ">>",
	DynamicReload: true,
}))

func index(responseWriter http.ResponseWriter, request *http.Request) {
	context := router.Context(request)
	user := context.Params["user"]
	indexTemplae.Execute(responseWriter, map[string]string{"user": user})
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
			log.Println(encodeErr)
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
