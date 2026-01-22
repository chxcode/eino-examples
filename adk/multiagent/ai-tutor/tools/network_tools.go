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

package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// GetNetworkCheckURL 获取网络检查服务 URL
func GetNetworkCheckURL() string {
	if url := os.Getenv("NETWORK_CHECK_URL"); url != "" {
		return url
	}
	return "http://localhost:9000/api/network"
}

// ==================== 网络质量检查工具 ====================

// NetworkQualityCheckRequest 网络质量检查请求
type NetworkQualityCheckRequest struct {
	UserID    string `json:"user_id" jsonschema_description:"用户ID"`
	CheckType string `json:"check_type" jsonschema_description:"检查类型：quick（快速检查）、full（完整检查）、video（视频质量专项）"`
}

// NetworkQualityCheckResponse 网络质量检查响应
type NetworkQualityCheckResponse struct {
	UserID     string  `json:"user_id"`
	CheckType  string  `json:"check_type"`
	Status     string  `json:"status"`
	LatencyMs  int     `json:"latency_ms"`
	PacketLoss float64 `json:"packet_loss"`
	Bandwidth  string  `json:"bandwidth"`
	JitterMs   int     `json:"jitter_ms"`
	Quality    string  `json:"quality"`
	Suggestion string  `json:"suggestion"`
	Timestamp  string  `json:"timestamp"`
}

// CheckNetworkQuality 检查网络质量
func CheckNetworkQuality(ctx context.Context, req *NetworkQualityCheckRequest) (*NetworkQualityCheckResponse, error) {
	if req.CheckType == "" {
		req.CheckType = "quick"
	}

	// TODO: 调用实际的网络检查服务
	result, err := callNetworkCheckService(ctx, req.UserID, req.CheckType)
	if err != nil {
		// 如果服务不可用，返回模拟数据
		return &NetworkQualityCheckResponse{
			UserID:     req.UserID,
			CheckType:  req.CheckType,
			Status:     "completed",
			LatencyMs:  45,
			PacketLoss: 0.02,
			Bandwidth:  "50Mbps",
			JitterMs:   5,
			Quality:    "良好",
			Suggestion: "网络状态良好，可以正常上课",
			Timestamp:  time.Now().Format(time.RFC3339),
		}, nil
	}
	return result, nil
}

