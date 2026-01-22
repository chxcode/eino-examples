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

// GetCourseServiceURL 获取课程服务 URL
func GetCourseServiceURL() string {
	if url := os.Getenv("COURSE_SERVICE_URL"); url != "" {
		return url
	}
	return "http://localhost:9001/api/course"
}

// ==================== 约课工具 ====================

// BookCourseRequest 约课请求
type BookCourseRequest struct {
	UserID            string `json:"user_id" jsonschema_description:"用户ID"`
	CourseType        string `json:"course_type" jsonschema_description:"课程类型：one_on_one（一对一）、group（小班课）、trial（试听课）"`
	PreferredDate     string `json:"preferred_date" jsonschema_description:"期望日期，格式：2024-01-15"`
	PreferredTime     string `json:"preferred_time" jsonschema_description:"期望时间段，格式：14:00-15:00"`
	TeacherPreference string `json:"teacher_preference" jsonschema_description:"教师偏好：指定教师ID或留空表示无偏好"`
	CourseLevel       string `json:"course_level" jsonschema_description:"课程级别：beginner（初级）、intermediate（中级）、advanced（高级）"`
}

// BookCourseResponse 约课响应
type BookCourseResponse struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	BookingID string `json:"booking_id"`
	Details   struct {
		UserID        string `json:"user_id"`
		CourseType    string `json:"course_type"`
		Date          string `json:"date"`
		Time          string `json:"time"`
		Teacher       string `json:"teacher"`
		ClassroomLink string `json:"classroom_link"`
		Reminder      string `json:"reminder"`
	} `json:"details"`
	Timestamp string `json:"timestamp"`
}

