/*
 * Copyright 2024 CloudWeGo Authors
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

package memory

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cloudwego/eino/schema"
)

// ConversationMemory 会话记忆管理
type ConversationMemory struct {
	mu            sync.Mutex
	dir           string
	maxWindowSize int
	conversations map[string]*Conversation
}

// NewConversationMemory 创建会话记忆管理器
func NewConversationMemory(dir string, maxWindowSize int) *ConversationMemory {
	if dir == "" {
		dir = "data/memory"
	}
	if maxWindowSize <= 0 {
		maxWindowSize = 10
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil
	}

	return &ConversationMemory{
		dir:           dir,
		maxWindowSize: maxWindowSize,
		conversations: make(map[string]*Conversation),
	}
}

// GetConversation 获取或创建会话
func (m *ConversationMemory) GetConversation(userID string, createIfNotExist bool) *Conversation {
	m.mu.Lock()
	defer m.mu.Unlock()

	if conv, ok := m.conversations[userID]; ok {
		return conv
	}

	filePath := filepath.Join(m.dir, userID+".jsonl")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if createIfNotExist {
			if err := os.WriteFile(filePath, []byte(""), 0644); err != nil {
				return nil
			}
		} else {
			return nil
		}
	}

	conv := &Conversation{
		UserID:        userID,
		Messages:      make([]*schema.Message, 0),
		filePath:      filePath,
		maxWindowSize: m.maxWindowSize,
	}
	conv.load()
	m.conversations[userID] = conv
	return conv
}

// ListConversations 列出所有会话
func (m *ConversationMemory) ListConversations() []string {
	m.mu.Lock()
	defer m.mu.Unlock()

	files, err := os.ReadDir(m.dir)
	if err != nil {
		return nil
	}

	userIDs := make([]string, 0, len(files))
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		userIDs = append(userIDs, strings.TrimSuffix(file.Name(), ".jsonl"))
	}

	return userIDs
}

// DeleteConversation 删除会话
func (m *ConversationMemory) DeleteConversation(userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	filePath := filepath.Join(m.dir, userID+".jsonl")
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	delete(m.conversations, userID)
	return nil
}

// Conversation 单个会话
type Conversation struct {
	mu sync.Mutex

	UserID   string            `json:"user_id"`
	Messages []*schema.Message `json:"messages"`

	filePath      string
	maxWindowSize int
}

// Append 添加消息
func (c *Conversation) Append(msg *schema.Message) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Messages = append(c.Messages, msg)
	c.save(msg)
}

// GetFullMessages 获取所有消息
func (c *Conversation) GetFullMessages() []*schema.Message {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.Messages
}

// GetMessages 获取最近的消息（滑动窗口）
func (c *Conversation) GetMessages() []*schema.Message {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.Messages) > c.maxWindowSize {
		return c.Messages[len(c.Messages)-c.maxWindowSize:]
	}

	return c.Messages
}

// Clear 清空会话
func (c *Conversation) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Messages = make([]*schema.Message, 0)
	// 清空文件
	_ = os.WriteFile(c.filePath, []byte(""), 0644)
}

func (c *Conversation) load() error {
	reader, err := os.Open(c.filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer reader.Close()

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		var msg schema.Message
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			continue
		}
		c.Messages = append(c.Messages, &msg)
	}

	return scanner.Err()
}

func (c *Conversation) save(msg *schema.Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	f, err := os.OpenFile(c.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	f.Write(data)
	f.WriteString("\n")
}
