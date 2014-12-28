package main

import (
  "time"
  "sync"
  "errors"

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

type MessageListCursor interface {
  NextItem() *MessageListItem
  Message() *Message
}

type MessageListItem struct {
  message *Message
  nextItem *MessageListItem
}

func (messageListItem *MessageListItem) NextItem() *MessageListItem {
  return messageListItem.nextItem
}

func (messageListItem *MessageListItem) Message() *Message {
  return messageListItem.message
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

  FirstItem *MessageListItem
  LastItem *MessageListItem
}

func NewMessageList() *MessageList {
  mutex:= new(sync.Mutex)
  message := &MessageList{
    Locker: mutex,
    Cond: sync.NewCond(mutex),
  }
  return message
}

func (messageList *MessageList) NextItem() *MessageListItem {
  return messageList.FirstItem
}

func (messageList *MessageList) Message() *Message {
  return nil
}

func (messageList *MessageList) Push(message *Message) {
  messageListItem := &MessageListItem { message, nil }

  if nil == messageList.FirstItem {
    messageList.FirstItem = messageListItem
  } else {
    messageList.LastItem.nextItem = messageListItem
  }

  messageList.LastItem = messageListItem

  messageList.Broadcast()
}

func (messageList *MessageList) Clean() {
  messageList.Lock()
  defer messageList.Unlock()

  messageList.FirstItem = nil
  messageList.LastItem = nil
}

type MessageIterator struct {
  *MessageList
  LastEventId uuid.UUID
  Messages chan *Message
  Closed chan struct{}
}

func (messageIterator *MessageIterator) StartItem() MessageListCursor {
  if nil != messageIterator.LastEventId && nil != messageIterator.MessageList.FirstItem {
    
    for currentItem := messageList.FirstItem; nil != currentItem.NextItem(); currentItem = currentItem.NextItem() {
      if uuid.Equal(currentItem.Message().Id, messageIterator.LastEventId) {
        return currentItem
      }
    }
  }
  return messageIterator.MessageList
}

func (messageIterator *MessageIterator) Publish(message *Message) error {
  select {
  case messageIterator.Messages <- message:
    return nil
  case <- messageIterator.Closed:
    return errors.New("Iterator closed")
  }
}

func (messageIterator *MessageIterator) Iterate() {
  messageIterator.Lock()
  defer messageIterator.Unlock()

  var currentItem MessageListCursor = messageIterator.StartItem()

  for {
    if nil == currentItem.NextItem() {
      messageIterator.Wait()
    }
    currentItem = currentItem.NextItem()

    if err := messageIterator.Publish(currentItem.Message()); nil != err {
      logger.Error(err.Error())
      break
    }
 }
}

func (messageIterator *MessageIterator) Close() {
  messageIterator.Closed <- struct{}{}
}

func (messageList *MessageList) Iterator(lastEventId uuid.UUID) *MessageIterator {
  return &MessageIterator {
    MessageList: messageList,
    LastEventId: lastEventId,
    Messages: make(chan *Message),
    Closed: make(chan struct{}),
  }
}

var messageList = NewMessageList()
