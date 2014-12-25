package main

import (
  "time"
  "sync"

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

func NewMessage(messageData MessageData) *Message {
  return &Message {
    Id: uuid.NewUUID(),
    Time: time.Now(),
    MessageData: messageData,
  }
}

type MessageList struct {
  sync.Locker
  *sync.Cond

  Messages []*Message
}

func NewMessageList() *MessageList {
  mutex:= new(sync.Mutex)
  message := &MessageList{
    Locker: mutex,
    Cond: sync.NewCond(mutex),
    Messages: make([]*Message, 0),
  }
  return message
}

func (messageList *MessageList) Push(message *Message) {
  messageList.Lock()
  messageList.Messages = append(messageList.Messages, message)
  messageList.Unlock()
  messageList.Broadcast()
}

func (messageList *MessageList) Iterate(channel chan <- *Message) {
  for index := 0; true; index++ {
    messageList.Lock()
    if len(messageList.Messages) == index {
      messageList.Wait()
    }
    channel <- messageList.Messages[index]
    messageList.Unlock()
  }
}

func (messageList *MessageList) Iterator() <- chan *Message {
  channel := make(chan *Message)
  go messageList.Iterate(channel)
  return channel
}

var messageList = NewMessageList()
