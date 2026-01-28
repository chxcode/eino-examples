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

package tool

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// NetworkCheckServiceURL 网络检查服务 URL（可通过配置修改）
var NetworkCheckServiceURL = "http://localhost:9000/api/network"

// ==================== 网络质量检查工具 ====================

// NetworkQualityCheckTool 网络质量检查工具
type NetworkQualityCheckTool struct {
	client     *http.Client
	serviceURL string
}

// NewNetworkQualityCheckTool 创建网络质量检查工具
func NewNetworkQualityCheckTool(serviceURL string) *NetworkQualityCheckTool {
	if serviceURL == "" {
		serviceURL = NetworkCheckServiceURL
	}
	return &NetworkQualityCheckTool{
		client:     &http.Client{Timeout: 30 * time.Second},
		serviceURL: serviceURL,
	}
}

func (t *NetworkQualityCheckTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "check_network_quality",
		Desc: "检查用户的网络质量，包括延迟、丢包率、带宽等指标。当用户反馈视频卡顿、声音断断续续、画面模糊等问题时使用此工具。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"user_id": {
				Type:     "string",
				Desc:     "用户ID",
				Required: true,
			},
			"check_type": {
				Type:     "string",
				Desc:     "检查类型：quick（快速检查）、full（完整检查）、video（视频质量专项）",
				Required: false,
			},
		}),
	}, nil
}

func (t *NetworkQualityCheckTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var params struct {
		UserID    string `json:"user_id"`
		CheckType string `json:"check_type"`
	}
	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}

	if params.CheckType == "" {
		params.CheckType = "quick"
	}

	// TODO: 调用实际的网络检查服务
	// 这里预留调用外部服务的位置
	result, err := t.callNetworkCheckService(ctx, params.UserID, params.CheckType)
	if err != nil {
		// 如果外部服务不可用，返回模拟数据用于测试
		return t.getMockResult(params.UserID, params.CheckType), nil
	}

	return result, nil
}

func (t *NetworkQualityCheckTool) callNetworkCheckService(ctx context.Context, userID, checkType string) (string, error) {
	reqBody := map[string]string{
		"user_id":    userID,
		"check_type": checkType,
	}
	reqData, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.serviceURL+"/quality", bytes.NewReader(reqData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (t *NetworkQualityCheckTool) getMockResult(userID, checkType string) string {
	result := map[string]interface{}{
		"user_id":     userID,
		"check_type":  checkType,
		"status":      "completed",
		"latency_ms":  45,
		"packet_loss": 0.02,
		"bandwidth":   "50Mbps",
		"jitter_ms":   5,
		"quality":     "良好",
		"suggestion":  "网络状态良好，可以正常上课",
		"timestamp":   time.Now().Format(time.RFC3339),
	}
	data, _ := json.MarshalIndent(result, "", "  ")
	return string(data)
}

// ==================== 网络诊断工具 ====================

// NetworkDiagnoseTool 网络诊断工具
type NetworkDiagnoseTool struct {
	client     *http.Client
	serviceURL string
}

// NewNetworkDiagnoseTool 创建网络诊断工具
func NewNetworkDiagnoseTool(serviceURL string) *NetworkDiagnoseTool {
	if serviceURL == "" {
		serviceURL = NetworkCheckServiceURL
	}
	return &NetworkDiagnoseTool{
		client:     &http.Client{Timeout: 60 * time.Second},
		serviceURL: serviceURL,
	}
}

func (t *NetworkDiagnoseTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "diagnose_network",
		Desc: "对用户网络进行详细诊断，分析问题原因并提供解决方案。当网络质量检查发现问题时使用此工具。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"user_id": {
				Type:     "string",
				Desc:     "用户ID",
				Required: true,
			},
			"symptoms": {
				Type:     "string",
				Desc:     "用户描述的症状，如：视频卡顿、声音断续、无法连接等",
				Required: true,
			},
		}),
	}, nil
}

func (t *NetworkDiagnoseTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var params struct {
		UserID   string `json:"user_id"`
		Symptoms string `json:"symptoms"`
	}
	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}

	// TODO: 调用实际的网络诊断服务
	result, err := t.callDiagnoseService(ctx, params.UserID, params.Symptoms)
	if err != nil {
		return t.getMockDiagnosis(params.UserID, params.Symptoms), nil
	}

	return result, nil
}

