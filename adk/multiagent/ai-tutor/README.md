# AI 学管助手 (ADK Supervisor 版本)

基于 Eino ADK Supervisor 模式实现的多 Agent 编排 AI 学管系统，专为在线英语教育场景设计。

## 与 Host-Specialist 版本的区别

本项目有两种实现方式：

| 特性 | Host-Specialist (`flow/agent/ai_tutor`) | ADK Supervisor (本目录) |
|------|----------------------------------------|------------------------|
| 框架 | `eino/flow/agent/multiagent/host` | `eino/adk/prebuilt/supervisor` |
| 路由机制 | HandOff (隐式) | Transfer Tool (显式) |
| 执行模型 | 直接调用 Invoke/Stream | Runner + 事件流 |
| 对话结束 | 自动 | Exit Tool (显式) |
| 中断恢复 | 不支持 | ✅ 支持 |
| Human-in-the-Loop | 需自行实现 | ✅ 内置支持 |
| 复杂度 | 简单 | 较复杂 |

## 架构设计

```
┌─────────────────────────────────────────────────────────────┐
│                    HTTP API / CLI                            │
└─────────────────────────┬───────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                     ADK Runner                               │
│              (事件流、中断恢复、状态管理)                       │
└─────────────────────────┬───────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              ai_tutor_supervisor                             │
│                  (学管协调员)                                 │
│                                                              │
│   指令: 分析问题 → 委派专家 → 汇总结果                         │
│   工具: Exit Tool (结束对话)                                  │
│   行为: Transfer to SubAgent                                 │
└─────────────────────────┬───────────────────────────────────┘
                          │
        ┌─────────────────┼─────────────────┐
        │                 │                 │
        ▼                 ▼                 ▼
┌───────────────┐ ┌───────────────┐ ┌───────────────┐
│ fault_handler │ │emotion_support│ │course_handler │
│  网络故障专家  │ │  情绪安抚专家  │ │  课程处理专家  │
│               │ │               │ │               │
│ Tools:        │ │ (纯对话)      │ │ Tools:        │
│ - 网络检测    │ │               │ │ - 约课        │
│ - 诊断        │ │               │ │ - 取消课      │
│ - 修复建议    │ │               │ │ - 查询课      │
│               │ │               │ │ - 调课        │
└───────────────┘ └───────────────┘ └───────────────┘
```

## 快速开始

### 1. 环境配置

```bash
# OpenAI 配置
export OPENAI_API_KEY="your-api-key"
export OPENAI_MODEL="gpt-4o"
export OPENAI_BASE_URL="https://api.openai.com/v1"  # 可选

# 或者使用 ARK 模型
export MODEL_TYPE="ark"
export ARK_API_KEY="your-ark-key"
export ARK_MODEL="your-ark-model"

# Langfuse 观测（可选）
export LANGFUSE_HOST="https://cloud.langfuse.com"
export LANGFUSE_PUBLIC_KEY="your-public-key"
export LANGFUSE_SECRET_KEY="your-secret-key"

# HTTP 端口（可选）
export HTTP_PORT="8080"
```

### 2. 运行服务

```bash
# HTTP 服务模式（默认）
go run ./adk/multiagent/ai-tutor

# 命令行交互模式
go run ./adk/multiagent/ai-tutor -mode cli

# 指定端口
go run ./adk/multiagent/ai-tutor -port 9090
```

### 3. 测试 API

```bash
# 约课场景
curl "http://localhost:8080/api/chat?uid=user001&message=我想约一节明天下午的一对一课程"

# 网络故障场景
curl "http://localhost:8080/api/chat?uid=user001&message=老师的视频一直卡顿"

# 情绪安抚场景
curl "http://localhost:8080/api/chat?uid=user001&message=英语太难了学不会"

# 查看事件流（响应中包含 events 字段）
curl -s "http://localhost:8080/api/chat?uid=user001&message=约课" | jq '.events'
```

## ADK 特性演示

### 1. 事件流

响应中的 `events` 字段展示了 Agent 的执行过程：

```json
{
  "status": "success",
  "message": "课程已预约成功...",
  "events": [
    {"agent_name": "ai_tutor_supervisor", "action": "transfer to course_handler"},
    {"agent_name": "course_handler", "action": "call tool: book_course"},
    {"agent_name": "course_handler", "content": "预约成功..."}
  ]
}
```

### 2. 扩展 Human-in-the-Loop

可以轻松添加审批流程（参考 `adk/human-in-the-loop/5_supervisor`）：

```go
// 将敏感工具包装为需要审批的工具
approvableTool := &tool2.InvokableApprovableTool{
    InvokableTool: cancelCourseTool,
}
```

### 3. 中断和恢复

ADK Runner 支持中断和恢复（参考 `adk/human-in-the-loop` 示例）：

```go
// 检查中断
info, ok := compose.ExtractInterruptInfo(err)
if ok {
    // 等待用户输入
    userInput := getUserInput()
    // 恢复执行
    runner.ResumeWithParams(ctx, userInput)
}
```

## 目录结构

```
adk/multiagent/ai-tutor/
├── main.go                 # 主程序入口
├── README.md               # 说明文档
├── agents/
│   ├── supervisor.go       # Supervisor Agent
│   ├── fault_handler.go    # 故障处理 Agent
│   ├── emotion_support.go  # 情绪安抚 Agent
│   └── course_handler.go   # 课程处理 Agent
├── tools/
│   ├── network_tools.go    # 网络检测工具
│   └── course_tools.go     # 课程管理工具
└── server/
    └── http.go             # HTTP 服务
```

## 与 flow/agent/ai_tutor 的代码对比

### Host-Specialist 模式

```go
// 创建 Host
hostAgent := &host.Host{
    ChatModel:    chatModel,
    SystemPrompt: "...",
}

// 创建 Specialist
specialist := &host.Specialist{
    AgentMeta: host.AgentMeta{
        Name:        "name",
        IntendedUse: "description",
    },
    Invokable:  func(...) { ... },
    Streamable: func(...) { ... },
}

// 组装 MultiAgent
multiAgent, _ := host.NewMultiAgent(ctx, &host.MultiAgentConfig{
    Host:        *hostAgent,
    Specialists: []*host.Specialist{specialist},
})

// 直接调用
result, _ := multiAgent.Generate(ctx, messages)
```

### ADK Supervisor 模式

```go
// 创建 Supervisor Agent
sv, _ := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
    Name:        "supervisor",
    Description: "...",
    Instruction: "...",
    Model:       model,
    Exit:        &adk.ExitTool{},  // 显式退出
})

// 创建 SubAgent
subAgent, _ := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
    Name:        "sub_agent",
    Description: "...",
    ToolsConfig: adk.ToolsConfig{...},
})

// 使用 Supervisor 模式组装
agent, _ := supervisor.New(ctx, &supervisor.Config{
    Supervisor: sv,
    SubAgents:  []adk.Agent{subAgent},
})

// 使用 Runner 执行
runner := adk.NewRunner(ctx, adk.RunnerConfig{
    Agent:           agent,
    EnableStreaming: true,
})
iter := runner.Query(ctx, query)

// 遍历事件流
for {
    event, hasEvent := iter.Next()
    if !hasEvent {
        break
    }
    // 处理事件...
}
```

## License

Apache-2.0
