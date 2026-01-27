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

package server

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"

	"github.com/cloudwego/eino-examples/adk/multiagent/ai-tutor/store"
)

// HTTPServer HTTP 服务
type HTTPServer struct {
	agent adk.Agent
	store store.ConversationStore
	port  string
}

// NewHTTPServer 创建 HTTP 服务
// store 参数为会话存储实现，如果为 nil 则使用内存存储
func NewHTTPServer(agent adk.Agent, conversationStore store.ConversationStore, port string) *HTTPServer {
	if port == "" {
		port = "8080"
	}
	if conversationStore == nil {
		conversationStore = store.NewMemoryStore(20)
	}
	return &HTTPServer{
		agent: agent,
		store: conversationStore,
		port:  port,
	}
}

// ChatRequest 聊天请求
type ChatRequest struct {
	UID     string `json:"uid"`
	Message string `json:"message"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
	UserID  string `json:"user_id,omitempty"`
	Events  []EventInfo `json:"events,omitempty"`
}

// EventInfo 事件信息
type EventInfo struct {
	AgentName string `json:"agent_name,omitempty"`
	Action    string `json:"action,omitempty"`
	Content   string `json:"content,omitempty"`
}

// Start 启动 HTTP 服务
func (s *HTTPServer) Start() error {
	mux := http.NewServeMux()

	// API 路由
	mux.HandleFunc("/api/chat", s.handleChat)
	mux.HandleFunc("/api/history", s.handleHistory)
	mux.HandleFunc("/api/health", s.handleHealth)
	mux.HandleFunc("/", s.handleIndex)

	addr := ":" + s.port
	log.Printf("[AI Tutor ADK] HTTP server starting on %s...\n", addr)
	return http.ListenAndServe(addr, mux)
}

// handleIndex 首页
func (s *HTTPServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	html := `<!DOCTYPE html>
<html>
<head>
    <title>AI 学管助手 (ADK Supervisor)</title>
    <meta charset="utf-8">
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
        h1 { color: #333; }
        .badge { background: #4CAF50; color: white; padding: 2px 8px; border-radius: 4px; font-size: 12px; }
        .endpoint { background: #f5f5f5; padding: 15px; margin: 10px 0; border-radius: 5px; }
        code { background: #e0e0e0; padding: 2px 6px; border-radius: 3px; }
    </style>
</head>
<body>
    <h1>🎓 AI 学管助手 API <span class="badge">ADK Supervisor</span></h1>
    <p>基于 Eino ADK Supervisor 模式实现的多 Agent 系统</p>
    
    <div class="endpoint">
        <h3>GET/POST /api/chat</h3>
        <p>发送聊天消息</p>
        <p>参数：<code>uid</code> - 用户ID，<code>message</code> - 消息内容</p>
        <p>示例：<code>/api/chat?uid=user123&message=我想约课</code></p>
    </div>
    
    <div class="endpoint">
        <h3>GET /api/history</h3>
        <p>获取会话历史</p>
        <p>参数：<code>uid</code> - 用户ID（可选）</p>
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

    <h2>架构特点</h2>
    <ul>
        <li>✅ ADK Supervisor 模式：显式任务委派</li>
        <li>✅ Runner + 事件流：统一执行模型</li>
        <li>✅ Exit Tool：显式控制对话结束</li>
        <li>✅ 支持扩展：可添加 Human-in-the-Loop</li>
    </ul>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

// handleChat 处理聊天请求
func (s *HTTPServer) handleChat(w http.ResponseWriter, r *http.Request) {
	var uid, message string

	if r.Method == http.MethodPost {
		var req ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.jsonResponse(w, http.StatusBadRequest, ChatResponse{
				Status: "error",
				Error:  "invalid request body",
			})
			return
		}
		uid = req.UID
		message = req.Message
	} else {
		uid = r.URL.Query().Get("uid")
		message = r.URL.Query().Get("message")
	}

	if uid == "" || message == "" {
		s.jsonResponse(w, http.StatusBadRequest, ChatResponse{
			Status: "error",
			Error:  "missing uid or message parameter",
		})
		return
	}

	log.Printf("[Chat] uid=%s, message=%s\n", uid, message)

	// 创建 context
	ctx := context.Background()

	// 获取历史消息
	history, err := s.store.GetHistory(ctx, uid, 0)
	if err != nil {
		log.Printf("[Chat] Failed to get history: %v\n", err)
		history = []*schema.Message{} // 出错时使用空历史
	}

	// 构建输入消息
	contextInfo := fmt.Sprintf("[用户ID: %s]", uid)
	fullQuery := contextInfo + "\n\n" + message
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           s.agent,
		EnableStreaming: true, // 必须启用流式模式
	})

	// 将历史消息转换为查询前缀
	queryWithHistory := fullQuery
	if len(history) > 0 {
		var historyStr strings.Builder
		historyStr.WriteString("[历史对话]\n")
		for _, msg := range history {
			historyStr.WriteString(fmt.Sprintf("%s: %s\n", msg.Role, msg.Content))
		}
		historyStr.WriteString("\n[当前问题]\n")
		queryWithHistory = historyStr.String() + fullQuery
	}

	iter := runner.Query(ctx, queryWithHistory)

	// 收集事件和最终消息
	var events []EventInfo
	var finalMessage string
	var subAgentResponses []string // 收集所有子 agent 的实质性回复
	var supervisorExitMessage string // supervisor 退出时的消息

	for {
		event, hasEvent := iter.Next()
		if !hasEvent {
			break
		}

		// 记录事件信息
		eventInfo := EventInfo{}
		eventInfo.AgentName = event.AgentName

		// 详细日志：记录每个事件
		log.Printf("[Event] AgentName=%s, RunPath=%v, HasAction=%v, HasOutput=%v, HasErr=%v\n",
			event.AgentName, event.RunPath, event.Action != nil, event.Output != nil, event.Err != nil)

		// 处理 Action 事件
		if event.Action != nil {
			if event.Action.TransferToAgent != nil {
				eventInfo.Action = fmt.Sprintf("transfer to %s", event.Action.TransferToAgent.DestAgentName)
				log.Printf("[Action] Transfer from %s to %s\n", event.AgentName, event.Action.TransferToAgent.DestAgentName)
			}
			if event.Action.Exit {
				eventInfo.Action = "exit"
				log.Printf("[Action] Exit by %s\n", event.AgentName)
			}
		}

		// 处理 Output 事件 - 使用 adk.GetMessage 获取消息
		if event.Output != nil {
			msg, _, err := adk.GetMessage(event)
			if err == nil {
				// 记录工具调用（发生在内容之前）
				if len(msg.ToolCalls) > 0 {
					for _, tc := range msg.ToolCalls {
						eventInfo.Action = fmt.Sprintf("call tool: %s", tc.Function.Name)
						log.Printf("[ToolCall] Agent=%s, Tool=%s, Args=%s\n",
							event.AgentName, tc.Function.Name, tc.Function.Arguments)
					}
				}

				// 记录消息内容
				if msg.Content != "" {
					log.Printf("[Output] Agent=%s, Content=%s\n", event.AgentName, truncate(msg.Content, 100))
					eventInfo.Content = msg.Content

					// 收集子 agent 的实质性回复（排除 transfer 成功消息）
					if event.AgentName != "ai_tutor_supervisor" {
						// 过滤掉 transfer 工具的成功消息
						if !strings.HasPrefix(msg.Content, "successfully transferred to agent") {
							subAgentResponses = append(subAgentResponses, msg.Content)
							log.Printf("[SubAgent Response] Agent=%s, Content=%s\n", event.AgentName, truncate(msg.Content, 50))
						}
					} else if event.Action != nil && event.Action.Exit {
						// 记录 supervisor 退出时的消息
						supervisorExitMessage = msg.Content
					}
				}
			} else {
				log.Printf("[Output] GetMessage error: %v\n", err)
			}
		}

		// 处理错误
		if event.Err != nil {
			log.Printf("[Error] Agent=%s, Error=%v\n", event.AgentName, event.Err)
		}

		if eventInfo.AgentName != "" || eventInfo.Action != "" || eventInfo.Content != "" {
			events = append(events, eventInfo)
		}
	}

	// 确定最终消息：优先使用子 agent 的实质性回复
	if len(subAgentResponses) > 0 {
		// 使用最后一个子 agent 的回复（通常是最相关的）
		finalMessage = subAgentResponses[len(subAgentResponses)-1]
		log.Printf("[FinalMessage] Using sub-agent response: %s\n", truncate(finalMessage, 50))
	} else if supervisorExitMessage != "" {
		// 如果没有子 agent 回复，使用 supervisor 退出消息
		finalMessage = supervisorExitMessage
		log.Printf("[FinalMessage] Using supervisor exit message: %s\n", truncate(finalMessage, 50))
	}

	// 保存会话
	if err := s.store.AppendMessage(ctx, uid, schema.UserMessage(message)); err != nil {
		log.Printf("[Chat] Failed to save user message: %v\n", err)
	}
	if finalMessage != "" {
		if err := s.store.AppendMessage(ctx, uid, schema.AssistantMessage(finalMessage, nil)); err != nil {
			log.Printf("[Chat] Failed to save assistant message: %v\n", err)
		}
	}

	s.jsonResponse(w, http.StatusOK, ChatResponse{
		Status:  "success",
		Message: finalMessage,
		UserID:  uid,
		Events:  events,
	})
}

// handleHistory 处理历史记录请求
func (s *HTTPServer) handleHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method == http.MethodDelete {
		uid := r.URL.Query().Get("uid")
		if uid == "" {
			s.jsonResponse(w, http.StatusBadRequest, map[string]string{
				"status": "error",
				"error":  "missing uid parameter",
			})
			return
		}
		if err := s.store.Clear(ctx, uid); err != nil {
			s.jsonResponse(w, http.StatusInternalServerError, map[string]string{
				"status": "error",
				"error":  fmt.Sprintf("failed to clear conversation: %v", err),
			})
			return
		}
		s.jsonResponse(w, http.StatusOK, map[string]string{
			"status":  "success",
			"message": fmt.Sprintf("conversation for user %s deleted", uid),
		})
		return
	}

	uid := r.URL.Query().Get("uid")
	if uid == "" {
		// 列出所有用户
		users, err := s.store.ListUsers(ctx)
		if err != nil {
			s.jsonResponse(w, http.StatusInternalServerError, map[string]string{
				"status": "error",
				"error":  fmt.Sprintf("failed to list users: %v", err),
			})
			return
		}
		s.jsonResponse(w, http.StatusOK, map[string]interface{}{
			"status": "success",
			"users":  users,
		})
		return
	}

	// 获取指定用户的历史
	history, err := s.store.GetHistory(ctx, uid, 0)
	if err != nil {
		s.jsonResponse(w, http.StatusInternalServerError, map[string]string{
			"status": "error",
			"error":  fmt.Sprintf("failed to get history: %v", err),
		})
		return
	}

	if len(history) == 0 {
		s.jsonResponse(w, http.StatusOK, map[string]interface{}{
			"status":  "success",
			"user_id": uid,
			"history": []map[string]string{},
		})
		return
	}

	messages := make([]map[string]string, 0, len(history))
	for _, msg := range history {
		messages = append(messages, map[string]string{
			"role":    string(msg.Role),
			"content": msg.Content,
		})
	}

	s.jsonResponse(w, http.StatusOK, map[string]interface{}{
		"status":  "success",
		"user_id": uid,
		"history": messages,
	})
}

// handleHealth 健康检查
func (s *HTTPServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.jsonResponse(w, http.StatusOK, map[string]interface{}{
		"status":  "healthy",
		"service": "AI Tutor (ADK Supervisor)",
	})
}

func (s *HTTPServer) jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// truncate 截断字符串用于日志
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// RunCLI 运行命令行交互模式
// conversationStore 为会话存储实现，如果为 nil 则使用内存存储
func RunCLI(agent adk.Agent, conversationStore store.ConversationStore) {
	ctx := context.Background()
	if conversationStore == nil {
		conversationStore = store.NewMemoryStore(20)
	}
	uid := "cli_user"

	fmt.Println("========================================")
	fmt.Println("    🎓 AI 学管助手 (ADK Supervisor)")
	fmt.Println("========================================")
	fmt.Println("输入 'exit' 退出，'clear' 清空会话历史")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("You: ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		if input == "exit" {
			fmt.Println("再见！👋")
			break
		}

		if input == "clear" {
			_ = conversationStore.Clear(ctx, uid)
			fmt.Println("[会话历史已清空]")
			continue
		}

		// 获取历史
		history, _ := conversationStore.GetHistory(ctx, uid, 0)

		// 创建 Runner
		runner := adk.NewRunner(ctx, adk.RunnerConfig{
			Agent:           agent,
			EnableStreaming: true,
		})

		// 将历史消息附加到查询
		queryWithHistory := input
		if len(history) > 0 {
			var historyStr strings.Builder
			historyStr.WriteString("[历史对话]\n")
			for _, msg := range history {
				historyStr.WriteString(fmt.Sprintf("%s: %s\n", msg.Role, msg.Content))
			}
			historyStr.WriteString("\n[当前问题]\n")
			queryWithHistory = historyStr.String() + input
		}

		// 执行查询
		iter := runner.Query(ctx, queryWithHistory)

		fmt.Println()
		var finalMessage string

		for {
			event, hasEvent := iter.Next()
			if !hasEvent {
				break
			}

			// 打印事件信息
			if event.Action != nil && event.Action.TransferToAgent != nil {
				fmt.Printf("  [%s → %s]\n", event.AgentName, event.Action.TransferToAgent.DestAgentName)
			}

			// 使用 adk.GetMessage 获取消息
			if event.Output != nil {
				msg, _, err := adk.GetMessage(event)
				if err == nil {
					if len(msg.ToolCalls) > 0 {
						for _, tc := range msg.ToolCalls {
							fmt.Printf("  [调用工具: %s]\n", tc.Function.Name)
						}
					}
					if msg.Content != "" {
						finalMessage = msg.Content
					}
				}
			}

			// 打印错误
			if event.Err != nil {
				fmt.Printf("  [错误: %v]\n", event.Err)
			}
		}

		if finalMessage != "" {
			fmt.Printf("\nAssistant: %s\n\n", finalMessage)

			// 保存会话
			_ = conversationStore.AppendMessage(ctx, uid, schema.UserMessage(input))
			_ = conversationStore.AppendMessage(ctx, uid, schema.AssistantMessage(finalMessage, nil))
		}
	}
}