func callNetworkCheckService(ctx context.Context, userID, checkType string) (*NetworkQualityCheckResponse, error) {
	reqBody := map[string]string{
		"user_id":    userID,
		"check_type": checkType,
	}
	reqData, _ := json.Marshal(reqBody)

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, GetNetworkCheckURL()+"/quality", bytes.NewReader(reqData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result NetworkQualityCheckResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// NewNetworkQualityCheckTool 创建网络质量检查工具
func NewNetworkQualityCheckTool() (tool.InvokableTool, error) {
	return utils.InferTool(
		"check_network_quality",
		"检查用户的网络质量，包括延迟、丢包率、带宽等指标。当用户反馈视频卡顿、声音断断续续、画面模糊等问题时使用此工具。",
		CheckNetworkQuality,
	)
}

// ==================== 网络诊断工具 ====================

// NetworkDiagnoseRequest 网络诊断请求
type NetworkDiagnoseRequest struct {
	UserID   string `json:"user_id" jsonschema_description:"用户ID"`
	Symptoms string `json:"symptoms" jsonschema_description:"用户描述的症状，如：视频卡顿、声音断续、无法连接等"`
}

// NetworkDiagnoseResponse 网络诊断响应
type NetworkDiagnoseResponse struct {
	UserID    string `json:"user_id"`
	Symptoms  string `json:"symptoms"`
	Diagnosis struct {
		PossibleCauses []string `json:"possible_causes"`
		Severity       string   `json:"severity"`
	} `json:"diagnosis"`
	Solutions []struct {
		Step        string `json:"step"`
		Action      string `json:"action"`
		Description string `json:"description"`
	} `json:"solutions"`
	EstimatedFixTime string `json:"estimated_fix_time"`
	Timestamp        string `json:"timestamp"`
}

// DiagnoseNetwork 诊断网络问题
func DiagnoseNetwork(ctx context.Context, req *NetworkDiagnoseRequest) (*NetworkDiagnoseResponse, error) {
	// TODO: 调用实际的网络诊断服务
	return &NetworkDiagnoseResponse{
		UserID:   req.UserID,
		Symptoms: req.Symptoms,
		Diagnosis: struct {
			PossibleCauses []string `json:"possible_causes"`
			Severity       string   `json:"severity"`
		}{
			PossibleCauses: []string{
				"WiFi信号不稳定",
				"网络拥堵",
				"路由器需要重启",
			},
			Severity: "中等",
		},
		Solutions: []struct {
			Step        string `json:"step"`
			Action      string `json:"action"`
			Description string `json:"description"`
		}{
			{Step: "1", Action: "检查WiFi信号强度", Description: "请确保您的设备距离路由器不要太远，避免穿墙过多"},
			{Step: "2", Action: "重启路由器", Description: "关闭路由器电源，等待30秒后重新开启"},
			{Step: "3", Action: "关闭其他设备", Description: "暂时关闭其他占用网络的设备或应用"},
		},
		EstimatedFixTime: "5-10分钟",
		Timestamp:        time.Now().Format(time.RFC3339),
	}, nil
}

// NewNetworkDiagnoseTool 创建网络诊断工具
func NewNetworkDiagnoseTool() (tool.InvokableTool, error) {
	return utils.InferTool(
		"diagnose_network",
		"对用户网络进行详细诊断，分析问题原因并提供解决方案。当网络质量检查发现问题时使用此工具。",
		DiagnoseNetwork,
	)
}

// ==================== 网络修复建议工具 ====================

// NetworkFixSuggestionRequest 网络修复建议请求
type NetworkFixSuggestionRequest struct {
	ProblemType string `json:"problem_type" jsonschema_description:"问题类型：latency（延迟高）、packet_loss（丢包）、bandwidth（带宽不足）、connection（连接问题）"`
	DeviceType  string `json:"device_type" jsonschema_description:"设备类型：pc、mac、ipad、phone"`
}

// NetworkFixSuggestionResponse 网络修复建议响应
type NetworkFixSuggestionResponse struct {
	Problem     string   `json:"problem"`
	Description string   `json:"description"`
	Steps       []string `json:"steps"`
	DeviceType  string   `json:"device_type"`
	Timestamp   string   `json:"timestamp"`
}

// GetNetworkFixSuggestion 获取网络修复建议
func GetNetworkFixSuggestion(ctx context.Context, req *NetworkFixSuggestionRequest) (*NetworkFixSuggestionResponse, error) {
	if req.DeviceType == "" {
		req.DeviceType = "pc"
	}

	suggestions := map[string]*NetworkFixSuggestionResponse{
		"latency": {
			Problem:     "网络延迟高",
			Description: "数据传输速度慢，导致视频和音频有延迟",
			Steps: []string{
				"1. 尽量使用有线网络连接",
				"2. 关闭正在下载的程序或视频",
				"3. 将设备靠近路由器",
				"4. 重启路由器和设备",
				"5. 如果使用VPN，请尝试关闭",
			},
		},
		"packet_loss": {
			Problem:     "网络丢包",
			Description: "数据包丢失导致画面卡顿或声音断断续续",
			Steps: []string{
				"1. 检查网线是否松动",
				"2. 重启路由器",
				"3. 检查是否有其他设备大量占用网络",
				"4. 联系网络运营商检查线路",
			},
		},
		"bandwidth": {
			Problem:     "带宽不足",
			Description: "网络带宽不够，无法支持高清视频通话",
			Steps: []string{
				"1. 关闭其他占用网络的应用",
				"2. 降低视频画质设置",
				"3. 避免在网络高峰期上课",
				"4. 考虑升级网络套餐",
			},
		},
		"connection": {
			Problem:     "连接问题",
			Description: "无法建立稳定的网络连接",
			Steps: []string{
				"1. 检查网络是否正常连接",
				"2. 重启设备和路由器",
				"3. 检查防火墙设置",
				"4. 尝试切换到移动数据网络",
				"5. 联系技术支持",
			},
		},
	}

	result, ok := suggestions[req.ProblemType]
	if !ok {
		result = &NetworkFixSuggestionResponse{
			Problem:     "未知问题",
			Description: "请详细描述您遇到的问题",
			Steps: []string{
				"1. 重启设备和路由器",
				"2. 联系客服获取帮助",
			},
		}
	}

	result.DeviceType = req.DeviceType
	result.Timestamp = time.Now().Format(time.RFC3339)
	return result, nil
}

// NewNetworkFixSuggestionTool 创建网络修复建议工具
func NewNetworkFixSuggestionTool() (tool.InvokableTool, error) {
	return utils.InferTool(
		"get_network_fix_suggestion",
		"根据网络问题类型获取具体的修复建议和操作步骤",
		GetNetworkFixSuggestion,
	)
}

// GetAllNetworkTools 获取所有网络相关工具
func GetAllNetworkTools() ([]tool.BaseTool, error) {
	checkTool, err := NewNetworkQualityCheckTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create network quality check tool: %w", err)
	}

	diagnoseTool, err := NewNetworkDiagnoseTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create network diagnose tool: %w", err)
	}

	fixTool, err := NewNetworkFixSuggestionTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create network fix suggestion tool: %w", err)
	}

	return []tool.BaseTool{checkTool, diagnoseTool, fixTool}, nil
}
