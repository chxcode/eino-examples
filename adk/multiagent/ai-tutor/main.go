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

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/cloudwego/eino-ext/callbacks/langfuse"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"
	"github.com/google/uuid"

	"github.com/cloudwego/eino-examples/adk/multiagent/ai-tutor/agents"
	"github.com/cloudwego/eino-examples/adk/multiagent/ai-tutor/server"
)

func main() {
	// 解析命令行参数
	mode := flag.String("mode", "http", "运行模式: http (HTTP服务) 或 cli (命令行交互)")
	port := flag.String("port", "", "HTTP 端口 (默认 8080)")
	flag.Parse()

	if *port == "" {
		*port = os.Getenv("HTTP_PORT")
		if *port == "" {
			*port = "8080"
		}
	}

	ctx := context.Background()

	// 初始化 Langfuse（如果配置了）
	initLangfuse(ctx)

	// 创建 ChatModel - 兼容 OPENAI_MODEL 和 OPENAI_MODEL_NAME 两种环境变量
	chatModel := newChatModel(ctx)

	// 构建 AI 学管 Supervisor
	supervisor, err := agents.BuildAITutorSupervisor(ctx, chatModel)
	if err != nil {
		log.Fatalf("Failed to build AI tutor supervisor: %v", err)
	}

	// 打印启动信息
	printStartupInfo(*mode, *port)

	// 根据模式运行
	switch *mode {
	case "cli":
		server.RunCLI(supervisor)
	case "http":
		httpServer := server.NewHTTPServer(supervisor, *port)
		if err := httpServer.Start(); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	default:
		log.Fatalf("Unknown mode: %s", *mode)
	}
}

// newChatModel 创建 ChatModel，兼容 Host-Specialist 版本的环境变量命名
func newChatModel(ctx context.Context) model.ToolCallingChatModel {
	// 获取模型名称，优先使用 OPENAI_MODEL，其次使用 OPENAI_MODEL_NAME（与 Host-Specialist 版本兼容）
	modelName := os.Getenv("OPENAI_MODEL")
	if modelName == "" {
		modelName = os.Getenv("OPENAI_MODEL_NAME")
	}
	if modelName == "" {
		modelName = "gpt-4o" // 默认值
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	baseURL := os.Getenv("OPENAI_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	byAzure := os.Getenv("OPENAI_BY_AZURE") == "true"

	log.Printf("[AI Tutor ADK] Model config: model=%s, baseURL=%s, byAzure=%v\n", modelName, baseURL, byAzure)

	cm, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  apiKey,
		Model:   modelName,
		BaseURL: baseURL,
		ByAzure: byAzure,
	})
	if err != nil {
		log.Fatalf("Failed to create OpenAI ChatModel: %v", err)
	}
	return cm
}

func initLangfuse(ctx context.Context) {
	publicKey := os.Getenv("LANGFUSE_PUBLIC_KEY")
	secretKey := os.Getenv("LANGFUSE_SECRET_KEY")

	if publicKey != "" && secretKey != "" {
		host := os.Getenv("LANGFUSE_HOST")
		if host == "" {
			host = "https://cloud.langfuse.com"
		}

		fmt.Printf("[AI Tutor ADK] Langfuse enabled: %s\n", host)

		handler, _ := langfuse.NewLangfuseHandler(&langfuse.Config{
			Host:      host,
			PublicKey: publicKey,
			SecretKey: secretKey,
			Name:      "AI Tutor (ADK)",
			Public:    true,
			Release:   "v1.0.0",
			Tags:      []string{"ai-tutor", "adk", "supervisor"},
		})
		callbacks.InitCallbackHandlers([]callbacks.Handler{handler})

		// 设置 trace ID
		_ = langfuse.SetTrace(ctx, langfuse.WithID(uuid.New().String()))
	} else {
		fmt.Println("[AI Tutor ADK] Langfuse not configured")
	}
}

func printStartupInfo(mode, port string) {
	fmt.Println("========================================")
	fmt.Println("  🎓 AI 学管助手 (ADK Supervisor 版本)")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("📌 架构特点：")
	fmt.Println("   - ADK Supervisor 模式")
	fmt.Println("   - Runner + 事件流执行模型")
	fmt.Println("   - 显式任务委派 (Transfer)")
	fmt.Println("   - Exit Tool 控制对话结束")
	fmt.Println()
	fmt.Println("👥 Agent 团队：")
	fmt.Println("   - ai_tutor_supervisor: 学管协调员")
	fmt.Println("   - fault_handler: 网络故障处理专家")
	fmt.Println("   - emotion_support: 情绪安抚专家")
	fmt.Println("   - course_handler: 课程处理专家")
	fmt.Println()

	if mode == "http" {
		fmt.Printf("🚀 HTTP 服务端口: %s\n", port)
		fmt.Println()
		fmt.Println("📝 API 接口:")
		fmt.Printf("   - 聊天: GET/POST http://localhost:%s/api/chat\n", port)
		fmt.Printf("   - 历史: GET http://localhost:%s/api/history\n", port)
		fmt.Printf("   - 健康: GET http://localhost:%s/api/health\n", port)
		fmt.Println()
		fmt.Println("📝 使用示例:")
		fmt.Printf("   curl \"http://localhost:%s/api/chat?uid=user001&message=我想约课\"\n", port)
		fmt.Printf("   curl \"http://localhost:%s/api/chat?uid=user001&message=视频卡顿了\"\n", port)
		fmt.Printf("   curl \"http://localhost:%s/api/chat?uid=user001&message=学不会好沮丧\"\n", port)
	} else {
		fmt.Println("🖥️  命令行交互模式")
	}
	fmt.Println()
	fmt.Println("========================================")
}

func init() {
	envHelp := `
AI 学管助手 (ADK Supervisor) - 环境变量配置
==========================================

必需配置 (OpenAI):
  OPENAI_API_KEY      - OpenAI API 密钥
  OPENAI_MODEL        - 模型名称 (如 gpt-4o)
  OPENAI_BASE_URL     - API 地址 (可选)
  OPENAI_BY_AZURE     - 是否使用 Azure (可选，true/false)

或者使用 ARK 模型:
  MODEL_TYPE          - 设置为 "ark"
  ARK_API_KEY         - ARK API 密钥
  ARK_MODEL           - ARK 模型名称

Langfuse 观测 (可选):
  LANGFUSE_HOST       - Langfuse 地址
  LANGFUSE_PUBLIC_KEY - Langfuse 公钥
  LANGFUSE_SECRET_KEY - Langfuse 私钥

HTTP 服务:
  HTTP_PORT           - HTTP 端口 (默认: 8080)

命令行参数:
  -mode http|cli      - 运行模式
  -port 8080          - HTTP 端口
`
	if len(os.Args) > 1 && (os.Args[1] == "-h" || os.Args[1] == "--help") {
		fmt.Println(envHelp)
		os.Exit(0)
	}
}