func (t *NetworkDiagnoseTool) callDiagnoseService(ctx context.Context, userID, symptoms string) (string, error) {
	reqBody := map[string]string{
		"user_id":  userID,
		"symptoms": symptoms,
	}
	reqData, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.serviceURL+"/diagnose", bytes.NewReader(reqData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (t *NetworkDiagnoseTool) getMockDiagnosis(userID, symptoms string) string {
	result := map[string]interface{}{
		"user_id":  userID,
		"symptoms": symptoms,
		"diagnosis": map[string]interface{}{
			"possible_causes": []string{
				"WiFi信号不稳定",
				"网络拥堵",
				"路由器需要重启",
			},
			"severity": "中等",
		},
		"solutions": []map[string]string{
			{
				"step":        "1",
				"action":      "检查WiFi信号强度",
				"description": "请确保您的设备距离路由器不要太远，避免穿墙过多",
			},
			{
				"step":        "2",
				"action":      "重启路由器",
				"description": "关闭路由器电源，等待30秒后重新开启",
			},
			{
				"step":        "3",
				"action":      "关闭其他设备",
				"description": "暂时关闭其他占用网络的设备或应用",
			},
		},
		"estimated_fix_time": "5-10分钟",
		"timestamp":          time.Now().Format(time.RFC3339),
	}
	data, _ := json.MarshalIndent(result, "", "  ")
	return string(data)
}

// ==================== 网络修复建议工具 ====================

// NetworkFixSuggestionTool 网络修复建议工具
type NetworkFixSuggestionTool struct {
	client     *http.Client
	serviceURL string
}

// NewNetworkFixSuggestionTool 创建网络修复建议工具
func NewNetworkFixSuggestionTool(serviceURL string) *NetworkFixSuggestionTool {
	if serviceURL == "" {
		serviceURL = NetworkCheckServiceURL
	}
	return &NetworkFixSuggestionTool{
		client:     &http.Client{Timeout: 30 * time.Second},
		serviceURL: serviceURL,
	}
}

func (t *NetworkFixSuggestionTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "get_network_fix_suggestion",
		Desc: "根据网络问题类型获取具体的修复建议和操作步骤",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"problem_type": {
				Type:     "string",
				Desc:     "问题类型：latency（延迟高）、packet_loss（丢包）、bandwidth（带宽不足）、connection（连接问题）",
				Required: true,
			},
			"device_type": {
				Type:     "string",
				Desc:     "设备类型：pc、mac、ipad、phone",
				Required: false,
			},
		}),
	}, nil
}

func (t *NetworkFixSuggestionTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var params struct {
		ProblemType string `json:"problem_type"`
		DeviceType  string `json:"device_type"`
	}
	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}

	if params.DeviceType == "" {
		params.DeviceType = "pc"
	}

	// TODO: 调用实际的服务获取修复建议
	return t.getMockSuggestion(params.ProblemType, params.DeviceType), nil
}

func (t *NetworkFixSuggestionTool) getMockSuggestion(problemType, deviceType string) string {
	suggestions := map[string]interface{}{
		"latency": map[string]interface{}{
			"problem":     "网络延迟高",
			"description": "数据传输速度慢，导致视频和音频有延迟",
			"steps": []string{
				"1. 尽量使用有线网络连接",
				"2. 关闭正在下载的程序或视频",
				"3. 将设备靠近路由器",
				"4. 重启路由器和设备",
				"5. 如果使用VPN，请尝试关闭",
			},
		},
		"packet_loss": map[string]interface{}{
			"problem":     "网络丢包",
			"description": "数据包丢失导致画面卡顿或声音断断续续",
			"steps": []string{
				"1. 检查网线是否松动",
				"2. 重启路由器",
				"3. 检查是否有其他设备大量占用网络",
				"4. 联系网络运营商检查线路",
			},
		},
		"bandwidth": map[string]interface{}{
			"problem":     "带宽不足",
			"description": "网络带宽不够，无法支持高清视频通话",
			"steps": []string{
				"1. 关闭其他占用网络的应用",
				"2. 降低视频画质设置",
				"3. 避免在网络高峰期上课",
				"4. 考虑升级网络套餐",
			},
		},
		"connection": map[string]interface{}{
			"problem":     "连接问题",
			"description": "无法建立稳定的网络连接",
			"steps": []string{
				"1. 检查网络是否正常连接",
				"2. 重启设备和路由器",
				"3. 检查防火墙设置",
				"4. 尝试切换到移动数据网络",
				"5. 联系技术支持",
			},
		},
	}

	result, ok := suggestions[problemType]
	if !ok {
		result = map[string]interface{}{
			"problem":     "未知问题",
			"description": "请详细描述您遇到的问题",
			"steps": []string{
				"1. 重启设备和路由器",
				"2. 联系客服获取帮助",
			},
		}
	}

	resultMap := result.(map[string]interface{})
	resultMap["device_type"] = deviceType
	resultMap["timestamp"] = time.Now().Format(time.RFC3339)

	data, _ := json.MarshalIndent(resultMap, "", "  ")
	return string(data)
}

// GetNetworkTools 获取所有网络相关工具
func GetNetworkTools(serviceURL string) []tool.BaseTool {
	return []tool.BaseTool{
		NewNetworkQualityCheckTool(serviceURL),
		NewNetworkDiagnoseTool(serviceURL),
		NewNetworkFixSuggestionTool(serviceURL),
	}
}
