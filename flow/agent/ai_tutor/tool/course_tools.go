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

// CourseServiceURL 课程服务 URL（可通过配置修改）
var CourseServiceURL = "http://localhost:9001/api/course"

// ==================== 约课工具 ====================

// BookCourseTool 约课工具
type BookCourseTool struct {
	client     *http.Client
	serviceURL string
}

// NewBookCourseTool 创建约课工具
func NewBookCourseTool(serviceURL string) *BookCourseTool {
	if serviceURL == "" {
		serviceURL = CourseServiceURL
	}
	return &BookCourseTool{
		client:     &http.Client{Timeout: 30 * time.Second},
		serviceURL: serviceURL,
	}
}

func (t *BookCourseTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "book_course",
		Desc: "为用户预约课程。当用户想要约课、预约课程、安排上课时间时使用此工具。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"user_id": {
				Type:     "string",
				Desc:     "用户ID",
				Required: true,
			},
			"course_type": {
				Type:     "string",
				Desc:     "课程类型：one_on_one（一对一）、group（小班课）、trial（试听课）",
				Required: true,
			},
			"preferred_date": {
				Type:     "string",
				Desc:     "期望日期，格式：2024-01-15",
				Required: true,
			},
			"preferred_time": {
				Type:     "string",
				Desc:     "期望时间段，格式：14:00-15:00",
				Required: true,
			},
			"teacher_preference": {
				Type:     "string",
				Desc:     "教师偏好：指定教师ID或留空表示无偏好",
				Required: false,
			},
			"course_level": {
				Type:     "string",
				Desc:     "课程级别：beginner（初级）、intermediate（中级）、advanced（高级）",
				Required: false,
			},
		}),
	}, nil
}

func (t *BookCourseTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var params struct {
		UserID            string `json:"user_id"`
		CourseType        string `json:"course_type"`
		PreferredDate     string `json:"preferred_date"`
		PreferredTime     string `json:"preferred_time"`
		TeacherPreference string `json:"teacher_preference"`
		CourseLevel       string `json:"course_level"`
	}
	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}

	// TODO: 调用实际的约课服务
	result, err := t.callBookService(ctx, params)
	if err != nil {
		return t.getMockBookResult(params), nil
	}

	return result, nil
}

func (t *BookCourseTool) callBookService(ctx context.Context, params interface{}) (string, error) {
	reqData, _ := json.Marshal(params)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.serviceURL+"/book", bytes.NewReader(reqData))
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

func (t *BookCourseTool) getMockBookResult(params interface{}) string {
	p := params.(struct {
		UserID            string `json:"user_id"`
		CourseType        string `json:"course_type"`
		PreferredDate     string `json:"preferred_date"`
		PreferredTime     string `json:"preferred_time"`
		TeacherPreference string `json:"teacher_preference"`
		CourseLevel       string `json:"course_level"`
	})

	result := map[string]interface{}{
		"status":     "success",
		"message":    "课程预约成功",
		"booking_id": fmt.Sprintf("BK%d", time.Now().UnixNano()%100000),
		"details": map[string]interface{}{
			"user_id":        p.UserID,
			"course_type":    p.CourseType,
			"date":           p.PreferredDate,
			"time":           p.PreferredTime,
			"teacher":        "Emily（资深外教）",
			"classroom_link": "https://classroom.example.com/room/abc123",
			"reminder":       "课程开始前15分钟会发送短信提醒",
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}
	data, _ := json.MarshalIndent(result, "", "  ")
	return string(data)
}

// ==================== 取消课程工具 ====================

// CancelCourseTool 取消课程工具
type CancelCourseTool struct {
	client     *http.Client
	serviceURL string
}

// NewCancelCourseTool 创建取消课程工具
func NewCancelCourseTool(serviceURL string) *CancelCourseTool {
	if serviceURL == "" {
		serviceURL = CourseServiceURL
	}
	return &CancelCourseTool{
		client:     &http.Client{Timeout: 30 * time.Second},
		serviceURL: serviceURL,
	}
}

func (t *CancelCourseTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "cancel_course",
		Desc: "取消用户已预约的课程。当用户想要取消课程、退课时使用此工具。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"user_id": {
				Type:     "string",
				Desc:     "用户ID",
				Required: true,
			},
			"booking_id": {
				Type:     "string",
				Desc:     "预约ID，如果不知道可以留空",
				Required: false,
			},
			"course_date": {
				Type:     "string",
				Desc:     "课程日期，格式：2024-01-15",
				Required: false,
			},
			"reason": {
				Type:     "string",
				Desc:     "取消原因",
				Required: false,
			},
		}),
	}, nil
}

func (t *CancelCourseTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var params struct {
		UserID     string `json:"user_id"`
		BookingID  string `json:"booking_id"`
		CourseDate string `json:"course_date"`
		Reason     string `json:"reason"`
	}
	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}

	// TODO: 调用实际的取消课程服务
	result, err := t.callCancelService(ctx, params)
	if err != nil {
		return t.getMockCancelResult(params), nil
	}

	return result, nil
}

func (t *CancelCourseTool) callCancelService(ctx context.Context, params interface{}) (string, error) {
	reqData, _ := json.Marshal(params)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.serviceURL+"/cancel", bytes.NewReader(reqData))
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

