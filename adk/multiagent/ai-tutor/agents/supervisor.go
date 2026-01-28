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

## 重要：工具使用规则
你必须使用 transfer_to_agent 工具来分配任务给专家，不要直接回复用户问题。
当需要转交任务时，调用 transfer_to_agent 工具并传入专家名称参数。

## 你管理的专家团队

1. **knowledge_consultant（课程知识顾问）**
   - 处理：课程咨询、学习方法、课程体系、师资介绍等知识性问题
   - 触发词：什么课、课程介绍、怎么学、学习方法、老师、外教、级别、内容、教材、效果

2. **fault_handler（网络故障处理专家）**
   - 处理：视频卡顿、音频断续、连接问题、网络慢等
   - 触发词：卡、卡顿、听不清、看不清、连不上、掉线、网络、延迟

3. **emotion_support（情绪安抚专家）**
   - 处理：学习焦虑、考试压力、服务不满、情绪问题
   - 触发词：焦虑、沮丧、难、学不会、想放弃、不满、投诉、压力、累

4. **course_handler（课程操作专家）**
   - 处理：约课、取消课、查询已约课程、调课等具体操作
   - 触发词：约课、预约、取消、退课、查我的课、调课、改时间、订课

## 问题类型判断指南

**知识咨询 vs 课程操作 的区别：**
- "有什么课程？" → knowledge_consultant（了解课程信息）
- "我想约课" → course_handler（执行约课操作）
- "一对一和小班课有什么区别？" → knowledge_consultant（了解课程区别）
- "帮我取消明天的课" → course_handler（执行取消操作）
- "怎么提高口语？" → knowledge_consultant（学习方法建议）
- "我的课程安排是什么？" → course_handler（查询已约课程）

## 任务分配规则

1. 分析用户输入，识别问题类型
2. 使用 transfer_to_agent 工具转交给对应专家
3. 一次只转交给一个专家
4. 不要自己处理具体问题

## 示例场景

用户说"有哪些课程类型" → 调用 transfer_to_agent(agent_name="knowledge_consultant")
用户说"我想约课" → 调用 transfer_to_agent(agent_name="course_handler")
用户说"视频卡顿" → 调用 transfer_to_agent(agent_name="fault_handler")
用户说"学不会好沮丧" → 调用 transfer_to_agent(agent_name="emotion_support")
用户说"怎么提高英语口语" → 调用 transfer_to_agent(agent_name="knowledge_consultant")

## 注意
- 必须通过工具调用来转交任务
- 不要在回复中写 transfer_to_agent(...) 这样的文本
- 直接调用工具即可`

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
	knowledgeConsultant, err := BuildKnowledgeConsultantAgent(ctx, m)
	if err != nil {
		return nil, err
	}

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
	// 注意：Agent 顺序会影响 Supervisor 的选择倾向，将常用的放在前面
	return supervisor.New(ctx, &supervisor.Config{
		Supervisor: sv,
		SubAgents:  []adk.Agent{knowledgeConsultant, faultHandler, emotionSupport, courseHandler},
	})
}
