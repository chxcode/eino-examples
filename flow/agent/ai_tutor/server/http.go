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

package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/hertz-contrib/sse"

	"github.com/cloudwego/eino-examples/flow/agent/ai_tutor/agent"
	"github.com/cloudwego/eino-examples/flow/agent/ai_tutor/config"
	"github.com/cloudwego/eino-examples/flow/agent/ai_tutor/memory"
	"github.com/cloudwego/eino/schema"
)

// HTTPServer HTTP 服务
type HTTPServer struct {
	cfg         *config.Config
	coordinator *agent.AITutorCoordinator
	memory      *memory.ConversationMemory
	server      *server.Hertz
}

// NewHTTPServer 创建 HTTP 服务
func NewHTTPServer(cfg *config.Config, coordinator *agent.AITutorCoordinator, mem *memory.ConversationMemory) *HTTPServer {
	return &HTTPServer{
		cfg:         cfg,
		coordinator: coordinator,
		memory:      mem,
	}
}

// ChatRequest 聊天请求
type ChatRequest struct {
	UID     string `json:"uid" query:"uid"`
	Message string `json:"message" query:"message"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
	UserID  string `json:"user_id,omitempty"`
}

// Start 启动 HTTP 服务
func (s *HTTPServer) Start() error {
	h := server.Default(server.WithHostPorts(":" + s.cfg.HTTPPort))
	s.server = h

	// API 路由
	api := h.Group("/api")
	{
		// 聊天接口 - 支持 GET 和 POST
		api.GET("/chat", s.handleChat)
		api.POST("/chat", s.handleChatPost)

		// 流式聊天接口
		api.GET("/chat/stream", s.handleChatStream)

		// 历史记录接口
		api.GET("/history", s.handleGetHistory)
		api.DELETE("/history", s.handleDeleteHistory)

		// 健康检查
		api.GET("/health", s.handleHealth)
	}

	// 根路由 - 简单的使用说明
	h.GET("/", s.handleIndex)

	log.Printf("[AI Tutor] HTTP server starting on port %s...\n", s.cfg.HTTPPort)
	h.Spin()
	return nil
}

// handleIndex 首页处理
func (s *HTTPServer) handleIndex(ctx context.Context, c *app.RequestContext) {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>AI 学管助手</title>
    <meta charset="utf-8">
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
        h1 { color: #333; }
        .endpoint { background: #f5f5f5; padding: 15px; margin: 10px 0; border-radius: 5px; }
        code { background: #e0e0e0; padding: 2px 6px; border-radius: 3px; }
    </style>
</head>
<body>
    <h1>🎓 AI 学管助手 API</h1>
    <p>欢迎使用 AI 学管助手 API！以下是可用的接口：</p>
    
    <div class="endpoint">
        <h3>GET /api/chat</h3>
        <p>发送聊天消息（同步响应）</p>
        <p>参数：<code>uid</code> - 用户ID，<code>message</code> - 消息内容</p>
        <p>示例：<code>/api/chat?uid=user123&message=我想约课</code></p>
    </div>
    
    <div class="endpoint">
        <h3>POST /api/chat</h3>
        <p>发送聊天消息（JSON 格式）</p>
        <p>Body：<code>{"uid": "user123", "message": "我想约课"}</code></p>
    </div>
    
    <div class="endpoint">
        <h3>GET /api/chat/stream</h3>
        <p>发送聊天消息（SSE 流式响应）</p>
        <p>参数：<code>uid</code> - 用户ID，<code>message</code> - 消息内容</p>
    </div>
    
    <div class="endpoint">
        <h3>GET /api/history</h3>
        <p>获取会话历史</p>
        <p>参数：<code>uid</code> - 用户ID（可选，不传则列出所有用户）</p>
    </div>
    
    <div class="endpoint">
        <h3>DELETE /api/history</h3>
        <p>删除会话历史</p>
        <p>参数：<code>uid</code> - 用户ID</p>
    </div>
    
    <div class="endpoint">
        <h3>GET /api/health</h3>
        <p>健康检查</p>
    </div>
</body>
</html>`
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(consts.StatusOK, html)
}

// handleChat GET 方式聊天
func (s *HTTPServer) handleChat(ctx context.Context, c *app.RequestContext) {
	uid := c.Query("uid")
	message := c.Query("message")

	if uid == "" || message == "" {
		c.JSON(consts.StatusBadRequest, ChatResponse{
			Status: "error",
			Error:  "missing uid or message parameter",
		})
		return
	}

	s.processChat(ctx, c, uid, message)
}

// handleChatPost POST 方式聊天
func (s *HTTPServer) handleChatPost(ctx context.Context, c *app.RequestContext) {
	var req ChatRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(consts.StatusBadRequest, ChatResponse{
			Status: "error",
			Error:  "invalid request body: " + err.Error(),
		})
		return
	}

	if req.UID == "" || req.Message == "" {
		c.JSON(consts.StatusBadRequest, ChatResponse{
			Status: "error",
			Error:  "missing uid or message",
		})
		return
	}

	s.processChat(ctx, c, req.UID, req.Message)
}

