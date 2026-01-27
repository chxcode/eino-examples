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

/*
MySQLStore MySQL 存储实现示例

使用前需要创建数据表：

CREATE TABLE IF NOT EXISTS conversations (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id)
);

CREATE TABLE IF NOT EXISTS messages (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    conversation_id BIGINT NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    message_id VARCHAR(50) NOT NULL,
    role VARCHAR(20) NOT NULL,
    content TEXT NOT NULL,
    metadata JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE,
    INDEX idx_user_id (user_id),
    INDEX idx_created_at (created_at)
);

实现示例（需要引入 database/sql 和 mysql 驱动）：

import (
    "context"
    "database/sql"
    "encoding/json"
    "time"

    _ "github.com/go-sql-driver/mysql"
    "github.com/cloudwego/eino/schema"
)

type MySQLStore struct {
    db          *sql.DB
    maxMessages int
}

func NewMySQLStore(dsn string, maxMessages int) (*MySQLStore, error) {
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return nil, err
    }

    if err := db.Ping(); err != nil {
        return nil, err
    }

    return &MySQLStore{
        db:          db,
        maxMessages: maxMessages,
    }, nil
}

func (s *MySQLStore) GetHistory(ctx context.Context, userID string, maxMessages int) ([]*schema.Message, error) {
    limit := s.maxMessages
    if maxMessages > 0 && maxMessages < limit {
        limit = maxMessages
    }

    query := `
        SELECT role, content FROM messages
        WHERE user_id = ?
        ORDER BY created_at DESC
        LIMIT ?
    `

    rows, err := s.db.QueryContext(ctx, query, userID, limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var messages []*schema.Message
    for rows.Next() {
        var role, content string
        if err := rows.Scan(&role, &content); err != nil {
            return nil, err
        }
        messages = append([]*schema.Message{{
            Role:    schema.RoleType(role),
            Content: content,
        }}, messages...) // 反转顺序
    }

    return messages, rows.Err()
}

func (s *MySQLStore) AppendMessage(ctx context.Context, userID string, msg *schema.Message) error {
    // 获取或创建会话
    convID, err := s.getOrCreateConversation(ctx, userID)
    if err != nil {
        return err
    }

    query := `
        INSERT INTO messages (conversation_id, user_id, message_id, role, content, metadata)
        VALUES (?, ?, ?, ?, ?, ?)
    `

    msgID := generateMessageID()
    _, err = s.db.ExecContext(ctx, query, convID, userID, msgID, string(msg.Role), msg.Content, "{}")
    return err
}

func (s *MySQLStore) getOrCreateConversation(ctx context.Context, userID string) (int64, error) {
    // 尝试获取现有会话
    var convID int64
    err := s.db.QueryRowContext(ctx,
        "SELECT id FROM conversations WHERE user_id = ?", userID,
    ).Scan(&convID)

    if err == sql.ErrNoRows {
        // 创建新会话
        result, err := s.db.ExecContext(ctx,
            "INSERT INTO conversations (user_id) VALUES (?)", userID,
        )
        if err != nil {
            return 0, err
        }
        return result.LastInsertId()
    }

    return convID, err
}

func (s *MySQLStore) AppendMessages(ctx context.Context, userID string, msgs []*schema.Message) error {
    for _, msg := range msgs {
        if err := s.AppendMessage(ctx, userID, msg); err != nil {
            return err
        }
    }
    return nil
}

func (s *MySQLStore) Clear(ctx context.Context, userID string) error {
    _, err := s.db.ExecContext(ctx, "DELETE FROM messages WHERE user_id = ?", userID)
    return err
}

func (s *MySQLStore) Delete(ctx context.Context, userID string) error {
    _, err := s.db.ExecContext(ctx, "DELETE FROM conversations WHERE user_id = ?", userID)
    return err
}

func (s *MySQLStore) ListUsers(ctx context.Context) ([]string, error) {
    rows, err := s.db.QueryContext(ctx, "SELECT DISTINCT user_id FROM conversations")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var users []string
    for rows.Next() {
        var userID string
        if err := rows.Scan(&userID); err != nil {
            return nil, err
        }
        users = append(users, userID)
    }
    return users, rows.Err()
}

func (s *MySQLStore) GetConversation(ctx context.Context, userID string) (*Conversation, error) {
    var conv Conversation
    err := s.db.QueryRowContext(ctx,
        "SELECT user_id, created_at, updated_at FROM conversations WHERE user_id = ?",
        userID,
    ).Scan(&conv.UserID, &conv.CreatedAt, &conv.UpdatedAt)

    if err == sql.ErrNoRows {
        return nil, ErrConversationNotFound
    }
    if err != nil {
        return nil, err
    }

    // 加载消息
    messages, err := s.GetHistory(ctx, userID, 0)
    if err != nil {
        return nil, err
    }

    for _, msg := range messages {
        conv.Messages = append(conv.Messages, FromSchemaMessage(msg))
    }

    return &conv, nil
}

func (s *MySQLStore) Close() error {
    return s.db.Close()
}
*/
