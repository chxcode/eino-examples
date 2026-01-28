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
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino/schema"
)

// FileStore 文件存储实现
// 每个用户的会话存储为一个 .jsonl 文件
type FileStore struct {
	mu          sync.RWMutex
	dir         string
	maxMessages int
	cache       map[string]*Conversation // 内存缓存
	closed      bool
}

// NewFileStore 创建文件存储
func NewFileStore(dir string, maxMessages int) (*FileStore, error) {
	if dir == "" {
		dir = "data/conversations"
	}
	if maxMessages <= 0 {
		maxMessages = 20
	}

	// 创建目录
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	return &FileStore{
		dir:         dir,
		maxMessages: maxMessages,
		cache:       make(map[string]*Conversation),
	}, nil
}

// getFilePath 获取用户会话文件路径
func (s *FileStore) getFilePath(userID string) string {
	return filepath.Join(s.dir, userID+".jsonl")
}

// loadConversation 从文件加载会话
func (s *FileStore) loadConversation(userID string) (*Conversation, error) {
	filePath := s.getFilePath(userID)

	// 检查文件是否存在
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return nil, ErrConversationNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	conv := &Conversation{
		UserID:    userID,
		CreatedAt: info.ModTime(), // 使用文件修改时间作为创建时间
		UpdatedAt: info.ModTime(),
		Messages:  make([]Message, 0),
	}

	// 读取文件
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// 增加缓冲区大小以支持长消息
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var msg Message
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			// 尝试解析旧格式（直接是 schema.Message）
			var schemaMsg schema.Message
			if err2 := json.Unmarshal([]byte(line), &schemaMsg); err2 == nil {
				msg = FromSchemaMessage(&schemaMsg)
			} else {
				continue // 跳过无法解析的行
			}
		}
		conv.Messages = append(conv.Messages, msg)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return conv, nil
}

// getOrLoadConversation 获取或加载会话（带缓存）
func (s *FileStore) getOrLoadConversation(userID string, createIfNotExist bool) (*Conversation, error) {
	// 先检查缓存
	if conv, ok := s.cache[userID]; ok {
		return conv, nil
	}

	// 从文件加载
	conv, err := s.loadConversation(userID)
	if err == ErrConversationNotFound && createIfNotExist {
		// 创建新会话
		conv = &Conversation{
			UserID:    userID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Messages:  make([]Message, 0),
		}
		// 创建空文件
		filePath := s.getFilePath(userID)
		if err := os.WriteFile(filePath, []byte(""), 0644); err != nil {
			return nil, fmt.Errorf("failed to create file: %w", err)
		}
	} else if err != nil {
		return nil, err
	}

	// 加入缓存
	s.cache[userID] = conv
	return conv, nil
}

// appendToFile 追加消息到文件
func (s *FileStore) appendToFile(userID string, msg Message) error {
	filePath := s.getFilePath(userID)

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}
	if _, err := f.WriteString("\n"); err != nil {
		return fmt.Errorf("failed to write newline: %w", err)
	}

	return nil
}

// GetHistory 获取用户的会话历史
func (s *FileStore) GetHistory(ctx context.Context, userID string, maxMessages int) ([]*schema.Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, ErrStoreClosed
	}

	conv, err := s.getOrLoadConversation(userID, false)
	if err == ErrConversationNotFound {
		return []*schema.Message{}, nil
	}
	if err != nil {
		return nil, err
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
func (s *FileStore) AppendMessage(ctx context.Context, userID string, msg *schema.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrStoreClosed
	}

	conv, err := s.getOrLoadConversation(userID, true)
	if err != nil {
		return err
	}

	// 创建持久化消息
	persistMsg := FromSchemaMessage(msg)

	// 追加到内存
	conv.Messages = append(conv.Messages, persistMsg)
	conv.UpdatedAt = time.Now()

	// 追加到文件
	return s.appendToFile(userID, persistMsg)
}

// AppendMessages 批量追加消息
func (s *FileStore) AppendMessages(ctx context.Context, userID string, msgs []*schema.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrStoreClosed
	}

	conv, err := s.getOrLoadConversation(userID, true)
	if err != nil {
		return err
	}

	for _, msg := range msgs {
		persistMsg := FromSchemaMessage(msg)
		conv.Messages = append(conv.Messages, persistMsg)

		if err := s.appendToFile(userID, persistMsg); err != nil {
			return err
		}
	}

	conv.UpdatedAt = time.Now()
	return nil
}

// Clear 清空用户会话
func (s *FileStore) Clear(ctx context.Context, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrStoreClosed
	}

	// 清空缓存
	if conv, ok := s.cache[userID]; ok {
		conv.Messages = make([]Message, 0)
		conv.UpdatedAt = time.Now()
	}

	// 清空文件
	filePath := s.getFilePath(userID)
	return os.WriteFile(filePath, []byte(""), 0644)
}

// Delete 删除用户会话
func (s *FileStore) Delete(ctx context.Context, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrStoreClosed
	}

	// 删除缓存
	delete(s.cache, userID)

	// 删除文件
	filePath := s.getFilePath(userID)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// ListUsers 列出所有有会话记录的用户
func (s *FileStore) ListUsers(ctx context.Context) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, ErrStoreClosed
	}

	files, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	users := make([]string, 0, len(files))
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		name := file.Name()
		if strings.HasSuffix(name, ".jsonl") {
			users = append(users, strings.TrimSuffix(name, ".jsonl"))
		}
	}

	return users, nil
}

// GetConversation 获取完整会话信息
func (s *FileStore) GetConversation(ctx context.Context, userID string) (*Conversation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, ErrStoreClosed
	}

	return s.getOrLoadConversation(userID, false)
}

// Close 关闭存储
func (s *FileStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.closed = true
	s.cache = nil
	return nil
}
