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

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent"
	"github.com/cloudwego/eino/flow/agent/multiagent/host"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"

	aitool "github.com/cloudwego/eino-examples/flow/agent/ai_tutor/tool"
)

const (
	FaultHandlerAgentName = "fault_handler"
	FaultHandlerAgentDesc = "网络故障处理专家，负责处理用户的网络问题，包括网络质量检查、诊断和修复建议。当用户遇到视频卡顿、音频断续、连接不稳定等网络相关问题时，请转交给此专家处理。"
)

// FaultHandlerAgentSystemPrompt 故障处理 Agent 的系统提示词
const FaultHandlerAgentSystemPrompt = `你是一位专业的网络故障处理专家，专门帮助在线英语教育平台的学员解决网络相关问题。

## 你的职责
1. 诊断用户的网络问题
2. 提供专业的网络检查和诊断
3. 给出清晰、易懂的解决方案
4. 确保用户能够顺利进行在线学习

## 处理流程
1. 首先使用 check_network_quality 工具检查用户的网络质量
2. 根据检查结果，如果发现问题，使用 diagnose_network 工具进行详细诊断
3. 根据诊断结果，使用 get_network_fix_suggestion 工具获取修复建议
4. 向用户清晰地解释问题原因和解决步骤

## 沟通要点
- 使用简单易懂的语言，避免过多技术术语
- 提供具体的操作步骤
- 表达理解和同理心
- 如果问题无法远程解决，建议用户联系技术支持

## 注意事项
- 保持耐心和专业
- 确保用户理解每一个步骤
- 在提供建议后，询问用户是否需要进一步帮助`

// NewFaultHandlerSpecialist 创建故障处理 Agent
func NewFaultHandlerSpecialist(ctx context.Context, chatModel model.ToolCallingChatModel, networkServiceURL string) (*host.Specialist, error) {
	// 获取网络相关工具
	tools := aitool.GetNetworkTools(networkServiceURL)

	// 创建 React Agent
	rAgent, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: chatModel,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: convertToBaseTools(tools),
		},
		MaxStep: 10,
	})
	if err != nil {
		return nil, err
	}

	// 创建 Specialist
	return &host.Specialist{
		AgentMeta: host.AgentMeta{
			Name:        FaultHandlerAgentName,
			IntendedUse: FaultHandlerAgentDesc,
		},
		Invokable: func(ctx context.Context, input []*schema.Message, opts ...agent.AgentOption) (*schema.Message, error) {
			// 添加系统提示词
			messages := append([]*schema.Message{
				schema.SystemMessage(FaultHandlerAgentSystemPrompt),
			}, input...)

			return rAgent.Generate(ctx, messages, opts...)
		},
		Streamable: func(ctx context.Context, input []*schema.Message, opts ...agent.AgentOption) (*schema.StreamReader[*schema.Message], error) {
			// 添加系统提示词
			messages := append([]*schema.Message{
				schema.SystemMessage(FaultHandlerAgentSystemPrompt),
			}, input...)

			return rAgent.Stream(ctx, messages, opts...)
		},
	}, nil
}

func convertToBaseTools(tools []tool.BaseTool) []tool.BaseTool {
	return tools
}
