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

import "context"

// NoopRetriever 空操作检索器
// 当不需要知识库检索时使用，返回空结果
type NoopRetriever struct{}

// NewNoopRetriever 创建空操作检索器
func NewNoopRetriever() *NoopRetriever {
	return &NoopRetriever{}
}

// Name 返回检索器名称
func (r *NoopRetriever) Name() string {
	return "noop"
}

// Retrieve 检索（返回空结果）
func (r *NoopRetriever) Retrieve(ctx context.Context, query string, opts *RetrieveOptions) ([]*Document, error) {
	return []*Document{}, nil
}

// RetrieveFormatted 检索并格式化（返回空字符串）
func (r *NoopRetriever) RetrieveFormatted(ctx context.Context, query string, opts *RetrieveOptions) (string, error) {
	return "", nil
}

// Close 关闭检索器（无操作）
func (r *NoopRetriever) Close() error {
	return nil
}
