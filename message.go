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

type MessageNextItem interface {
  NextItem() *MessageListItem
}

type MessageListItem struct {
  *Message
  Next *MessageListItem
}

func (messageListItem *MessageListItem) NextItem() *MessageListItem {
  return messageListItem.Next
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

func (messageList *MessageList) Push(message *Message) {
  messageListItem := &MessageListItem {
    Message: message,
  }

  if nil == messageList.FirstItem {
    messageList.FirstItem = messageListItem
  } else {
    messageList.LastItem.Next = messageListItem
  }

  messageList.LastItem = messageListItem

  messageList.Broadcast()
}

type MessageIterator struct {
  *MessageList
  LastEventId uuid.UUID
  Channel chan *Message
}

func (messageIterator *MessageIterator) StartItem() *MessageListItem {
  if nil != messageIterator.LastEventId && nil != messageIterator.MessageList.FirstItem {
    
    for currentItem := messageList.FirstItem; nil != currentItem.NextItem(); currentItem = currentItem.NextItem() {
      if uuid.Equal(currentItem.Message.Id, messageIterator.LastEventId) {
        return currentItem
      }
    }
  }
  return messageIterator.MessageList.FirstItem
}

func (messageIterator *MessageIterator) Iterate() {
  messageIterator.Lock()
  defer messageIterator.Unlock()

  currentItem := messageIterator.StartItem()
  if nil == currentItem {
    logger.Debug("wait start item")
    messageIterator.Wait()
    currentItem = messageIterator.StartItem()
  }
  logger.Debug("start write to channel")
  messageIterator.Channel <- currentItem.Message
  logger.Debug("end write to channel")
  for {
    if nil == currentItem.NextItem() {
      logger.Debug("wait next imte")
      messageIterator.Wait()
    }
    logger.Debug("next item")
    currentItem = currentItem.NextItem()

    logger.Debug("start write to channel")
    select {
      case messageIterator.Channel <- currentItem. Message:
        logger.Debug("end write to channel")
        continue
      default:
        logger.Debug("error write to channel")
        break
    }
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
  messageList.Lock()
  defer messageList.Unlock()

  messageList.FirstItem = nil
  messageList.LastItem = nil
}

var messageList = NewMessageList()
