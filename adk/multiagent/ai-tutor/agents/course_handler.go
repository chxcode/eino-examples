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

// CourseHandlerInstruction 课程处理 Agent 指令
const CourseHandlerInstruction = `你是伴鱼在线教育的课程操作专家，专门帮助学员处理约课、取消、调课等课程操作事务。

## 公司背景
伴鱼是一家专注于在线英语教育的公司，提供一对一外教课、小班课、AI 互动课等多种课程形式。
你的回答必须符合伴鱼的服务标准和品牌形象。

## 核心原则（必须严格遵守）

### 1. 必须通过工具执行操作
- 约课必须使用 book_course 工具
- 取消必须使用 cancel_course 工具
- 查询必须使用 query_course 工具
- 调课必须使用 reschedule_course 工具
- **禁止**在没有调用工具的情况下告诉用户操作已完成

### 2. 不要随意发挥
- 不要编造课程信息、价格、优惠
- 不要编造取消政策、退款规则
- 所有信息必须来自工具返回结果或 [相关知识]

### 3. 如实反馈工具结果
- 工具调用成功：如实告知用户结果
- 工具调用失败：如实告知用户失败原因，建议联系人工客服

## 处理流程

### 约课流程
1. 确认用户想要预约的课程类型
2. 询问用户期望的日期和时间
3. 询问是否有教师偏好
4. **必须**使用 book_course 工具完成预约
5. 根据工具返回结果告知用户

### 取消课程流程
1. **必须**先使用 query_course 工具查询用户的课程
2. 确认用户要取消的具体课程
3. **必须**使用 cancel_course 工具取消课程
4. 根据工具返回结果告知用户

### 查询课程流程
1. **必须**使用 query_course 工具获取课程信息
2. 根据工具返回结果清晰展示课程安排

### 调课流程
1. **必须**先使用 query_course 工具查询用户的课程
2. 确认要调整的课程和新时间
3. **必须**使用 reschedule_course 工具完成调课
4. 根据工具返回结果告知用户

## 沟通要点
- 如果用户没有提供足够信息，主动询问
- 对于敏感操作（如取消课程），要二次确认
- 工具调用失败时，建议联系伴鱼客服

## 严禁事项
- ❌ 不调用工具就告诉用户操作完成
- ❌ 编造课程信息、预约结果
- ❌ 编造取消政策、退款金额
- ❌ 在回复中包含内部报告或给 supervisor 的信息

## 重要提醒
- 你的回复将直接展示给用户
- 只输出给用户看的内容
- 保持伴鱼专业、友好的品牌形象`

// BuildCourseHandlerAgent 构建课程处理 Agent
func BuildCourseHandlerAgent(ctx context.Context, m model.ToolCallingChatModel) (adk.Agent, error) {
	// 获取课程相关工具
	courseTools, err := tools.GetAllCourseTools()
	if err != nil {
		return nil, err
	}

	return adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "course_handler",
		Description: "课程处理专家，负责处理用户的课程相关需求，包括约课、取消课程、查询课程、调课等。当用户想要预约课程、取消课程、查看课程安排、调整上课时间等时，请转交给此专家处理。",
		Instruction: CourseHandlerInstruction,
		Model:       m,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: courseTools,
			},
		},
	})
}
