package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"video-transcoder/config"
	"video-transcoder/db"
	"video-transcoder/models"

	"github.com/google/uuid"
)

type TaskService struct {
	tasks          map[string]*models.Task  // 内存缓存
	queue          chan *models.Task
	maxConcurrency int
	executor       *DockerExecutor
	db             *db.Database
	wg             sync.WaitGroup
	mu             sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
}

func NewTaskService(cfg *config.Config) (*TaskService, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// 初始化数据库
	database, err := db.NewDatabase("./tasks.db")
	if err != nil {
		return nil, fmt.Errorf("初始化数据库失败: %w", err)
	}

	ts := &TaskService{
		tasks:          make(map[string]*models.Task),
		queue:          make(chan *models.Task, 1000), // 队列缓冲区大小
		maxConcurrency: cfg.MaxConcurrency,
		executor:       NewDockerExecutor(cfg),
		db:             database,
		ctx:            ctx,
		cancel:         cancel,
	}

	// 恢复未完成的任务
	if err := ts.loadTasksFromDB(); err != nil {
		return nil, fmt.Errorf("恢复任务失败: %w", err)
	}

	// 启动工作池
	for i := 0; i < cfg.MaxConcurrency; i++ {
		ts.wg.Add(1)
		go ts.worker(i)
	}

	return ts, nil
}

func (ts *TaskService) worker(id int) {
	defer ts.wg.Done()

	for {
		select {
		case <-ts.ctx.Done():
			fmt.Printf("Worker %d 退出\n", id)
			return
		case task := <-ts.queue:
			fmt.Printf("Worker %d 取出任务: %s (类型: %s)\n", id, task.ID, task.TaskType)

			// 检查任务状态，如果已经被取消则跳过执行
			ts.mu.RLock()
			currentTask, exists := ts.tasks[task.ID]
			ts.mu.RUnlock()

			if !exists {
				fmt.Printf("Worker %d 跳过任务 %s: 任务不存在\n", id, task.ID)
				continue
			}

			// 如果任务已经被取消（状态为failed且错误信息表明已被取消）
			if currentTask.Status == models.StatusFailed && currentTask.Error == "任务被取消" {
				fmt.Printf("Worker %d 跳过任务 %s: 任务已被取消\n", id, task.ID)
				continue
			}

			// 如果任务不是pending状态（可能是running或已经completed/failed），也跳过
			if currentTask.Status != models.StatusPending {
				fmt.Printf("Worker %d 跳过任务 %s: 任务状态为 %s\n", id, task.ID, currentTask.Status)
				continue
			}

			fmt.Printf("Worker %d 开始执行任务: %s\n", id, task.ID)

			// 更新任务状态为running
			ts.mu.Lock()
			task.Status = models.StatusRunning
			now := time.Now()
			task.StartedAt = &now
			ts.mu.Unlock()

			// 保存到数据库
			if dbErr := ts.db.UpdateTask(task); dbErr != nil {
				fmt.Printf("Worker %d 更新任务 %s 状态失败: %v\n", id, task.ID, dbErr)
			}

			// 根据任务类型执行不同的操作
			var err error
			switch task.TaskType {
			case models.TaskTypeTranscode:
				err = ts.executor.RunTranscodeTask(ts.ctx, task)
			case models.TaskTypeCommand:
				err = ts.executor.RunCommandTask(ts.ctx, task)
			default:
				err = fmt.Errorf("未知的任务类型: %s", task.TaskType)
			}

			if err != nil {
				fmt.Printf("Worker %d 执行任务 %s 失败: %v\n", id, task.ID, err)
			}

			// 任务完成后保存到数据库（包括状态、错误、日志等）
			if dbErr := ts.db.UpdateTask(task); dbErr != nil {
				fmt.Printf("Worker %d 保存任务 %s 到数据库失败: %v\n", id, task.ID, dbErr)
			}
		}
	}
}

