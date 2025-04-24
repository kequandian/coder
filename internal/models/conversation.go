package models

import (
	"sync"

	"github.com/cloudwego/eino/schema"
)

// Conversation stores the messages of a conversation
type Conversation struct {
	mu       sync.Mutex
	ID       string
	Messages []*schema.Message
}

// Append adds a message to the conversation
func (c *Conversation) Append(msg *schema.Message) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Messages = append(c.Messages, msg)
}

// GetMessages returns all messages in the conversation
func (c *Conversation) GetMessages() []*schema.Message {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Messages
}

// ConversationStore stores all conversations
type ConversationStore struct {
	mu            sync.Mutex
	conversations map[string]*Conversation
}

// NewConversationStore creates a new conversation store
func NewConversationStore() *ConversationStore {
	return &ConversationStore{
		conversations: make(map[string]*Conversation),
	}
}

// GetOrCreate gets or creates a conversation
func (s *ConversationStore) GetOrCreate(id string) *Conversation {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.conversations[id]; !ok {
		s.conversations[id] = &Conversation{
			ID:       id,
			Messages: make([]*schema.Message, 0),
		}
	}

	return s.conversations[id]
}