// BookCourse 预约课程
func BookCourse(ctx context.Context, req *BookCourseRequest) (*BookCourseResponse, error) {
	// TODO: 调用实际的约课服务
	result, err := callCourseService(ctx, "/book", req)
	if err != nil {
		// 返回模拟数据
		return &BookCourseResponse{
			Status:    "success",
			Message:   "课程预约成功",
			BookingID: fmt.Sprintf("BK%d", time.Now().UnixNano()%100000),
			Details: struct {
				UserID        string `json:"user_id"`
				CourseType    string `json:"course_type"`
				Date          string `json:"date"`
				Time          string `json:"time"`
				Teacher       string `json:"teacher"`
				ClassroomLink string `json:"classroom_link"`
				Reminder      string `json:"reminder"`
			}{
				UserID:        req.UserID,
				CourseType:    req.CourseType,
				Date:          req.PreferredDate,
				Time:          req.PreferredTime,
				Teacher:       "Emily（资深外教）",
				ClassroomLink: "https://classroom.example.com/room/abc123",
				Reminder:      "课程开始前15分钟会发送短信提醒",
			},
			Timestamp: time.Now().Format(time.RFC3339),
		}, nil
	}

	var resp BookCourseResponse
	if err := json.Unmarshal(result, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func callCourseService(ctx context.Context, endpoint string, req interface{}) ([]byte, error) {
	reqData, _ := json.Marshal(req)
	client := &http.Client{Timeout: 30 * time.Second}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, GetCourseServiceURL()+endpoint, bytes.NewReader(reqData))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// NewBookCourseTool 创建约课工具
func NewBookCourseTool() (tool.InvokableTool, error) {
	return utils.InferTool(
		"book_course",
		"为用户预约课程。当用户想要约课、预约课程、安排上课时间时使用此工具。",
		BookCourse,
	)
}

// ==================== 取消课程工具 ====================

// CancelCourseRequest 取消课程请求
type CancelCourseRequest struct {
	UserID     string `json:"user_id" jsonschema_description:"用户ID"`
	BookingID  string `json:"booking_id" jsonschema_description:"预约ID，如果不知道可以留空"`
	CourseDate string `json:"course_date" jsonschema_description:"课程日期，格式：2024-01-15"`
	Reason     string `json:"reason" jsonschema_description:"取消原因"`
}

// CancelCourseResponse 取消课程响应
type CancelCourseResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Details struct {
		RefundPolicy string `json:"refund_policy"`
		RefundHours  int    `json:"refund_hours"`
		Suggestion   string `json:"suggestion"`
	} `json:"details"`
	Timestamp string `json:"timestamp"`
}

// CancelCourse 取消课程
func CancelCourse(ctx context.Context, req *CancelCourseRequest) (*CancelCourseResponse, error) {
	// TODO: 调用实际的取消课程服务
	return &CancelCourseResponse{
		Status:  "success",
		Message: "课程取消成功",
		Details: struct {
			RefundPolicy string `json:"refund_policy"`
			RefundHours  int    `json:"refund_hours"`
			Suggestion   string `json:"suggestion"`
		}{
			RefundPolicy: "课程开始前24小时取消，课时将返还至账户",
			RefundHours:  1,
			Suggestion:   "您可以随时重新预约课程",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}

// NewCancelCourseTool 创建取消课程工具
func NewCancelCourseTool() (tool.InvokableTool, error) {
	return utils.InferTool(
		"cancel_course",
		"取消用户已预约的课程。当用户想要取消课程、退课时使用此工具。",
		CancelCourse,
	)
}

// ==================== 查询课程工具 ====================

// QueryCourseRequest 查询课程请求
type QueryCourseRequest struct {
	UserID    string `json:"user_id" jsonschema_description:"用户ID"`
	QueryType string `json:"query_type" jsonschema_description:"查询类型：upcoming（即将开始的课程）、history（历史课程）、all（所有课程）"`
	DateRange string `json:"date_range" jsonschema_description:"日期范围，格式：2024-01-01,2024-01-31"`
}

// CourseInfo 课程信息
type CourseInfo struct {
	BookingID  string `json:"booking_id"`
	CourseType string `json:"course_type"`
	Date       string `json:"date"`
	Time       string `json:"time"`
	Teacher    string `json:"teacher"`
	Status     string `json:"status"`
	Topic      string `json:"topic"`
}

// QueryCourseResponse 查询课程响应
type QueryCourseResponse struct {
	Status         string       `json:"status"`
	Courses        []CourseInfo `json:"courses"`
	RemainingHours int          `json:"remaining_hours"`
	Timestamp      string       `json:"timestamp"`
}

// QueryCourse 查询课程
func QueryCourse(ctx context.Context, req *QueryCourseRequest) (*QueryCourseResponse, error) {
	if req.QueryType == "" {
		req.QueryType = "upcoming"
	}

	// TODO: 调用实际的查询服务
	return &QueryCourseResponse{
		Status: "success",
		Courses: []CourseInfo{
			{
				BookingID:  "BK12345",
				CourseType: "one_on_one",
				Date:       time.Now().AddDate(0, 0, 1).Format("2006-01-02"),
				Time:       "14:00-15:00",
				Teacher:    "Emily",
				Status:     "confirmed",
				Topic:      "口语练习 - 日常对话",
			},
			{
				BookingID:  "BK12346",
				CourseType: "group",
				Date:       time.Now().AddDate(0, 0, 3).Format("2006-01-02"),
				Time:       "19:00-20:00",
				Teacher:    "Michael",
				Status:     "confirmed",
				Topic:      "语法专项 - 时态",
			},
		},
		RemainingHours: 15,
		Timestamp:      time.Now().Format(time.RFC3339),
	}, nil
}

// NewQueryCourseTool 创建查询课程工具
func NewQueryCourseTool() (tool.InvokableTool, error) {
	return utils.InferTool(
		"query_course",
		"查询用户的课程信息，包括已预约的课程、课程历史等。",
		QueryCourse,
	)
}

// ==================== 调课工具 ====================

// RescheduleCourseRequest 调课请求
type RescheduleCourseRequest struct {
	UserID    string `json:"user_id" jsonschema_description:"用户ID"`
	BookingID string `json:"booking_id" jsonschema_description:"原预约ID"`
	NewDate   string `json:"new_date" jsonschema_description:"新日期，格式：2024-01-15"`
	NewTime   string `json:"new_time" jsonschema_description:"新时间段，格式：14:00-15:00"`
}

// RescheduleCourseResponse 调课响应
type RescheduleCourseResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Details struct {
		BookingID string `json:"booking_id"`
		NewDate   string `json:"new_date"`
		NewTime   string `json:"new_time"`
		Teacher   string `json:"teacher"`
		Reminder  string `json:"reminder"`
	} `json:"details"`
	Timestamp string `json:"timestamp"`
}

// RescheduleCourse 调整课程时间
func RescheduleCourse(ctx context.Context, req *RescheduleCourseRequest) (*RescheduleCourseResponse, error) {
	// TODO: 调用实际的调课服务
	return &RescheduleCourseResponse{
		Status:  "success",
		Message: "课程调整成功",
		Details: struct {
			BookingID string `json:"booking_id"`
			NewDate   string `json:"new_date"`
			NewTime   string `json:"new_time"`
			Teacher   string `json:"teacher"`
			Reminder  string `json:"reminder"`
		}{
			BookingID: req.BookingID,
			NewDate:   req.NewDate,
			NewTime:   req.NewTime,
			Teacher:   "原教师不变",
			Reminder:  "新的课程提醒已更新",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}

// NewRescheduleCourseTool 创建调课工具
func NewRescheduleCourseTool() (tool.InvokableTool, error) {
	return utils.InferTool(
		"reschedule_course",
		"调整已预约课程的时间。当用户想要改时间、调课时使用此工具。",
		RescheduleCourse,
	)
}

// GetAllCourseTools 获取所有课程相关工具
func GetAllCourseTools() ([]tool.BaseTool, error) {
	bookTool, err := NewBookCourseTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create book course tool: %w", err)
	}

	cancelTool, err := NewCancelCourseTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create cancel course tool: %w", err)
	}

	queryTool, err := NewQueryCourseTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create query course tool: %w", err)
	}

	rescheduleTool, err := NewRescheduleCourseTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create reschedule course tool: %w", err)
	}

	return []tool.BaseTool{bookTool, cancelTool, queryTool, rescheduleTool}, nil
}
