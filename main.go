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
)

func main() {
	router := router.NewRouter()

	router.Get("/api/messages", http.HandlerFunc(getMessages))
	router.Post("/api/messages", http.HandlerFunc(createMessage))

	mux := http.NewServeMux()
	mux.HandleFunc("/", index)
	mux.Handle("/assets/", http.FileServer(http.Dir(".")))
	mux.Handle("/api/", router)

	address := fmt.Sprintf("%s:%s", os.Getenv("HOST"), os.Getenv("PORT"))
	log.Fatal(http.ListenAndServe(address, mux))
}

var indexTemplae = template.Must(ace.Load("templates/index", "", &ace.Options{
	DelimLeft: "<<",
	DelimRight: ">>",
	DynamicReload: true,
}))

func index(responseWriter http.ResponseWriter, request *http.Request) {
	indexTemplae.Execute(responseWriter, map[string]string{"user": "he"})
}

func createMessage(responseWriter http.ResponseWriter, request *http.Request) {
	var messageData MessageData

	json.NewDecoder(request.Body).Decode(&messageData)

	message := NewMessage(messageData)

	messageList.Push(message)

	responseWriter.WriteHeader(http.StatusAccepted)
}

func getMessages(responseWriter http.ResponseWriter, request *http.Request) {
	flusher := responseWriter.(http.Flusher)

	responseWriter.Header().Add("Content-Type", "text/event-stream")
	responseWriter.WriteHeader(http.StatusOK)

	for message := range messageList.Iterator() {
		fmt.Fprint(responseWriter, "data: ")
		if encodeErr := json.NewEncoder(responseWriter).Encode(message); nil != encodeErr {
			log.Fatal(encodeErr)
		}
		fmt.Fprint(responseWriter, "\n\n")
		flusher.Flush()
	}
}