func (ts *TaskService) CreateTask(req models.CreateTaskRequest) (*models.Task, error) {
	// 生成任务ID
	taskID := uuid.New().String()

	// 确定任务类型，默认为transcode
	taskType := models.TaskType(req.TaskType)
	if taskType == "" {
		taskType = models.TaskTypeTranscode
	}

	// 验证并创建任务
	task := &models.Task{
		ID:        taskID,
		TaskType:  taskType,
		Status:    models.StatusPending,
		CreatedAt: time.Now(),
	}

	switch taskType {
	case models.TaskTypeTranscode:
		// 验证输入文件路径（不包含路径分隔符，必须是相对文件名）
		if strings.Contains(req.InputFile, "/") || strings.Contains(req.InputFile, "..") {
			return nil, fmt.Errorf("输入文件不能包含路径，请只提供文件名（例如: a.mp4）")
		}

		// 生成输出文件名（添加_av1后缀）
		ext := filepath.Ext(req.InputFile)
		nameWithoutExt := strings.TrimSuffix(req.InputFile, ext)
		outputFile := fmt.Sprintf("%s_av1%s", nameWithoutExt, ext)

		// 检查输入文件是否存在（在实际挂载的输入目录中）
		fullInputPath := filepath.Join(ts.executor.cfg.InputDir, req.InputFile)
		if _, err := os.Stat(fullInputPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("输入文件不存在: %s", fullInputPath)
		}

		task.InputFile = req.InputFile
		task.OutputFile = outputFile
		task.Bitrate = req.Bitrate

	case models.TaskTypeCommand:
		// 验证命令不为空
		if req.Command == "" {
			return nil, fmt.Errorf("命令不能为空")
		}

		task.Command = req.Command
		task.Timeout = req.Timeout
		// 命令任务不需要input_file和output_file
		task.InputFile = ""
		task.OutputFile = ""

	default:
		return nil, fmt.Errorf("不支持的任务类型: %s", taskType)
	}

	// 保存到数据库
	if err := ts.db.CreateTask(task); err != nil {
		return nil, fmt.Errorf("保存任务到数据库失败: %w", err)
	}

	// 保存到内存
	ts.mu.Lock()
	ts.tasks[taskID] = task
	ts.mu.Unlock()

	// 添加到队列
	select {
	case ts.queue <- task:
		fmt.Printf("任务 %s 已加入队列 (类型: %s)\n", taskID, task.TaskType)
	default:
		// 如果队列已满，从数据库中删除任务
		if delErr := ts.db.DeleteTask(taskID); delErr != nil {
			fmt.Printf("删除任务失败: %v\n", delErr)
		}
		return nil, fmt.Errorf("队列已满，请稍后再试")
	}

	return task, nil
}

func (ts *TaskService) GetTask(id string) (*models.Task, error) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	task, exists := ts.tasks[id]
	if !exists {
		return nil, fmt.Errorf("任务不存在")
	}

	return task, nil
}

func (ts *TaskService) GetAllTasks() []*models.Task {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	tasks := make([]*models.Task, 0, len(ts.tasks))
	for _, task := range ts.tasks {
		tasks = append(tasks, task)
	}

	return tasks
}

// GetTasksWithPagination 从数据库获取分页任务（保证排序）
func (ts *TaskService) GetTasksWithPagination(page, limit int) ([]*models.Task, error) {
	// 从数据库查询，保证按创建时间降序排列
	tasks, err := ts.db.GetTasksWithPagination(page, limit)
	if err != nil {
		return nil, fmt.Errorf("获取分页任务失败: %w", err)
	}

	// 同时更新内存缓存（保持数据一致性）
	ts.mu.Lock()
	for _, task := range tasks {
		ts.tasks[task.ID] = task
	}
	ts.mu.Unlock()

	return tasks, nil
}

func (ts *TaskService) GetStats() map[string]interface{} {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	stats := map[string]interface{}{
		"total_tasks":     len(ts.tasks),
		"max_concurrency": ts.maxConcurrency,
	}

	// 统计各状态任务数
	statusCount := map[models.TaskStatus]int{}
	for _, task := range ts.tasks {
		statusCount[task.Status]++
	}
	stats["status_count"] = statusCount

	return stats
}

func (ts *TaskService) CancelTask(id string) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	task, exists := ts.tasks[id]
	if !exists {
		return fmt.Errorf("任务不存在")
	}

	// 只允许取消 pending 和 running 状态的任务
	if task.Status != models.StatusPending && task.Status != models.StatusRunning {
		return fmt.Errorf("任务已完成或失败，无法取消")
	}

	// 如果是运行中的任务，需要停止 Docker 容器
	if task.Status == models.StatusRunning {
		// 停止容器
		if err := ts.executor.StopContainer(fmt.Sprintf("transcode-%s", id)); err != nil {
			fmt.Printf("停止容器失败: %v\n", err)
			// 继续处理，即使停止失败也要标记任务为取消状态
		}

		// 更新任务状态
		task.Status = models.StatusFailed
		task.Error = "任务被取消"
		if err := ts.db.UpdateTask(task); err != nil {
			fmt.Printf("更新任务状态失败: %v\n", err)
		}

		return nil
	}

	// 如果是 pending 状态，从队列中移除
	if task.Status == models.StatusPending {
		// 需要从 channel 中移除任务（这是比较复杂的操作）
		// 简单方案：标记为失败，worker 在执行时会检查状态
		task.Status = models.StatusFailed
		task.Error = "任务被取消"
		if err := ts.db.UpdateTask(task); err != nil {
			fmt.Printf("更新任务状态失败: %v\n", err)
		}

		return nil
	}

	return nil
}

