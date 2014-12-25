package main

import (
  "time"

  "code.google.com/p/go-uuid/uuid"
)

type MessageData struct {
  Text string `json:"text"`
  User string `json:"user"`
}

type Message struct {
  Id uuid.UUID `json:"id"`
  Time time.Time `json:"time"`

  MessageData
}

var messages = make(chan *Message, 10)

func NewMessage(messageData MessageData) *Message {
  return &Message {
    Id: uuid.NewUUID(),
    Time: time.Now(),
    MessageData: messageData,
  }
}
