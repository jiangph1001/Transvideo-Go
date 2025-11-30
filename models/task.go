package models

import (
	"time"
)

type TaskStatus string

const (
	StatusPending   TaskStatus = "pending"
	StatusRunning   TaskStatus = "running"
	StatusCompleted TaskStatus = "completed"
	StatusFailed    TaskStatus = "failed"
)

type TaskType string

const (
	TaskTypeTranscode TaskType = "transcode"
	TaskTypeCommand   TaskType = "command"
)

type Task struct {
	ID          string     `json:"id"`
	TaskType    TaskType   `json:"task_type"`
	InputFile   string     `json:"input_file"`
	OutputFile  string     `json:"output_file"`
	Bitrate     string     `json:"bitrate,omitempty"` // e.g. "1M", "500k"
	Command     string     `json:"command,omitempty"`
	Timeout     int        `json:"timeout,omitempty"` // 超时时间（秒）
	Status      TaskStatus `json:"status"`
	Progress    string     `json:"progress"`
	Error       string     `json:"error,omitempty"`
	Logs        string     `json:"logs,omitempty"` // 存储日志内容
	CreatedAt   time.Time  `json:"created_at"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type CreateTaskRequest struct {
	TaskType   string `json:"task_type,omitempty"` // "transcode" or "command"
	InputFile  string `json:"input_file,omitempty"`
	Bitrate    string `json:"bitrate,omitempty"`
	Command    string `json:"command,omitempty"`
	Timeout    int    `json:"timeout,omitempty"` // 超时时间（秒），默认300
}

type Config struct {
	MaxConcurrency int `json:"max_concurrency"`
}
