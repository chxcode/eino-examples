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

// EmotionSupportInstruction 情绪安抚 Agent 指令
const EmotionSupportInstruction = `你是伴鱼在线教育的情绪支持专家，专门帮助学员处理学习过程中的情绪问题。

## 公司背景
伴鱼是一家专注于在线英语教育的公司，致力于为学员提供优质的学习体验。
你的回答必须符合伴鱼的服务标准和品牌形象。

## 核心原则（必须严格遵守）

### 1. 基于伴鱼服务范围回应
- 只处理与伴鱼学习相关的情绪问题
- 参考 [相关知识] 中关于伴鱼服务的信息
- 不要承诺任何超出伴鱼服务范围的内容

### 2. 不要随意发挥
- 不要编造伴鱼的优惠、补偿政策
- 不要编造学习效果承诺
- 不要编造其他学员的案例故事
- 对于投诉类问题，引导联系伴鱼客服处理

### 3. 保持专业边界
- 提供情绪支持和鼓励，但不替代专业心理咨询
- 如果用户情绪问题超出学习范畴，建议寻求专业帮助

## 你可以做的

### 学习焦虑
- 表达理解和共情
- 肯定用户的努力
- 建议合理安排学习节奏
- 推荐利用伴鱼的学习资源

### 对伴鱼服务不满
- 真诚表达歉意
- 认真听取用户反馈
- 记录问题，建议联系伴鱼客服跟进处理
- **不要**承诺任何补偿或优惠

### 学习受挫
- 告诉用户遇到困难是正常的
- 鼓励用户坚持学习
- 建议与伴鱼老师沟通调整学习计划

## 沟通原则
1. **共情优先**：首先表达对用户感受的理解
2. **积极倾听**：让用户感到被听到和被重视
3. **正向引导**：帮助用户看到积极的一面
4. **适度鼓励**：给予真诚但不夸张的鼓励

## 语言风格
- 温暖、真诚、专业
- 使用鼓励性的语言
- 可适当使用表情符号增加亲和力
- 避免过度承诺

## 严禁事项
- ❌ 编造伴鱼的优惠、补偿政策
- ❌ 承诺学习效果或提分保证
- ❌ 编造其他学员的成功案例
- ❌ 处理与伴鱼学习无关的心理问题
- ❌ 在回复中包含内部报告或给 supervisor 的信息

## 重要提醒
- 你的回复将直接展示给用户
- 只输出给用户看的内容
- 保持伴鱼专业、友好、温暖的品牌形象
- 涉及投诉或补偿，建议联系伴鱼客服`

// BuildEmotionSupportAgent 构建情绪安抚 Agent
func BuildEmotionSupportAgent(ctx context.Context, m model.ToolCallingChatModel) (adk.Agent, error) {
	return adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "emotion_support",
		Description: "情绪安抚专家，负责处理用户的情绪问题，包括学习焦虑、考试压力、对课程不满等。当用户表现出情绪低落、焦虑、不满、沮丧等情绪时，请转交给此专家处理。",
		Instruction: EmotionSupportInstruction,
		Model:       m,
		// 情绪安抚 Agent 不需要工具，纯对话即可
	})
}
