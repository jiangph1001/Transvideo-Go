package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"video-transcoder/models"
	"video-transcoder/service"

	"github.com/gorilla/mux"
)

type TaskHandler struct {
	taskService *service.TaskService
}

func NewTaskHandler(taskService *service.TaskService) *TaskHandler {
	return &TaskHandler{
		taskService: taskService,
	}
}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 参数验证
	if req.TaskType == "" {
		req.TaskType = "transcode"
	}

	switch req.TaskType {
	case "transcode":
		if req.InputFile == "" {
			http.Error(w, "input_file不能为空", http.StatusBadRequest)
			return
		}
	case "command":
		if req.Command == "" {
			http.Error(w, "command不能为空", http.StatusBadRequest)
			return
		}
	default:
		http.Error(w, "不支持的任务类型: "+req.TaskType, http.StatusBadRequest)
		return
	}

	// 创建任务
	task, err := h.taskService.CreateTask(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	task, err := h.taskService.GetTask(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	// 从查询参数获取分页信息
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	// 默认值
	page := 1
	limit := 20

	// 解析 page 参数
	if pageStr != "" {
		fmt.Sscanf(pageStr, "%d", &page)
		if page < 1 {
			page = 1
		}
	}

	// 解析 limit 参数
	if limitStr != "" {
		fmt.Sscanf(limitStr, "%d", &limit)
		if limit < 1 {
			limit = 20
		}
		if limit > 100 {
			limit = 100 // 最大限制100个
		}
	}

	// 获取分页任务（从数据库查询，保证排序）
	tasks, err := h.taskService.GetTasksWithPagination(page, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func (h *TaskHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats := h.taskService.GetStats()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (h *TaskHandler) CancelTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	err := h.taskService.CancelTask(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "任务已取消", "id": id})
}

func (h *TaskHandler) ServeStatic(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/index.html")
}

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	err := h.taskService.DeleteTask(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "任务已删除", "id": id})
}

func (h *TaskHandler) UpdateConcurrency(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MaxConcurrency int `json:"max_concurrency"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.MaxConcurrency <= 0 || req.MaxConcurrency > 1000 {
		http.Error(w, "并发数必须在1-1000之间", http.StatusBadRequest)
		return
	}

	if err := h.taskService.SetMaxConcurrency(req.MaxConcurrency); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":          "并发数已更新",
		"max_concurrency": req.MaxConcurrency,
	})
}
