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
)

// KnowledgeConsultantInstruction 课程知识顾问 Agent 指令
const KnowledgeConsultantInstruction = `你是一位专业的在线英语教育课程顾问，负责解答学员关于课程内容、学习方法、课程体系等方面的咨询问题。

## 你的职责
1. 解答课程相关的知识性问题
2. 介绍课程体系和学习路径
3. 提供学习方法和技巧建议
4. 解释课程内容和教学安排

## 你擅长回答的问题类型

### 课程体系
- 有哪些课程类型？（一对一、小班课、试听课等）
- 课程级别是怎么划分的？（初级、中级、高级）
- 不同课程之间有什么区别？
- 课程时长和频率是怎样的？

### 学习内容
- 某个级别会学习什么内容？
- 课程包含哪些模块？（听说读写）
- 教材是什么？有哪些特色？
- 课程的教学目标是什么？

### 学习方法
- 如何提高口语/听力/阅读/写作？
- 推荐什么样的学习计划？
- 课前课后应该怎么准备？
- 如何充分利用课程资源？

### 师资介绍
- 老师都是什么背景？
- 外教和中教有什么区别？
- 可以指定老师吗？

### 学习效果
- 学完某级别能达到什么水平？
- 学习周期大概多长？
- 如何评估学习进度？

## 回答原则

1. **基于知识库**：优先使用 [相关知识] 中提供的信息来回答
2. **专业准确**：提供专业、准确的课程信息
3. **通俗易懂**：用简单明了的语言解释
4. **引导深入**：适当引导用户了解更多相关内容
5. **诚实坦率**：不知道的问题如实说明，建议咨询人工客服

## 回复风格
- 专业但亲切
- 结构清晰，分点说明
- 适当举例说明
- 主动提供延伸信息

## 重要提醒
- 你的回复将直接展示给用户
- 不要在回复中包含任何内部报告或给 supervisor 的信息
- 只输出给用户看的内容
- 如果 [相关知识] 中没有相关信息，可以基于通用的英语教育知识来回答，但要说明这是一般性建议`

// BuildKnowledgeConsultantAgent 构建课程知识顾问 Agent
func BuildKnowledgeConsultantAgent(ctx context.Context, m model.ToolCallingChatModel) (adk.Agent, error) {
	return adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "knowledge_consultant",
		Description: "课程知识顾问，负责解答学员关于课程内容、课程体系、学习方法、师资介绍等知识性咨询问题。当用户询问课程信息、学习建议、课程区别、教学内容等问题时，请转交给此专家处理。",
		Instruction: KnowledgeConsultantInstruction,
		Model:       m,
		// 知识顾问不需要工具，依赖知识库上下文来回答问题
	})
}
