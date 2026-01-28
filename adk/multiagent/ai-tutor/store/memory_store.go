/*
 * Copyright 2025 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package store

import (
	"context"
	"sync"
	"time"

	"github.com/cloudwego/eino/schema"
)

// MemoryStore 内存存储实现
// 适用于测试和开发环境，数据不会持久化
type MemoryStore struct {
	mu            sync.RWMutex
	conversations map[string]*Conversation
	maxMessages   int
	closed        bool
}

// NewMemoryStore 创建内存存储
func NewMemoryStore(maxMessages int) *MemoryStore {
	if maxMessages <= 0 {
		maxMessages = 20
	}
	return &MemoryStore{
		conversations: make(map[string]*Conversation),
		maxMessages:   maxMessages,
	}
}

// GetHistory 获取用户的会话历史
func (s *MemoryStore) GetHistory(ctx context.Context, userID string, maxMessages int) ([]*schema.Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, ErrStoreClosed
	}

	conv, ok := s.conversations[userID]
	if !ok {
		return []*schema.Message{}, nil
	}

	// 确定返回的消息数量
	limit := len(conv.Messages)
	if maxMessages > 0 && maxMessages < limit {
		limit = maxMessages
	} else if s.maxMessages > 0 && s.maxMessages < limit {
		limit = s.maxMessages
	}

	// 返回最近的消息
	start := len(conv.Messages) - limit
	if start < 0 {
		start = 0
	}

	result := make([]*schema.Message, 0, limit)
	for _, msg := range conv.Messages[start:] {
		result = append(result, msg.ToSchemaMessage())
	}
	return result, nil
}

// AppendMessage 追加消息到会话
func (s *MemoryStore) AppendMessage(ctx context.Context, userID string, msg *schema.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrStoreClosed
	}

	conv, ok := s.conversations[userID]
	if !ok {
		conv = &Conversation{
			UserID:    userID,
			CreatedAt: time.Now(),
			Messages:  make([]Message, 0),
		}
		s.conversations[userID] = conv
	}

	conv.Messages = append(conv.Messages, FromSchemaMessage(msg))
	conv.UpdatedAt = time.Now()
	return nil
}

// AppendMessages 批量追加消息
func (s *MemoryStore) AppendMessages(ctx context.Context, userID string, msgs []*schema.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrStoreClosed
	}

	conv, ok := s.conversations[userID]
	if !ok {
		conv = &Conversation{
			UserID:    userID,
			CreatedAt: time.Now(),
			Messages:  make([]Message, 0),
		}
		s.conversations[userID] = conv
	}

	for _, msg := range msgs {
		conv.Messages = append(conv.Messages, FromSchemaMessage(msg))
	}
	conv.UpdatedAt = time.Now()
	return nil
}

// Clear 清空用户会话
func (s *MemoryStore) Clear(ctx context.Context, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrStoreClosed
	}

	if conv, ok := s.conversations[userID]; ok {
		conv.Messages = make([]Message, 0)
		conv.UpdatedAt = time.Now()
	}
	return nil
}

// Delete 删除用户会话
func (s *MemoryStore) Delete(ctx context.Context, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrStoreClosed
	}

	delete(s.conversations, userID)
	return nil
}

// ListUsers 列出所有有会话记录的用户
func (s *MemoryStore) ListUsers(ctx context.Context) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, ErrStoreClosed
	}

	users := make([]string, 0, len(s.conversations))
	for userID := range s.conversations {
		users = append(users, userID)
	}
	return users, nil
}

// GetConversation 获取完整会话信息
func (s *MemoryStore) GetConversation(ctx context.Context, userID string) (*Conversation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, ErrStoreClosed
	}

	conv, ok := s.conversations[userID]
	if !ok {
		return nil, ErrConversationNotFound
	}
	return conv, nil
}

// Close 关闭存储（内存存储无需特殊处理）
func (s *MemoryStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.closed = true
	s.conversations = nil
	return nil
}
