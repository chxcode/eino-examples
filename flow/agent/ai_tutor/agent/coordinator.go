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

package agent

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/flow/agent/multiagent/host"
	"github.com/cloudwego/eino/schema"

	"github.com/cloudwego/eino-examples/flow/agent/ai_tutor/config"
	"github.com/cloudwego/eino-examples/flow/agent/ai_tutor/retriever"
)

// ChatModelProvider 提供不同类型的 ChatModel
type ChatModelProvider struct {
	// ToolCallingModel 用于需要工具调用的 Agent
	ToolCallingModel model.ToolCallingChatModel
	// ChatModel 用于普通对话的 Agent
	ChatModel model.ChatModel
}

// CoordinatorSystemPrompt 协调员 Agent 的系统提示词
const CoordinatorSystemPrompt = `你是 AI 学管助手，一位专业的在线英语教育学习顾问。你的职责是帮助学员解决学习过程中遇到的各种问题。

## 你的角色
作为学管助手，你需要：
1. 理解学员的问题和需求
2. 将问题分配给合适的专家处理
3. 提供专业、友好的服务

## 你可以调用的专家团队
1. **网络故障处理专家** (fault_handler)：处理视频卡顿、音频断续、连接问题等网络相关问题
2. **情绪安抚专家** (emotion_support)：处理学习焦虑、考试压力、服务不满等情绪问题
3. **课程处理专家** (course_handler)：处理约课、取消课、查询课程、调课等课程相关需求

## 问题分类指南

### 转交给 fault_handler（网络故障处理专家）的情况：
- 用户说"视频卡"、"画面卡顿"、"看不清"
- 用户说"听不清"、"声音断断续续"、"音频有问题"
- 用户说"连接不上"、"进不去教室"、"掉线"
- 用户说"网络慢"、"加载很久"
- 任何与网络质量相关的问题

### 转交给 emotion_support（情绪安抚专家）的情况：
- 用户表达焦虑、沮丧、不开心等负面情绪
- 用户说"学不会"、"太难了"、"想放弃"
- 用户对服务不满、投诉、抱怨
- 用户压力大、紧张、害怕考试
- 用户需要鼓励和支持

### 转交给 course_handler（课程处理专家）的情况：
- 用户想要约课、预约课程
- 用户想要取消课程、退课
- 用户想要查询课程安排
- 用户想要调整上课时间、调课
- 用户询问剩余课时、套餐信息

### 由你直接处理的情况：
- 一般性的问题咨询
- 学习建议和方法
- 平台功能介绍
- 其他不属于以上类别的问题

## 使用知识库
在回答问题时，会自动检索相关的知识库内容。请参考知识库中的信息来回答用户问题。

## 沟通风格
- 友好、专业、耐心
- 使用简单易懂的语言
- 适当使用表情符号增加亲和力
- 始终保持积极乐观的态度

## 注意事项
- 如果不确定问题类型，可以先询问用户更多信息
- 确保用户问题得到妥善处理
- 如果专家无法解决，引导用户联系人工客服
- 记住用户的上下文，提供连贯的服务`

// AITutorCoordinator AI 学管协调员
type AITutorCoordinator struct {
	multiAgent    *host.MultiAgent
	difyRetriever *retriever.DifyRetriever
	config        *config.Config
}

// NewAITutorCoordinator 创建 AI 学管协调员
func NewAITutorCoordinator(ctx context.Context, modelProvider *ChatModelProvider, cfg *config.Config) (*AITutorCoordinator, error) {
	// 创建 Dify 知识库检索器
	var difyRet *retriever.DifyRetriever
	if cfg.DifyAPIKey != "" && cfg.DifyDataset != "" {
		var err error
		difyRet, err = retriever.NewDifyRetriever(&retriever.DifyRetrieverConfig{
			BaseURL:   cfg.DifyBaseURL,
			APIKey:    cfg.DifyAPIKey,
			DatasetID: cfg.DifyDataset,
			TopK:      5,
		})
		if err != nil {
			fmt.Printf("[Warning] Failed to create Dify retriever: %v\n", err)
		}
	}

	// 创建各个专家 Agent
	faultHandler, err := NewFaultHandlerSpecialist(ctx, modelProvider.ToolCallingModel, cfg.NetworkCheckURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create fault handler: %w", err)
	}

	emotionSupport, err := NewEmotionSupportSpecialist(ctx, modelProvider.ChatModel)
	if err != nil {
		return nil, fmt.Errorf("failed to create emotion support: %w", err)
	}

	courseHandler, err := NewCourseHandlerSpecialist(ctx, modelProvider.ToolCallingModel, cfg.CourseServiceURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create course handler: %w", err)
	}

	// 创建 Host（协调员）
	hostAgent := &host.Host{
		ChatModel:    modelProvider.ChatModel,
		SystemPrompt: CoordinatorSystemPrompt,
	}

	// 创建 MultiAgent
	multiAgent, err := host.NewMultiAgent(ctx, &host.MultiAgentConfig{
		Host: *hostAgent,
		Specialists: []*host.Specialist{
			faultHandler,
			emotionSupport,
			courseHandler,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create multi-agent: %w", err)
	}

	return &AITutorCoordinator{
		multiAgent:    multiAgent,
		difyRetriever: difyRet,
		config:        cfg,
	}, nil
}

// Process 处理用户消息
func (c *AITutorCoordinator) Process(ctx context.Context, userID string, message string, history []*schema.Message) (*schema.Message, error) {
	// 构建输入消息
	messages := make([]*schema.Message, 0, len(history)+2)

	// 添加历史消息
	messages = append(messages, history...)

	// 检索知识库（如果配置了）
	if c.difyRetriever != nil {
		knowledge, err := c.difyRetriever.RetrieveWithFormatted(ctx, message)
		if err == nil && knowledge != "" {
			// 将知识库内容作为系统消息添加
			messages = append(messages, schema.SystemMessage(fmt.Sprintf("以下是相关的知识库内容，请参考：\n\n%s", knowledge)))
		}
	}

	// 添加用户 ID 信息
	contextInfo := fmt.Sprintf("[用户ID: %s]", userID)
	userMessage := schema.UserMessage(contextInfo + "\n\n" + message)
	messages = append(messages, userMessage)

	// 调用 MultiAgent
	return c.multiAgent.Generate(ctx, messages)
}

// ProcessStream 流式处理用户消息
func (c *AITutorCoordinator) ProcessStream(ctx context.Context, userID string, message string, history []*schema.Message) (*schema.StreamReader[*schema.Message], error) {
	// 构建输入消息
	messages := make([]*schema.Message, 0, len(history)+2)

	// 添加历史消息
	messages = append(messages, history...)

	// 检索知识库（如果配置了）
	if c.difyRetriever != nil {
		knowledge, err := c.difyRetriever.RetrieveWithFormatted(ctx, message)
		if err == nil && knowledge != "" {
			messages = append(messages, schema.SystemMessage(fmt.Sprintf("以下是相关的知识库内容，请参考：\n\n%s", knowledge)))
		}
	}

	// 添加用户 ID 信息
	contextInfo := fmt.Sprintf("[用户ID: %s]", userID)
	userMessage := schema.UserMessage(contextInfo + "\n\n" + message)
	messages = append(messages, userMessage)

	// 调用 MultiAgent 流式处理
	return c.multiAgent.Stream(ctx, messages)
}

// GetMultiAgent 获取 MultiAgent 实例（用于导出图等操作）
func (c *AITutorCoordinator) GetMultiAgent() *host.MultiAgent {
	return c.multiAgent
}
