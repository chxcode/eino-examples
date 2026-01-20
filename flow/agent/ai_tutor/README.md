# AI 学管助手 (AI Tutor)

基于 Eino 框架实现的多 Agent 编排 AI 学管系统，专为在线英语教育场景设计。

## 功能特性

- 🤖 **多 Agent 协作**：采用集中式协调架构，支持多个专家 Agent 协同工作
- 🌐 **网络故障处理**：诊断和解决用户的网络问题（视频卡顿、音频断续等）
- 💬 **情绪安抚**：为学员提供心理支持和情绪疏导
- 📚 **课程管理**：支持约课、取消课、调课、查询课程等操作
- 📊 **Langfuse 集成**：完整的观测和追踪能力
- 💾 **会话管理**：支持多轮对话，保持上下文连贯
- 📖 **知识库集成**：支持 Dify 知识库检索

## 架构设计

```
┌─────────────────────────────────────────────────────────────┐
│                    HTTP API Layer                            │
│  /api/chat  /api/chat/stream  /api/history  /api/health     │
└─────────────────────────┬───────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                  Coordinator Agent                           │
│              (集中式协调员 - AI 学管)                         │
│                                                              │
│   ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│   │  Knowledge   │  │   Memory     │  │   Langfuse   │      │
│   │  Retriever   │  │   Manager    │  │   Tracing    │      │
│   │  (Dify)      │  │              │  │              │      │
│   └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────┬───────────────────────────────────┘
                          │
        ┌─────────────────┼─────────────────┐
        │                 │                 │
        ▼                 ▼                 ▼
┌───────────────┐ ┌───────────────┐ ┌───────────────┐
│ Fault Handler │ │   Emotion     │ │    Course     │
│    Agent      │ │   Support     │ │   Handler     │
│               │ │    Agent      │ │    Agent      │
│ ┌───────────┐ │ │               │ │ ┌───────────┐ │
│ │ Tools:    │ │ │  Pure LLM    │ │ │ Tools:    │ │
│ │ - 网络检测 │ │ │  对话能力     │ │ │ - 约课    │ │
│ │ - 诊断    │ │ │               │ │ │ - 取消课  │ │
│ │ - 修复建议│ │ │               │ │ │ - 查询课  │ │
│ └───────────┘ │ │               │ │ │ - 调课    │ │
└───────────────┘ └───────────────┘ │ └───────────┘ │
                                    └───────────────┘
```

## 快速开始

### 1. 环境配置

```bash
# 必需配置
export OPENAI_API_KEY="your-openai-api-key"

# 可选：OpenAI 配置
export OPENAI_BASE_URL="https://api.openai.com/v1"  # 默认值
export OPENAI_MODEL_NAME="gpt-4o"                    # 默认值

# 可选：HTTP 服务配置
export HTTP_PORT="8080"  # 默认值

# 可选：Langfuse 观测
export LANGFUSE_HOST="https://cloud.langfuse.com"
export LANGFUSE_PUBLIC_KEY="your-public-key"
export LANGFUSE_SECRET_KEY="your-secret-key"

# 可选：Dify 知识库
export DIFY_BASE_URL="https://api.dify.ai/v1"
export DIFY_API_KEY="your-dify-api-key"
export DIFY_DATASET_ID="your-dataset-id"

# 可选：外部服务（用于工具实际调用）
export NETWORK_CHECK_URL="http://localhost:9000/api/network"
export COURSE_SERVICE_URL="http://localhost:9001/api/course"
```

### 2. 安装依赖

```bash
cd flow/agent/ai_tutor
go mod tidy
```

### 3. 运行服务

```bash
go run main.go
```

### 4. 测试 API

```bash
# 约课场景
curl "http://localhost:8080/api/chat?uid=user001&message=我想约一节明天下午的一对一课程"

# 网络故障场景
curl "http://localhost:8080/api/chat?uid=user001&message=老师的视频一直卡顿，看不清画面"

# 情绪安抚场景
curl "http://localhost:8080/api/chat?uid=user001&message=我感觉英语太难了，学不会好沮丧"

# 查询课程
curl "http://localhost:8080/api/chat?uid=user001&message=帮我查一下我预约的课程"

# 流式聊天
curl "http://localhost:8080/api/chat/stream?uid=user001&message=你好"

# 查看历史记录
curl "http://localhost:8080/api/history?uid=user001"

# 删除历史记录
curl -X DELETE "http://localhost:8080/api/history?uid=user001"
```

## API 接口

### 聊天接口

**GET/POST `/api/chat`**

参数:
- `uid` (必需): 用户 ID
- `message` (必需): 消息内容

响应:
```json
{
  "status": "success",
  "message": "AI 回复内容",
  "user_id": "user001"
}
```

### 流式聊天接口

**GET `/api/chat/stream`**

参数:
- `uid` (必需): 用户 ID  
- `message` (必需): 消息内容

响应: Server-Sent Events (SSE) 流

### 历史记录接口

**GET `/api/history`**

参数:
- `uid` (可选): 用户 ID，不传则列出所有用户

**DELETE `/api/history`**

参数:
- `uid` (必需): 用户 ID

### 健康检查

**GET `/api/health`**

响应:
```json
{
  "status": "healthy",
  "service": "AI Tutor"
}
```

## Agent 职责说明

### 1. Coordinator Agent (协调员)
- 理解用户意图
- 将问题路由到合适的专家 Agent
- 整合知识库信息
- 提供通用性的服务

### 2. Fault Handler Agent (故障处理专家)
- 网络质量检查
- 问题诊断
- 提供修复建议

配置的工具:
- `check_network_quality`: 检查网络质量
- `diagnose_network`: 诊断网络问题
- `get_network_fix_suggestion`: 获取修复建议

### 3. Emotion Support Agent (情绪安抚专家)
- 学习焦虑疏导
- 考试压力缓解
- 服务投诉处理
- 鼓励与支持

### 4. Course Handler Agent (课程处理专家)
- 课程预约
- 课程取消
- 课程查询
- 课程调整

配置的工具:
- `book_course`: 预约课程
- `cancel_course`: 取消课程
- `query_course`: 查询课程
- `reschedule_course`: 调整课程时间

## 扩展开发

### 添加新的 Agent

1. 在 `agent/` 目录下创建新的 Agent 文件
2. 实现 `host.Specialist` 接口
3. 在 `coordinator.go` 中注册新的 Agent

### 添加新的工具

1. 在 `tool/` 目录下创建新的工具文件
2. 实现 `tool.InvokableTool` 接口
3. 在对应的 Agent 中注册工具

### 集成实际服务

工具代码中已预留了调用外部 HTTP 服务的位置，只需:
1. 配置相应的环境变量（如 `NETWORK_CHECK_URL`）
2. 确保外部服务实现了对应的 API 接口

## 注意事项

1. 当前工具返回的是模拟数据，实际使用时需要集成真实的后端服务
2. Dify 知识库需要提前配置并创建相应的数据集
3. 建议在生产环境中配置 Langfuse 进行监控和调试

## License

Apache-2.0
