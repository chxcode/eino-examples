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

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/cloudwego/eino-ext/callbacks/langfuse"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/callbacks"
	"github.com/google/uuid"

	"github.com/cloudwego/eino-examples/flow/agent/ai_tutor/agent"
	"github.com/cloudwego/eino-examples/flow/agent/ai_tutor/config"
	"github.com/cloudwego/eino-examples/flow/agent/ai_tutor/memory"
	"github.com/cloudwego/eino-examples/flow/agent/ai_tutor/server"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 验证必要配置
	if cfg.OpenAIAPIKey == "" {
		log.Fatal("OPENAI_API_KEY is required")
	}

	ctx := context.Background()

	// 初始化 Langfuse（如果配置了）
	if cfg.LangfusePublicKey != "" && cfg.LangfuseSecretKey != "" {
		fmt.Println("[AI Tutor] Langfuse enabled, tracing at:", cfg.LangfuseHost)
		handler, flusher := langfuse.NewLangfuseHandler(&langfuse.Config{
			Host:      cfg.LangfuseHost,
			PublicKey: cfg.LangfusePublicKey,
			SecretKey: cfg.LangfuseSecretKey,
			Name:      "AI Tutor",
			Public:    true,
			Release:   "v1.0.0",
			Tags:      []string{"ai-tutor", "multiagent"},
		})
		callbacks.InitCallbackHandlers([]callbacks.Handler{handler})
		defer flusher()

		// 设置 trace ID
		ctx = langfuse.SetTrace(ctx, langfuse.WithID(uuid.New().String()))
	} else {
		fmt.Println("[AI Tutor] Langfuse not configured, running without tracing")
	}

	// 创建 OpenAI 聊天模型
	var temp float32 = 0.7
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:      cfg.OpenAIAPIKey,
		BaseURL:     cfg.OpenAIBaseURL,
		Model:       cfg.OpenAIModel,
		Temperature: &temp,
	})
	if err != nil {
		log.Fatalf("Failed to create chat model: %v", err)
	}

	// 创建模型提供者
	modelProvider := &agent.ChatModelProvider{
		ToolCallingModel: chatModel, // openai 模型同时实现了 ToolCallingChatModel
		ChatModel:        chatModel, // openai 模型也实现了 ChatModel
	}

	// 创建会话管理器
	mem := memory.NewConversationMemory("data/ai_tutor_memory", 10)
	if mem == nil {
		log.Fatal("Failed to create conversation memory")
	}

	// 创建 AI 学管协调员
	coordinator, err := agent.NewAITutorCoordinator(ctx, modelProvider, cfg)
	if err != nil {
		log.Fatalf("Failed to create AI tutor coordinator: %v", err)
	}

	// 打印配置信息
	printConfig(cfg)

	// 启动 HTTP 服务
	httpServer := server.NewHTTPServer(cfg, coordinator, mem)
	if err := httpServer.Start(); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}

func printConfig(cfg *config.Config) {
	fmt.Println("========================================")
	fmt.Println("        🎓 AI 学管助手启动中...")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Printf("📌 HTTP 端口: %s\n", cfg.HTTPPort)
	fmt.Printf("🤖 OpenAI 模型: %s\n", cfg.OpenAIModel)
	fmt.Printf("🔗 OpenAI BaseURL: %s\n", maskString(cfg.OpenAIBaseURL))

	if cfg.LangfusePublicKey != "" {
		fmt.Printf("📊 Langfuse: 已启用 (%s)\n", cfg.LangfuseHost)
	} else {
		fmt.Println("📊 Langfuse: 未配置")
	}

	if cfg.DifyAPIKey != "" {
		fmt.Printf("📚 Dify 知识库: 已配置 (Dataset: %s)\n", cfg.DifyDataset)
	} else {
		fmt.Println("📚 Dify 知识库: 未配置")
	}

	fmt.Println()
	fmt.Println("🚀 API 接口:")
	fmt.Printf("   - 聊天: GET/POST http://localhost:%s/api/chat\n", cfg.HTTPPort)
	fmt.Printf("   - 流式聊天: GET http://localhost:%s/api/chat/stream\n", cfg.HTTPPort)
	fmt.Printf("   - 历史记录: GET http://localhost:%s/api/history\n", cfg.HTTPPort)
	fmt.Printf("   - 健康检查: GET http://localhost:%s/api/health\n", cfg.HTTPPort)
	fmt.Println()
	fmt.Println("📝 使用示例:")
	fmt.Printf("   curl \"http://localhost:%s/api/chat?uid=user001&message=我想约课\"\n", cfg.HTTPPort)
	fmt.Printf("   curl \"http://localhost:%s/api/chat?uid=user001&message=视频卡顿了\"\n", cfg.HTTPPort)
	fmt.Printf("   curl \"http://localhost:%s/api/chat?uid=user001&message=我学不会,好沮丧\"\n", cfg.HTTPPort)
	fmt.Println()
	fmt.Println("========================================")
}

func maskString(s string) string {
	if len(s) <= 20 {
		return s
	}
	return s[:10] + "..." + s[len(s)-10:]
}

// 环境变量说明
func init() {
	envHelp := `
AI 学管助手 - 环境变量配置说明
==============================

必需配置:
  OPENAI_API_KEY      - OpenAI API 密钥

可选配置:
  OPENAI_BASE_URL     - OpenAI API 地址 (默认: https://api.openai.com/v1)
  OPENAI_MODEL_NAME   - 模型名称 (默认: gpt-4o)
  HTTP_PORT           - HTTP 服务端口 (默认: 8080)

Langfuse 观测 (可选):
  LANGFUSE_HOST       - Langfuse 地址 (默认: https://cloud.langfuse.com)
  LANGFUSE_PUBLIC_KEY - Langfuse 公钥
  LANGFUSE_SECRET_KEY - Langfuse 私钥

Dify 知识库 (可选):
  DIFY_BASE_URL       - Dify API 地址 (默认: https://api.dify.ai/v1)
  DIFY_API_KEY        - Dify API 密钥
  DIFY_DATASET_ID     - Dify 知识库 ID

外部服务 (工具调用):
  NETWORK_CHECK_URL   - 网络检查服务地址 (默认: http://localhost:9000/api/network)
  COURSE_SERVICE_URL  - 课程服务地址 (默认: http://localhost:9001/api/course)
`
	if len(os.Args) > 1 && (os.Args[1] == "-h" || os.Args[1] == "--help") {
		fmt.Println(envHelp)
		os.Exit(0)
	}
}
