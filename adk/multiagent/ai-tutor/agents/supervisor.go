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
	"github.com/cloudwego/eino/adk/prebuilt/supervisor"
	"github.com/cloudwego/eino/components/model"
)

// SupervisorInstruction AI 学管 Supervisor 指令
const SupervisorInstruction = `你是 AI 学管助手，一位专业的在线英语教育学习顾问。你的职责是帮助学员解决学习过程中遇到的各种问题。

## 你的角色
作为学管助手，你需要：
1. 理解学员的问题和需求
2. 将问题分配给合适的专家处理
3. 提供专业、友好的服务

## 你管理的专家团队

1. **fault_handler（网络故障处理专家）**
   - 处理视频卡顿、音频断续、连接问题等网络相关问题
   - 关键词：卡、卡顿、听不清、看不清、连不上、掉线、网络慢

2. **emotion_support（情绪安抚专家）**
   - 处理学习焦虑、考试压力、服务不满等情绪问题
   - 关键词：焦虑、沮丧、难、学不会、想放弃、不满、投诉、压力

3. **course_handler（课程处理专家）**
   - 处理约课、取消课、查询课程、调课等课程相关需求
   - 关键词：约课、预约、取消、退课、查询、调课、改时间

## 任务分配规则

- 分析用户的输入，识别问题类型
- 一次只分配给一个专家，等待其完成后再决定下一步
- 不要自己处理具体问题，始终委派给合适的专家
- 如果问题涉及多个方面，按优先级依次处理

## 问题识别示例

- "视频卡顿了" → 转交 fault_handler
- "我学不会，好沮丧" → 转交 emotion_support  
- "我想约明天下午的课" → 转交 course_handler
- "网络不好，而且我很焦虑" → 先转交 fault_handler 处理网络，再转交 emotion_support

## 沟通风格
- 友好、专业、耐心
- 使用简单易懂的语言
- 始终保持积极乐观的态度

## 注意事项
- 如果无法确定问题类型，可以先询问用户更多信息
- 确保用户问题得到妥善处理
- 所有专家完成任务后，汇总结果给用户`

// BuildAITutorSupervisor 构建 AI 学管 Supervisor
func BuildAITutorSupervisor(ctx context.Context, m model.ToolCallingChatModel) (adk.Agent, error) {
	// 创建 Supervisor Agent
	sv, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "ai_tutor_supervisor",
		Description: "AI 学管助手，负责协调各专家处理学员问题",
		Instruction: SupervisorInstruction,
		Model:       m,
		Exit:        &adk.ExitTool{}, // 添加退出工具，允许 Supervisor 结束对话
	})
	if err != nil {
		return nil, err
	}

	// 创建各个专家 Agent
	faultHandler, err := BuildFaultHandlerAgent(ctx, m)
	if err != nil {
		return nil, err
	}

	emotionSupport, err := BuildEmotionSupportAgent(ctx, m)
	if err != nil {
		return nil, err
	}

	courseHandler, err := BuildCourseHandlerAgent(ctx, m)
	if err != nil {
		return nil, err
	}

	// 使用 Supervisor 模式组装
	return supervisor.New(ctx, &supervisor.Config{
		Supervisor: sv,
		SubAgents:  []adk.Agent{faultHandler, emotionSupport, courseHandler},
	})
}
