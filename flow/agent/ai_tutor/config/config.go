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

package config

import (
	"os"
)

// Config 系统配置
type Config struct {
	// OpenAI 配置
	OpenAIAPIKey  string
	OpenAIBaseURL string
	OpenAIModel   string

	// Langfuse 配置
	LangfuseHost      string
	LangfusePublicKey string
	LangfuseSecretKey string

	// Dify 知识库配置
	DifyBaseURL string
	DifyAPIKey  string
	DifyDataset string

	// HTTP 服务配置
	HTTPPort string

	// 外部服务配置（用于工具调用）
	NetworkCheckURL  string // 网络质量检查服务 URL
	CourseServiceURL string // 课程服务 URL
}

// LoadConfig 从环境变量加载配置
func LoadConfig() *Config {
	return &Config{
		// OpenAI 配置
		OpenAIAPIKey:  getEnvOrDefault("OPENAI_API_KEY", ""),
		OpenAIBaseURL: getEnvOrDefault("OPENAI_BASE_URL", "https://api.openai.com/v1"),
		OpenAIModel:   getEnvOrDefault("OPENAI_MODEL_NAME", "gpt-4o"),

		// Langfuse 配置
		LangfuseHost:      getEnvOrDefault("LANGFUSE_HOST", "https://cloud.langfuse.com"),
		LangfusePublicKey: getEnvOrDefault("LANGFUSE_PUBLIC_KEY", ""),
		LangfuseSecretKey: getEnvOrDefault("LANGFUSE_SECRET_KEY", ""),

		// Dify 知识库配置
		DifyBaseURL: getEnvOrDefault("DIFY_BASE_URL", "https://api.dify.ai/v1"),
		DifyAPIKey:  getEnvOrDefault("DIFY_API_KEY", ""),
		DifyDataset: getEnvOrDefault("DIFY_DATASET_ID", ""),

		// HTTP 服务配置
		HTTPPort: getEnvOrDefault("HTTP_PORT", "8080"),

		// 外部服务配置
		NetworkCheckURL:  getEnvOrDefault("NETWORK_CHECK_URL", "http://localhost:9000/api/network"),
		CourseServiceURL: getEnvOrDefault("COURSE_SERVICE_URL", "http://localhost:9001/api/course"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
