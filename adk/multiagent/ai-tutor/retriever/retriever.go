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

// Package retriever 提供知识库检索接口和实现
// 支持多种知识库后端（Dify、Elasticsearch、Milvus 等）
package retriever

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Document 检索到的文档
type Document struct {
	ID       string                 `json:"id"`
	Content  string                 `json:"content"`
	Score    float64                `json:"score"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// RetrieveOptions 检索选项
type RetrieveOptions struct {
	TopK           int     // 返回结果数量
	ScoreThreshold float64 // 分数阈值（0 表示不过滤）
	Filter         map[string]interface{} // 过滤条件
}

// DefaultRetrieveOptions 默认检索选项
func DefaultRetrieveOptions() *RetrieveOptions {
	return &RetrieveOptions{
		TopK:           5,
		ScoreThreshold: 0,
		Filter:         nil,
	}
}

// KnowledgeRetriever 知识库检索接口
// 实现此接口可以支持不同的知识库后端（Dify、Elasticsearch、Milvus 等）
type KnowledgeRetriever interface {
	// Retrieve 检索知识库
	// query: 查询文本
	// opts: 检索选项（可选）
	Retrieve(ctx context.Context, query string, opts *RetrieveOptions) ([]*Document, error)

	// RetrieveFormatted 检索并格式化结果为字符串
	// 适合直接作为上下文注入到 prompt 中
	RetrieveFormatted(ctx context.Context, query string, opts *RetrieveOptions) (string, error)

	// Name 返回检索器名称（用于日志和调试）
	Name() string

	// Close 关闭检索器（释放资源）
	Close() error
}

// RetrieverConfig 检索器配置
type RetrieverConfig struct {
	// 通用配置
	Type    RetrieverType `json:"type"`
	Timeout time.Duration `json:"timeout"`
	TopK    int           `json:"top_k"`

	// Dify 配置
	DifyBaseURL   string `json:"dify_base_url"`
	DifyAPIKey    string `json:"dify_api_key"`
	DifyDatasetID string `json:"dify_dataset_id"`

	// Elasticsearch 配置（预留）
	ESAddresses []string `json:"es_addresses"`
	ESIndex     string   `json:"es_index"`
	ESUsername  string   `json:"es_username"`
	ESPassword  string   `json:"es_password"`

	// Milvus 配置（预留）
	MilvusAddress    string `json:"milvus_address"`
	MilvusCollection string `json:"milvus_collection"`
}

// RetrieverType 检索器类型
type RetrieverType string

const (
	RetrieverTypeDify          RetrieverType = "dify"
	RetrieverTypeElasticsearch RetrieverType = "elasticsearch"
	RetrieverTypeMilvus        RetrieverType = "milvus"
	RetrieverTypeNone          RetrieverType = "none" // 不使用知识库
)

// 错误定义
var (
	ErrRetrieverNotConfigured   = errors.New("retriever not configured")
	ErrRetrieverTypeNotSupported = errors.New("retriever type not supported")
	ErrRetrieverClosed          = errors.New("retriever is closed")
	ErrEmptyQuery               = errors.New("query is empty")
)

// NewRetriever 根据类型创建检索器实例
func NewRetriever(config *RetrieverConfig) (KnowledgeRetriever, error) {
	if config == nil {
		return nil, ErrRetrieverNotConfigured
	}

	if config.Timeout <= 0 {
		config.Timeout = 30 * time.Second
	}
	if config.TopK <= 0 {
		config.TopK = 5
	}

	switch config.Type {
	case RetrieverTypeDify:
		return NewDifyRetriever(&DifyConfig{
			BaseURL:   config.DifyBaseURL,
			APIKey:    config.DifyAPIKey,
			DatasetID: config.DifyDatasetID,
			TopK:      config.TopK,
			Timeout:   config.Timeout,
		})
	case RetrieverTypeElasticsearch:
		// TODO: 实现 Elasticsearch 检索器
		return nil, ErrRetrieverTypeNotSupported
	case RetrieverTypeMilvus:
		// TODO: 实现 Milvus 检索器
		return nil, ErrRetrieverTypeNotSupported
	case RetrieverTypeNone, "":
		return NewNoopRetriever(), nil
	default:
		return nil, ErrRetrieverTypeNotSupported
	}
}

// FormatDocuments 将文档列表格式化为字符串
func FormatDocuments(docs []*Document, format string) string {
	if len(docs) == 0 {
		return ""
	}

	var result string
	for i, doc := range docs {
		switch format {
		case "numbered":
			result += fmt.Sprintf("【知识点 %d】\n%s\n\n", i+1, doc.Content)
		case "bullet":
			result += fmt.Sprintf("• %s\n\n", doc.Content)
		case "plain":
			result += doc.Content + "\n\n"
		default:
			result += fmt.Sprintf("【知识点 %d】\n%s\n\n", i+1, doc.Content)
		}
	}

	return result
}
