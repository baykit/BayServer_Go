package util

import (
	"fmt"
)

type Message interface {
	MessageMap() map[string]string
	Get(key string, args ...interface{}) string
}

type messageImpl struct {
	messageMap map[string]string
}

func NewMessage() Message {
	return &messageImpl{
		messageMap: make(map[string]string),
	}
}

func (m messageImpl) MessageMap() map[string]string {
	return m.messageMap
}

func (m messageImpl) Get(key string, args ...interface{}) string {
	msg := m.messageMap[key]
	if msg == "" {
		msg = key
	}
	return fmt.Sprintf(msg, args...)
}
