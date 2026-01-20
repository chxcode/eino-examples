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

package retriever

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
)

// DifyRetrieverConfig Dify 知识库检索器配置
type DifyRetrieverConfig struct {
	BaseURL   string
	APIKey    string
	DatasetID string
	TopK      int
	Timeout   time.Duration
}

// DifyRetriever Dify 知识库检索器
type DifyRetriever struct {
	config *DifyRetrieverConfig
	client *http.Client
}

// NewDifyRetriever 创建 Dify 知识库检索器
func NewDifyRetriever(config *DifyRetrieverConfig) (*DifyRetriever, error) {
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

// difyRetrievalRequest Dify 检索请求
type difyRetrievalRequest struct {
	Query          string                 `json:"query"`
	RetrievalModel map[string]interface{} `json:"retrieval_model"`
}

// difyRetrievalResponse Dify 检索响应
type difyRetrievalResponse struct {
	Query   string `json:"query"`
	Records []struct {
		Segment struct {
			ID         string  `json:"id"`
			Position   int     `json:"position"`
			DocumentID string  `json:"document_id"`
			Content    string  `json:"content"`
			Answer     string  `json:"answer"`
			WordCount  int     `json:"word_count"`
			Tokens     int     `json:"tokens"`
			Keywords   []string `json:"keywords"`
			IndexNodeID string  `json:"index_node_id"`
			IndexNodeHash string `json:"index_node_hash"`
			HitCount    int     `json:"hit_count"`
			Enabled     bool    `json:"enabled"`
			Status      string  `json:"status"`
			CreatedBy   string  `json:"created_by"`
			CreatedAt   int64   `json:"created_at"`
			IndexedAt   int64   `json:"indexed_at"`
			CompletedAt int64   `json:"completed_at"`
			Error       string  `json:"error"`
			StoppedAt   int64   `json:"stopped_at"`
		} `json:"segment"`
		Score float64 `json:"score"`
		TSNE  struct {
			X float64 `json:"x"`
			Y float64 `json:"y"`
		} `json:"tsne_position"`
	} `json:"records"`
}

// Retrieve 检索知识库
func (r *DifyRetriever) Retrieve(ctx context.Context, query string, opts ...retriever.Option) ([]*schema.Document, error) {
	// 构建请求
	reqBody := &difyRetrievalRequest{
		Query: query,
		RetrievalModel: map[string]interface{}{
			"search_method":              "hybrid_search",
			"reranking_enable":           true,
			"reranking_mode":             nil,
			"reranking_model": map[string]interface{}{
				"reranking_provider_name": "",
				"reranking_model_name":    "",
			},
			"weights":       nil,
			"top_k":         r.config.TopK,
			"score_threshold_enabled": false,
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

	// 转换为 schema.Document
	docs := make([]*schema.Document, 0, len(difyResp.Records))
	for _, record := range difyResp.Records {
		doc := &schema.Document{
			ID:      record.Segment.ID,
			Content: record.Segment.Content,
			MetaData: map[string]interface{}{
				"document_id": record.Segment.DocumentID,
				"position":    record.Segment.Position,
				"keywords":    record.Segment.Keywords,
				"answer":      record.Segment.Answer,
			},
		}
		doc.WithScore(record.Score)
		docs = append(docs, doc)
	}

	return docs, nil
}

// RetrieveWithFormatted 检索并格式化结果为字符串
func (r *DifyRetriever) RetrieveWithFormatted(ctx context.Context, query string) (string, error) {
	docs, err := r.Retrieve(ctx, query)
	if err != nil {
		return "", err
	}

	if len(docs) == 0 {
		return "", nil
	}

	var result string
	for i, doc := range docs {
		result += fmt.Sprintf("【知识点 %d】\n%s\n\n", i+1, doc.Content)
	}

	return result, nil
}