func (ts *TaskService) Shutdown() {
	fmt.Println("正在关闭任务服务...")
	ts.cancel()
	ts.wg.Wait()
	close(ts.queue)

	// 关闭数据库连接
	if ts.db != nil {
		if err := ts.db.Close(); err != nil {
			fmt.Printf("关闭数据库失败: %v\n", err)
		}
	}

	fmt.Println("任务服务已关闭")
}

func (ts *TaskService) SetMaxConcurrency(n int) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if n <= 0 {
		return fmt.Errorf("并发数必须大于0")
	}

	if n > 1000 {
		return fmt.Errorf("并发数不能超过1000")
	}

	currentWorkers := ts.maxConcurrency
	ts.maxConcurrency = n

	fmt.Printf("并发数调整: %d -> %d\n", currentWorkers, n)

	// 如果增加worker数量，启动新的worker
	if n > currentWorkers {
		for i := currentWorkers; i < n; i++ {
			ts.wg.Add(1)
			go ts.worker(i)
			fmt.Printf("启动新 worker %d\n", i)
		}
	}
	// 如果减少worker数量，worker会自然退出（通过ctx.Done）
	// 但这里不需要特别处理，因为worker是通过channel读取任务的
	// 当任务队列没有更多任务时，多余的worker会自然等待

	return nil
}

func (ts *TaskService) DeleteTask(id string) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	task, exists := ts.tasks[id]
	if !exists {
		return fmt.Errorf("任务不存在")
	}

	// 不能删除运行中的任务
	if task.Status == models.StatusRunning {
		return fmt.Errorf("不能删除运行中的任务，请先取消任务")
	}

	// 从数据库删除
	if err := ts.db.DeleteTask(id); err != nil {
		return fmt.Errorf("从数据库删除任务失败: %w", err)
	}

	// 从内存删除
	delete(ts.tasks, id)

	fmt.Printf("任务 %s 已删除\n", id)
	return nil
}

// SaveTaskLogs 保存任务日志到数据库
func (ts *TaskService) SaveTaskLogs(task *models.Task) error {
	return ts.db.UpdateTask(task)
}

// loadTasksFromDB 从数据库加载所有任务
func (ts *TaskService) loadTasksFromDB() error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	fmt.Println("从数据库加载所有任务...")

	// 加载所有状态的任务到内存，以便在前端显示
	allTasks, err := ts.db.GetAllTasks()
	if err != nil {
		return fmt.Errorf("加载任务失败: %w", err)
	}

	fmt.Printf("从数据库加载了 %d 个任务\n", len(allTasks))

	// 将所有任务添加到内存map
	for _, task := range allTasks {
		ts.tasks[task.ID] = task
	}

	// 重新添加未完成的任务到队列（pending和running）
	pendingTasks, err := ts.db.GetTasksByStatus(models.StatusPending)
	if err != nil {
		return fmt.Errorf("加载pending任务失败: %w", err)
	}

	runningTasks, err := ts.db.GetTasksByStatus(models.StatusRunning)
	if err != nil {
		return fmt.Errorf("加载running任务失败: %w", err)
	}

	// 将未完成的任务添加到队列
	for _, task := range append(pendingTasks, runningTasks...) {
		if task.Status == models.StatusPending {
			go func(t *models.Task) {
				// 异步添加，避免阻塞
				select {
				case ts.queue <- t:
					fmt.Printf("任务 %s 已重新加入队列\n", t.ID)
				default:
					fmt.Printf("队列已满，任务 %s 无法加入队列\n", t.ID)
				}
			}(task)
		}

		// running状态的任务需要特别处理：标记为failed（因为程序重启了）
		if task.Status == models.StatusRunning {
			task.Status = models.StatusFailed
			task.Error = "程序重启导致任务中断"
			if err := ts.db.UpdateTask(task); err != nil {
				fmt.Printf("更新running任务状态失败: %v\n", err)
			}
		}
	}

	fmt.Printf("成功加载 %d 个任务到内存（%d 个pending, %d 个running）\n",
		len(allTasks), len(pendingTasks), len(runningTasks))

	return nil
}
