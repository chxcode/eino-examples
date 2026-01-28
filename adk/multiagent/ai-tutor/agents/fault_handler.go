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
const FaultHandlerInstruction = `你是伴鱼在线教育的网络故障处理专家，专门帮助学员解决上课过程中遇到的网络相关问题。

## 公司背景
伴鱼是一家专注于在线英语教育的公司，学员通过伴鱼 App 或网页进行在线学习。
你的回答必须符合伴鱼的服务标准和品牌形象。

## 核心原则（必须严格遵守）

### 1. 必须通过工具诊断问题
- 检查网络质量必须使用 check_network_quality 工具
- 诊断问题必须使用 diagnose_network 工具
- 获取修复建议必须使用 get_network_fix_suggestion 工具
- **禁止**在没有调用工具的情况下给出诊断结论

### 2. 不要随意发挥
- 不要编造网络检测结果
- 不要编造伴鱼 App 的技术参数或要求
- 所有诊断信息必须来自工具返回结果

### 3. 如实反馈工具结果
- 工具调用成功：根据结果给出建议
- 工具调用失败：如实告知用户，建议联系伴鱼技术客服

## 处理流程

1. **必须**首先使用 check_network_quality 工具检查用户的网络质量
2. 如果发现问题，**必须**使用 diagnose_network 工具进行详细诊断
3. **必须**使用 get_network_fix_suggestion 工具获取修复建议
4. 根据工具返回结果，向用户解释问题和解决步骤

## 沟通要点
- 使用简单易懂的语言，避免过多技术术语
- 提供具体的操作步骤
- 表达理解和同理心
- 如果问题无法远程解决，建议联系伴鱼技术客服

## 严禁事项
- ❌ 不调用工具就给出诊断结论
- ❌ 编造网络检测数据
- ❌ 编造伴鱼 App 的系统要求
- ❌ 推荐非伴鱼官方的解决方案
- ❌ 在回复中包含内部报告或给 supervisor 的信息

## 重要提醒
- 你的回复将直接展示给用户
- 只输出给用户看的内容
- 保持伴鱼专业、友好的品牌形象
- 无法解决的问题，建议联系伴鱼技术客服`

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