// processChat 处理聊天请求
func (s *HTTPServer) processChat(ctx context.Context, c *app.RequestContext, uid, message string) {
	log.Printf("[Chat] uid=%s, message=%s\n", uid, message)

	// 获取或创建会话
	conversation := s.memory.GetConversation(uid, true)
	history := conversation.GetMessages()

	// 处理消息
	resp, err := s.coordinator.Process(ctx, uid, message, history)
	if err != nil {
		log.Printf("[Chat] Error: %v\n", err)
		c.JSON(consts.StatusInternalServerError, ChatResponse{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}

	// 保存会话
	conversation.Append(schema.UserMessage(message))
	conversation.Append(resp)

	c.JSON(consts.StatusOK, ChatResponse{
		Status:  "success",
		Message: resp.Content,
		UserID:  uid,
	})
}

// handleChatStream 流式聊天
func (s *HTTPServer) handleChatStream(ctx context.Context, c *app.RequestContext) {
	uid := c.Query("uid")
	message := c.Query("message")

	if uid == "" || message == "" {
		c.JSON(consts.StatusBadRequest, ChatResponse{
			Status: "error",
			Error:  "missing uid or message parameter",
		})
		return
	}

	log.Printf("[ChatStream] uid=%s, message=%s\n", uid, message)

	// 获取或创建会话
	conversation := s.memory.GetConversation(uid, true)
	history := conversation.GetMessages()

	// 流式处理消息
	sr, err := s.coordinator.ProcessStream(ctx, uid, message, history)
	if err != nil {
		log.Printf("[ChatStream] Error: %v\n", err)
		c.JSON(consts.StatusInternalServerError, ChatResponse{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}

	// 创建 SSE 流
	stream := sse.NewStream(c)
	var fullContent string

	defer func() {
		sr.Close()
		c.Flush()

		// 保存会话
		conversation.Append(schema.UserMessage(message))
		if fullContent != "" {
			conversation.Append(schema.AssistantMessage(fullContent, nil))
		}

		log.Printf("[ChatStream] Finished uid=%s\n", uid)
	}()

outer:
	for {
		select {
		case <-ctx.Done():
			log.Printf("[ChatStream] Context done for uid=%s\n", uid)
			return
		default:
			msg, err := sr.Recv()
			if errors.Is(err, io.EOF) {
				break outer
			}
			if err != nil {
				log.Printf("[ChatStream] Error receiving: %v\n", err)
				break outer
			}

			fullContent += msg.Content

			err = stream.Publish(&sse.Event{
				Data: []byte(msg.Content),
			})
			if err != nil {
				log.Printf("[ChatStream] Error publishing: %v\n", err)
				break outer
			}
		}
	}
}

// handleGetHistory 获取历史记录
func (s *HTTPServer) handleGetHistory(ctx context.Context, c *app.RequestContext) {
	uid := c.Query("uid")

	if uid == "" {
		// 列出所有用户
		users := s.memory.ListConversations()
		c.JSON(consts.StatusOK, map[string]interface{}{
			"status": "success",
			"users":  users,
		})
		return
	}

	// 获取指定用户的会话
	conversation := s.memory.GetConversation(uid, false)
	if conversation == nil {
		c.JSON(consts.StatusNotFound, map[string]interface{}{
			"status": "error",
			"error":  "conversation not found",
		})
		return
	}

	messages := conversation.GetFullMessages()
	history := make([]map[string]string, 0, len(messages))
	for _, msg := range messages {
		history = append(history, map[string]string{
			"role":    string(msg.Role),
			"content": msg.Content,
		})
	}

	c.JSON(consts.StatusOK, map[string]interface{}{
		"status":  "success",
		"user_id": uid,
		"history": history,
	})
}

// handleDeleteHistory 删除历史记录
func (s *HTTPServer) handleDeleteHistory(ctx context.Context, c *app.RequestContext) {
	uid := c.Query("uid")

	if uid == "" {
		c.JSON(consts.StatusBadRequest, map[string]interface{}{
			"status": "error",
			"error":  "missing uid parameter",
		})
		return
	}

	if err := s.memory.DeleteConversation(uid); err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	c.JSON(consts.StatusOK, map[string]interface{}{
		"status":  "success",
		"message": fmt.Sprintf("conversation for user %s deleted", uid),
	})
}

// handleHealth 健康检查
func (s *HTTPServer) handleHealth(ctx context.Context, c *app.RequestContext) {
	c.JSON(consts.StatusOK, map[string]interface{}{
		"status":  "healthy",
		"service": "AI Tutor",
	})
}
