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
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent"
	"github.com/cloudwego/eino/flow/agent/multiagent/host"
	"github.com/cloudwego/eino/schema"
)

const (
	EmotionAgentName = "emotion_support"
	EmotionAgentDesc = "情绪安抚专家，负责处理用户的情绪问题，包括学习焦虑、考试压力、对课程不满等。当用户表现出情绪低落、焦虑、不满、沮丧等情绪时，请转交给此专家处理。"
)

// EmotionAgentSystemPrompt 情绪安抚 Agent 的系统提示词
const EmotionAgentSystemPrompt = `你是一位富有同理心的情绪支持专家，专门帮助在线英语教育平台的学员处理情绪问题。

## 你的职责
1. 倾听和理解用户的情绪
2. 提供温暖、真诚的情感支持
3. 帮助用户缓解学习压力和焦虑
4. 给予积极的鼓励和建议

## 沟通原则
1. **共情优先**：首先表达对用户感受的理解
2. **积极倾听**：让用户感到被听到和被重视
3. **正向引导**：帮助用户看到积极的一面
4. **具体建议**：提供实际可行的建议

## 常见场景处理

### 学习焦虑
- 肯定用户的努力
- 分享学习是一个过程，每个人都有自己的节奏
- 建议分解目标，小步前进
- 推荐放松和减压的方法

### 考试压力
- 理解考试带来的压力
- 分享一些考前准备和放松技巧
- 强调过程比结果更重要
- 鼓励用户相信自己的准备

### 对服务不满
- 真诚道歉
- 认真听取反馈
- 表示会改进
- 提供可能的补偿方案或解决方案

### 学习受挫
- 告诉用户遇到困难是正常的
- 分享一些名人学习英语的励志故事
- 建议调整学习方法
- 鼓励坚持下去

## 语言风格
- 温暖、真诚
- 使用鼓励性的语言
- 避免说教式的建议
- 适当使用表情符号增加亲和力

## 注意事项
- 如果用户情绪非常严重，建议寻求专业心理帮助
- 不要做出无法兑现的承诺
- 尊重用户的感受，不要否定他们的情绪`

// NewEmotionSupportSpecialist 创建情绪安抚 Agent
func NewEmotionSupportSpecialist(ctx context.Context, chatModel model.ChatModel) (*host.Specialist, error) {
	// 创建一个简单的对话链
	chain := compose.NewChain[[]*schema.Message, *schema.Message]()
	chain.AppendLambda(compose.InvokableLambda(func(ctx context.Context, input []*schema.Message) ([]*schema.Message, error) {
		// 添加系统提示词
		return append([]*schema.Message{schema.SystemMessage(EmotionAgentSystemPrompt)}, input...), nil
	})).AppendChatModel(chatModel)

	runner, err := chain.Compile(ctx)
	if err != nil {
		return nil, err
	}

	// 创建流式版本
	streamChain := compose.NewChain[[]*schema.Message, *schema.Message]()
	streamChain.AppendLambda(compose.InvokableLambda(func(ctx context.Context, input []*schema.Message) ([]*schema.Message, error) {
		return append([]*schema.Message{schema.SystemMessage(EmotionAgentSystemPrompt)}, input...), nil
	})).AppendChatModel(chatModel)

	streamRunner, err := streamChain.Compile(ctx)
	if err != nil {
		return nil, err
	}

	return &host.Specialist{
		AgentMeta: host.AgentMeta{
			Name:        EmotionAgentName,
			IntendedUse: EmotionAgentDesc,
		},
		Invokable: func(ctx context.Context, input []*schema.Message, opts ...agent.AgentOption) (*schema.Message, error) {
			return runner.Invoke(ctx, input, agent.GetComposeOptions(opts...)...)
		},
		Streamable: func(ctx context.Context, input []*schema.Message, opts ...agent.AgentOption) (*schema.StreamReader[*schema.Message], error) {
			return streamRunner.Stream(ctx, input, agent.GetComposeOptions(opts...)...)
		},
	}, nil
}
