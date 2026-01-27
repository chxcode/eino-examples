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

package agents

import (
	"context"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"

	"github.com/cloudwego/eino-examples/adk/multiagent/ai-tutor/tools"
)

// FaultHandlerInstruction 故障处理 Agent 指令
const FaultHandlerInstruction = `你是一位专业的网络故障处理专家，专门帮助在线英语教育平台的学员解决网络相关问题。

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
- 在提供建议后，询问用户是否需要进一步帮助

## 重要提醒
- 你的回复将直接展示给用户
- 不要在回复中包含任何内部报告、状态说明或给 supervisor 的信息
- 只输出给用户看的内容`

// BuildFaultHandlerAgent 构建故障处理 Agent
func BuildFaultHandlerAgent(ctx context.Context, m model.ToolCallingChatModel) (adk.Agent, error) {
	// 获取网络相关工具
	networkTools, err := tools.GetAllNetworkTools()
	if err != nil {
		return nil, err
	}

	return adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "fault_handler",
		Description: "网络故障处理专家，负责处理用户的网络问题，包括网络质量检查、诊断和修复建议。当用户遇到视频卡顿、音频断续、连接不稳定等网络相关问题时，请转交给此专家处理。",
		Instruction: FaultHandlerInstruction,
		Model:       m,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: networkTools,
			},
		},
	})
}
