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

package retriever

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// DifyConfig Dify 检索器配置
type DifyConfig struct {
	BaseURL   string        // Dify API 地址
	APIKey    string        // API 密钥
	DatasetID string        // 数据集 ID
	TopK      int           // 返回结果数量
	Timeout   time.Duration // 请求超时时间
}

// DifyRetriever Dify 知识库检索器实现
type DifyRetriever struct {
	config *DifyConfig
	client *http.Client
	mu     sync.RWMutex
	closed bool
}

// NewDifyRetriever 创建 Dify 知识库检索器
func NewDifyRetriever(config *DifyConfig) (*DifyRetriever, error) {
	if config.BaseURL == "" {
		return nil, fmt.Errorf("dify base URL is required")
	}
	if config.APIKey == "" {
		return nil, fmt.Errorf("dify API key is required")
	}
	if config.TopK <= 0 {
		config.TopK = 5
	}
	if config.Timeout <= 0 {
		config.Timeout = 30 * time.Second
	}

	return &DifyRetriever{
		config: config,
		client: &http.Client{Timeout: config.Timeout},
	}, nil
}

// ============== Dify API 数据结构 ==============

// difyRetrievalRequest Dify 检索请求
type difyRetrievalRequest struct {
	Query          string                 `json:"query"`
	RetrievalModel map[string]interface{} `json:"retrieval_model"`
}

// difyQueryObject Dify 查询对象
type difyQueryObject struct {
	Content string `json:"content"`
}

// difySegment Dify 分段信息
type difySegment struct {
	ID            string   `json:"id"`
	Position      int      `json:"position"`
	DocumentID    string   `json:"document_id"`
	Content       string   `json:"content"`
	Answer        string   `json:"answer"`
	WordCount     int      `json:"word_count"`
	Tokens        int      `json:"tokens"`
	Keywords      []string `json:"keywords"`
	IndexNodeID   string   `json:"index_node_id"`
	IndexNodeHash string   `json:"index_node_hash"`
	HitCount      int      `json:"hit_count"`
	Enabled       bool     `json:"enabled"`
	DisabledAt    int64    `json:"disabled_at"`
	DisabledBy    string   `json:"disabled_by"`
	Status        string   `json:"status"`
	CreatedBy     string   `json:"created_by"`
	CreatedAt     int64    `json:"created_at"`
	IndexingAt    int64    `json:"indexing_at"`
	CompletedAt   int64    `json:"completed_at"`
	Error         string   `json:"error"`
	StoppedAt     int64    `json:"stopped_at"`
}

// difyRecord Dify 检索记录
type difyRecord struct {
	Segment      difySegment `json:"segment"`
	Score        float64     `json:"score"`
	TSNEPosition struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
	} `json:"tsne_position"`
}

// difyRetrievalResponse Dify 检索响应
type difyRetrievalResponse struct {
	Query   difyQueryObject `json:"query"`
	Records []difyRecord    `json:"records"`
}

// ============== KnowledgeRetriever 接口实现 ==============

// Name 返回检索器名称
func (r *DifyRetriever) Name() string {
	return "dify"
}

// Retrieve 检索知识库
func (r *DifyRetriever) Retrieve(ctx context.Context, query string, opts *RetrieveOptions) ([]*Document, error) {
	r.mu.RLock()
	if r.closed {
		r.mu.RUnlock()
		return nil, ErrRetrieverClosed
	}
	r.mu.RUnlock()

	if query == "" {
		return nil, ErrEmptyQuery
	}

	// 使用默认选项或合并选项
	if opts == nil {
		opts = DefaultRetrieveOptions()
	}
	topK := opts.TopK
	if topK <= 0 {
		topK = r.config.TopK
	}

	// 构建请求
	reqBody := &difyRetrievalRequest{
		Query: query,
		RetrievalModel: map[string]interface{}{
			"search_method":    "hybrid_search",
			"reranking_enable": true,
			"reranking_mode":   nil,
			"reranking_model": map[string]interface{}{
				"reranking_provider_name": "",
				"reranking_model_name":    "",
			},
			"weights":                 nil,
			"top_k":                   topK,
			"score_threshold_enabled": opts.ScoreThreshold > 0,
			"score_threshold":         opts.ScoreThreshold,
		},
	}

	reqData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 发送请求
	url := fmt.Sprintf("%s/datasets/%s/retrieve", r.config.BaseURL, r.config.DatasetID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+r.config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("dify API error: status=%d, body=%s", resp.StatusCode, string(body))
	}

	// 解析响应
	var difyResp difyRetrievalResponse
	if err := json.NewDecoder(resp.Body).Decode(&difyResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// 转换为 Document
	docs := make([]*Document, 0, len(difyResp.Records))
	for _, record := range difyResp.Records {
		doc := &Document{
			ID:      record.Segment.ID,
			Content: record.Segment.Content,
			Score:   record.Score,
			Metadata: map[string]interface{}{
				"document_id": record.Segment.DocumentID,
				"position":    record.Segment.Position,
				"keywords":    record.Segment.Keywords,
				"answer":      record.Segment.Answer,
				"word_count":  record.Segment.WordCount,
				"tokens":      record.Segment.Tokens,
			},
		}
		docs = append(docs, doc)
	}

	return docs, nil
}

// RetrieveFormatted 检索并格式化结果为字符串
func (r *DifyRetriever) RetrieveFormatted(ctx context.Context, query string, opts *RetrieveOptions) (string, error) {
	docs, err := r.Retrieve(ctx, query, opts)
	if err != nil {
		return "", err
	}

	return FormatDocuments(docs, "numbered"), nil
}

// Close 关闭检索器
func (r *DifyRetriever) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.closed = true
	r.client.CloseIdleConnections()
	return nil
}
