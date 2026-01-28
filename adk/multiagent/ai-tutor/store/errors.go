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

package store

import "errors"

var (
	// ErrConversationNotFound 会话不存在
	ErrConversationNotFound = errors.New("conversation not found")

	// ErrUnknownStoreType 未知的存储类型
	ErrUnknownStoreType = errors.New("unknown store type")

	// ErrStoreTypeNotImplemented 存储类型未实现
	ErrStoreTypeNotImplemented = errors.New("store type not implemented")

	// ErrStoreClosed 存储已关闭
	ErrStoreClosed = errors.New("store is closed")
)
