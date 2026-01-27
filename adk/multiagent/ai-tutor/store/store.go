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

// Package store 提供会话持久化接口和实现
// 支持多种存储后端（文件、MySQL、Redis 等）
package store

import (
	"context"
	"time"

	"github.com/cloudwego/eino/schema"
)

// Message 消息结构，用于持久化存储
type Message struct {
	ID        string          `json:"id"`
	Role      schema.RoleType `json:"role"`
	Content   string          `json:"content"`
	Timestamp time.Time       `json:"timestamp"`
	Metadata  map[string]any  `json:"metadata,omitempty"` // 扩展字段，可存储额外信息
}

// Conversation 会话结构
type Conversation struct {
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Messages  []Message `json:"messages"`
}

// ConversationStore 会话存储接口
// 实现此接口可以支持不同的存储后端（文件、MySQL、Redis 等）
type ConversationStore interface {
	// GetHistory 获取用户的会话历史
	// maxMessages: 最大返回消息数，0 表示返回所有
	GetHistory(ctx context.Context, userID string, maxMessages int) ([]*schema.Message, error)

	// AppendMessage 追加消息到会话
	AppendMessage(ctx context.Context, userID string, msg *schema.Message) error

	// AppendMessages 批量追加消息
	AppendMessages(ctx context.Context, userID string, msgs []*schema.Message) error

	// Clear 清空用户会话
	Clear(ctx context.Context, userID string) error

	// Delete 删除用户会话
	Delete(ctx context.Context, userID string) error

	// ListUsers 列出所有有会话记录的用户
	ListUsers(ctx context.Context) ([]string, error)

	// GetConversation 获取完整会话信息
	GetConversation(ctx context.Context, userID string) (*Conversation, error)

	// Close 关闭存储连接（用于数据库连接等资源释放）
	Close() error
}

// StoreConfig 存储配置
type StoreConfig struct {
	// 通用配置
	MaxMessages int `json:"max_messages"` // 滑动窗口大小，0 表示不限制

	// 文件存储配置
	FileDir string `json:"file_dir"` // 文件存储目录

	// MySQL 配置（预留）
	MySQLDSN string `json:"mysql_dsn"` // MySQL 连接字符串

	// Redis 配置（预留）
	RedisAddr     string `json:"redis_addr"`     // Redis 地址
	RedisPassword string `json:"redis_password"` // Redis 密码
	RedisDB       int    `json:"redis_db"`       // Redis 数据库
}

// StoreType 存储类型
type StoreType string

const (
	StoreTypeMemory StoreType = "memory" // 内存存储（用于测试）
	StoreTypeFile   StoreType = "file"   // 文件存储
	StoreTypeMySQL  StoreType = "mysql"  // MySQL 存储
	StoreTypeRedis  StoreType = "redis"  // Redis 存储
)

// NewStore 根据类型创建存储实例
func NewStore(storeType StoreType, config *StoreConfig) (ConversationStore, error) {
	if config == nil {
		config = &StoreConfig{
			MaxMessages: 20,
		}
	}

	switch storeType {
	case StoreTypeMemory:
		return NewMemoryStore(config.MaxMessages), nil
	case StoreTypeFile:
		return NewFileStore(config.FileDir, config.MaxMessages)
	case StoreTypeMySQL:
		// TODO: 实现 MySQL 存储
		// return NewMySQLStore(config.MySQLDSN, config.MaxMessages)
		return nil, ErrStoreTypeNotImplemented
	case StoreTypeRedis:
		// TODO: 实现 Redis 存储
		// return NewRedisStore(config.RedisAddr, config.RedisPassword, config.RedisDB, config.MaxMessages)
		return nil, ErrStoreTypeNotImplemented
	default:
		return nil, ErrUnknownStoreType
	}
}

// ToSchemaMessage 将持久化消息转换为 schema.Message
func (m *Message) ToSchemaMessage() *schema.Message {
	return &schema.Message{
		Role:    m.Role,
		Content: m.Content,
	}
}

// FromSchemaMessage 从 schema.Message 创建持久化消息
func FromSchemaMessage(msg *schema.Message) Message {
	return Message{
		ID:        generateMessageID(),
		Role:      msg.Role,
		Content:   msg.Content,
		Timestamp: time.Now(),
	}
}

// 辅助函数：生成消息 ID
func generateMessageID() string {
	return time.Now().Format("20060102150405.000000")
}
