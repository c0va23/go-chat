package main

import (
	"net/http"
	"fmt"
	"os"
	"log"
	"encoding/json"
	
	"github.com/toonketels/router"
)

func main() {
	router := router.NewRouter()

	router.Get("/api/messages", http.HandlerFunc(getMessages))
	router.Post("/api/messages", http.HandlerFunc(createMessage))

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("./assets")))
	mux.Handle("/assets/", http.FileServer(http.Dir(".")))
	mux.Handle("/api/", router)

	address := fmt.Sprintf("%s:%s", os.Getenv("HOST"), os.Getenv("PORT"))
	log.Fatal(http.ListenAndServe(address, mux))
}

func createMessage(responseWriter http.ResponseWriter, request *http.Request) {
	var messageData MessageData

	json.NewDecoder(request.Body).Decode(&messageData)

	message := NewMessage(messageData)

	messages <- message

	responseWriter.WriteHeader(http.StatusAccepted)
}

func getMessages(responseWriter http.ResponseWriter, request *http.Request) {
	flusher := responseWriter.(http.Flusher)

	responseWriter.Header().Add("Content-Type", "text/event-stream")
	flusher.Flush()

	for {
		message := <- messages
		responseWriter.Write([]byte("data:"))
		if encodeErr := json.NewEncoder(responseWriter).Encode(message); nil != encodeErr {
			log.Fatal(encodeErr)
		}
		responseWriter.Write([]byte("\n\n"))
		flusher.Flush()
	}
}