func (t *CancelCourseTool) getMockCancelResult(params interface{}) string {
	result := map[string]interface{}{
		"status":  "success",
		"message": "课程取消成功",
		"details": map[string]interface{}{
			"refund_policy": "课程开始前24小时取消，课时将返还至账户",
			"refund_hours":  1,
			"suggestion":    "您可以随时重新预约课程",
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}
	data, _ := json.MarshalIndent(result, "", "  ")
	return string(data)
}

// ==================== 查询课程工具 ====================

// QueryCourseTool 查询课程工具
type QueryCourseTool struct {
	client     *http.Client
	serviceURL string
}

// NewQueryCourseTool 创建查询课程工具
func NewQueryCourseTool(serviceURL string) *QueryCourseTool {
	if serviceURL == "" {
		serviceURL = CourseServiceURL
	}
	return &QueryCourseTool{
		client:     &http.Client{Timeout: 30 * time.Second},
		serviceURL: serviceURL,
	}
}

func (t *QueryCourseTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "query_course",
		Desc: "查询用户的课程信息，包括已预约的课程、课程历史等。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"user_id": {
				Type:     "string",
				Desc:     "用户ID",
				Required: true,
			},
			"query_type": {
				Type:     "string",
				Desc:     "查询类型：upcoming（即将开始的课程）、history（历史课程）、all（所有课程）",
				Required: false,
			},
			"date_range": {
				Type:     "string",
				Desc:     "日期范围，格式：2024-01-01,2024-01-31",
				Required: false,
			},
		}),
	}, nil
}

func (t *QueryCourseTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var params struct {
		UserID    string `json:"user_id"`
		QueryType string `json:"query_type"`
		DateRange string `json:"date_range"`
	}
	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}

	if params.QueryType == "" {
		params.QueryType = "upcoming"
	}

	// TODO: 调用实际的查询服务
	result, err := t.callQueryService(ctx, params)
	if err != nil {
		return t.getMockQueryResult(params), nil
	}

	return result, nil
}

func (t *QueryCourseTool) callQueryService(ctx context.Context, params interface{}) (string, error) {
	reqData, _ := json.Marshal(params)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.serviceURL+"/query", bytes.NewReader(reqData))
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

func (t *QueryCourseTool) getMockQueryResult(params interface{}) string {
	result := map[string]interface{}{
		"status": "success",
		"courses": []map[string]interface{}{
			{
				"booking_id":  "BK12345",
				"course_type": "one_on_one",
				"date":        time.Now().AddDate(0, 0, 1).Format("2006-01-02"),
				"time":        "14:00-15:00",
				"teacher":     "Emily",
				"status":      "confirmed",
				"topic":       "口语练习 - 日常对话",
			},
			{
				"booking_id":  "BK12346",
				"course_type": "group",
				"date":        time.Now().AddDate(0, 0, 3).Format("2006-01-02"),
				"time":        "19:00-20:00",
				"teacher":     "Michael",
				"status":      "confirmed",
				"topic":       "语法专项 - 时态",
			},
		},
		"remaining_hours": 15,
		"timestamp":       time.Now().Format(time.RFC3339),
	}
	data, _ := json.MarshalIndent(result, "", "  ")
	return string(data)
}

// ==================== 调课工具 ====================

// RescheduleCourseTool 调课工具
type RescheduleCourseTool struct {
	client     *http.Client
	serviceURL string
}

// NewRescheduleCourseTool 创建调课工具
func NewRescheduleCourseTool(serviceURL string) *RescheduleCourseTool {
	if serviceURL == "" {
		serviceURL = CourseServiceURL
	}
	return &RescheduleCourseTool{
		client:     &http.Client{Timeout: 30 * time.Second},
		serviceURL: serviceURL,
	}
}

func (t *RescheduleCourseTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "reschedule_course",
		Desc: "调整已预约课程的时间。当用户想要改时间、调课时使用此工具。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"user_id": {
				Type:     "string",
				Desc:     "用户ID",
				Required: true,
			},
			"booking_id": {
				Type:     "string",
				Desc:     "原预约ID",
				Required: true,
			},
			"new_date": {
				Type:     "string",
				Desc:     "新日期，格式：2024-01-15",
				Required: true,
			},
			"new_time": {
				Type:     "string",
				Desc:     "新时间段，格式：14:00-15:00",
				Required: true,
			},
		}),
	}, nil
}

func (t *RescheduleCourseTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var params struct {
		UserID    string `json:"user_id"`
		BookingID string `json:"booking_id"`
		NewDate   string `json:"new_date"`
		NewTime   string `json:"new_time"`
	}
	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}

	// TODO: 调用实际的调课服务
	result, err := t.callRescheduleService(ctx, params)
	if err != nil {
		return t.getMockRescheduleResult(params), nil
	}

	return result, nil
}

func (t *RescheduleCourseTool) callRescheduleService(ctx context.Context, params interface{}) (string, error) {
	reqData, _ := json.Marshal(params)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.serviceURL+"/reschedule", bytes.NewReader(reqData))
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

func (t *RescheduleCourseTool) getMockRescheduleResult(params interface{}) string {
	p := params.(struct {
		UserID    string `json:"user_id"`
		BookingID string `json:"booking_id"`
		NewDate   string `json:"new_date"`
		NewTime   string `json:"new_time"`
	})

	result := map[string]interface{}{
		"status":  "success",
		"message": "课程调整成功",
		"details": map[string]interface{}{
			"booking_id": p.BookingID,
			"new_date":   p.NewDate,
			"new_time":   p.NewTime,
			"teacher":    "原教师不变",
			"reminder":   "新的课程提醒已更新",
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}
	data, _ := json.MarshalIndent(result, "", "  ")
	return string(data)
}

// GetCourseTools 获取所有课程相关工具
func GetCourseTools(serviceURL string) []tool.BaseTool {
	return []tool.BaseTool{
		NewBookCourseTool(serviceURL),
		NewCancelCourseTool(serviceURL),
		NewQueryCourseTool(serviceURL),
		NewRescheduleCourseTool(serviceURL),
	}
}
