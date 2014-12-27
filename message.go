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

type MessageIterator struct {
  *MessageList
  LastEventId uuid.UUID
  Channel chan *Message
}

func (messageIterator *MessageIterator) StartIndex() int {
  if nil != messageIterator.LastEventId {
    messageIterator.Lock()
    for index, message := range messageIterator.Messages {
      if uuid.Equal(message.Id, messageIterator.LastEventId) {
        return index
      }
    }
    messageIterator.Unlock()
  }
  return 0
}

func (messageIterator *MessageIterator) Iterate() {
  for index := messageIterator.StartIndex(); true; index++ {
    messageIterator.Lock()
    if len(messageIterator.Messages) == index {
      messageIterator.Wait()
    }
    messageIterator.Channel <- messageIterator.Messages[index]
    messageIterator.Unlock()
  }
}


func (messageList *MessageList) Iterator(lastEventId uuid.UUID) <- chan *Message {
  messageIterator := &MessageIterator {
    MessageList: messageList,
    LastEventId: lastEventId,
    Channel: make(chan *Message),
  }
  go messageIterator.Iterate()
  return messageIterator.Channel
}

func (messageList *MessageList) Clean() {
  messageList.Messages = make([]*Message, 0)
}

var messageList = NewMessageList()
