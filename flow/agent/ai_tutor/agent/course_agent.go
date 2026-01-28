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
	CourseAgentName = "course_handler"
	CourseAgentDesc = "课程处理专家，负责处理用户的课程相关需求，包括约课、取消课程、查询课程、调课等。当用户想要预约课程、取消课程、查看课程安排、调整上课时间等时，请转交给此专家处理。"
)

// CourseAgentSystemPrompt 课程处理 Agent 的系统提示词
const CourseAgentSystemPrompt = `你是一位专业的课程顾问，专门帮助在线英语教育平台的学员处理课程相关事务。

## 你的职责
1. 帮助用户预约课程
2. 处理课程取消和调整
3. 查询用户的课程信息
4. 提供课程安排建议

## 处理流程

### 约课流程
1. 确认用户想要预约的课程类型（一对一、小班课、试听课）
2. 询问用户期望的日期和时间
3. 询问是否有教师偏好
4. 使用 book_course 工具完成预约
5. 确认预约详情并提醒用户注意事项

### 取消课程流程
1. 使用 query_course 工具查询用户的课程
2. 确认用户要取消的具体课程
3. 使用 cancel_course 工具取消课程
4. 说明取消政策和课时返还情况

### 查询课程流程
1. 询问用户想查询什么（即将开始的、历史的、全部）
2. 使用 query_course 工具获取课程信息
3. 清晰地展示课程安排

### 调课流程
1. 使用 query_course 工具查询用户的课程
2. 确认要调整的课程
3. 询问新的期望时间
4. 使用 reschedule_course 工具完成调课
5. 确认新的课程安排

## 沟通要点
- 始终确认用户的需求是否被正确理解
- 提供清晰的课程信息
- 主动提醒重要事项（如上课时间、教室链接等）
- 推荐合适的课程安排

## 注意事项
- 如果用户没有提供足够信息，主动询问
- 对于敏感操作（如取消课程），要二次确认
- 如果遇到无法处理的情况，说明原因并建议联系人工客服`

// NewCourseHandlerSpecialist 创建课程处理 Agent
func NewCourseHandlerSpecialist(ctx context.Context, chatModel model.ToolCallingChatModel, courseServiceURL string) (*host.Specialist, error) {
	// 获取课程相关工具
	tools := aitool.GetCourseTools(courseServiceURL)

	// 创建 React Agent
	rAgent, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: chatModel,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: convertToolsToBase(tools),
		},
		MaxStep: 10,
	})
	if err != nil {
		return nil, err
	}

	return &host.Specialist{
		AgentMeta: host.AgentMeta{
			Name:        CourseAgentName,
			IntendedUse: CourseAgentDesc,
		},
		Invokable: func(ctx context.Context, input []*schema.Message, opts ...agent.AgentOption) (*schema.Message, error) {
			// 添加系统提示词
			messages := append([]*schema.Message{
				schema.SystemMessage(CourseAgentSystemPrompt),
			}, input...)

			return rAgent.Generate(ctx, messages, opts...)
		},
		Streamable: func(ctx context.Context, input []*schema.Message, opts ...agent.AgentOption) (*schema.StreamReader[*schema.Message], error) {
			// 添加系统提示词
			messages := append([]*schema.Message{
				schema.SystemMessage(CourseAgentSystemPrompt),
			}, input...)

			return rAgent.Stream(ctx, messages, opts...)
		},
	}, nil
}

func convertToolsToBase(tools []tool.BaseTool) []tool.BaseTool {
	return tools
}
