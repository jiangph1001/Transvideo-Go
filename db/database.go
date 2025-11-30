package db

import (
	"database/sql"
	"fmt"
	"time"
	"video-transcoder/models"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	db *sql.DB
}

// NewDatabase 创建新的数据库连接
func NewDatabase(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}

	database := &Database{db: db}

	// 创建表
	if err := database.createTables(); err != nil {
		return nil, fmt.Errorf("创建表失败: %w", err)
	}

	// 设置连接参数
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(time.Hour)

	return database, nil
}

// createTables 创建数据库表
func (d *Database) createTables() error {
	query := `
	CREATE TABLE IF NOT EXISTS tasks (
		id TEXT PRIMARY KEY,
		task_type TEXT DEFAULT 'transcode',
		input_file TEXT,
		output_file TEXT,
		bitrate TEXT,
		command TEXT,
		timeout INTEGER DEFAULT 300,
		status TEXT NOT NULL,
		progress TEXT,
		error TEXT,
		logs TEXT,
		created_at TIMESTAMP NOT NULL,
		started_at TIMESTAMP,
		completed_at TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
	CREATE INDEX IF NOT EXISTS idx_tasks_created_at ON tasks(created_at);
	`

	_, err := d.db.Exec(query)
	if err != nil {
		return err
	}

	// 对现有数据库进行迁移
	if err := d.migrateExistingDB(); err != nil {
		// 迁移不是必须的，即使失败也继续
		fmt.Printf("数据库迁移警告: %v\n", err)
	}

	return nil
}

// migrateExistingDB 为现有数据库添加缺失的列
func (d *Database) migrateExistingDB() error {
	// 检查并添加 task_type 列
	if err := d.addColumnIfNotExists("task_type", "TEXT DEFAULT 'transcode'"); err != nil {
		return fmt.Errorf("添加task_type列失败: %w", err)
	}

	// 检查并添加 command 列
	if err := d.addColumnIfNotExists("command", "TEXT"); err != nil {
		return fmt.Errorf("添加command列失败: %w", err)
	}

	// 检查并添加 timeout 列
	if err := d.addColumnIfNotExists("timeout", "INTEGER DEFAULT 300"); err != nil {
		return fmt.Errorf("添加timeout列失败: %w", err)
	}

	// 更新现有任务的 task_type（如果为空）
	updateQuery := `UPDATE tasks SET task_type = 'transcode' WHERE task_type IS NULL OR task_type = ''`
	if _, err := d.db.Exec(updateQuery); err != nil {
		return fmt.Errorf("更新现有任务task_type失败: %w", err)
	}

	return nil
}

// addColumnIfNotExists 检查列是否存在，如果不存在则添加
func (d *Database) addColumnIfNotExists(columnName, columnDef string) error {
	// 尝试查询该列，如果失败则说明列不存在
	query := fmt.Sprintf(`SELECT %s FROM tasks LIMIT 1`, columnName)
	_, err := d.db.Exec(query)
	if err != nil {
		// 列不存在，添加它
		alterQuery := fmt.Sprintf(`ALTER TABLE tasks ADD COLUMN %s %s`, columnName, columnDef)
		_, alterErr := d.db.Exec(alterQuery)
		if alterErr != nil {
			return fmt.Errorf("添加列 %s 失败: %w", columnName, alterErr)
		}
	}
	return nil
}

// CreateTask 创建任务
func (d *Database) CreateTask(task *models.Task) error {
	query := `
	INSERT INTO tasks (
		id, task_type, input_file, output_file, bitrate, command, timeout, status, progress,
		error, logs, created_at, started_at, completed_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := d.db.Exec(
		query,
		task.ID,
		task.TaskType,
		task.InputFile,
		task.OutputFile,
		task.Bitrate,
		task.Command,
		task.Timeout,
		task.Status,
		task.Progress,
		task.Error,
		task.Logs,
		task.CreatedAt,
		task.StartedAt,
		task.CompletedAt,
	)
	if err != nil {
		return fmt.Errorf("创建任务失败: %w", err)
	}

	return nil
}

// GetTask 获取单个任务
func (d *Database) GetTask(id string) (*models.Task, error) {
	query := `
	SELECT id, task_type, input_file, output_file, bitrate, command, timeout,
		status, progress, error, logs, created_at, started_at, completed_at
	FROM tasks
	WHERE id = ?
	`

	var task models.Task
	var startedAt, completedAt sql.NullTime

	err := d.db.QueryRow(query, id).Scan(
		&task.ID,
		&task.TaskType,
		&task.InputFile,
		&task.OutputFile,
		&task.Bitrate,
		&task.Command,
		&task.Timeout,
		&task.Status,
		&task.Progress,
		&task.Error,
		&task.Logs,
		&task.CreatedAt,
		&startedAt,
		&completedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("任务不存在")
	}
	if err != nil {
		return nil, fmt.Errorf("查询任务失败: %w", err)
	}

	if startedAt.Valid {
		task.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		task.CompletedAt = &completedAt.Time
	}

	return &task, nil
}

// GetAllTasks 获取所有任务
func (d *Database) GetAllTasks() ([]*models.Task, error) {
	query := `
	SELECT id, task_type, input_file, output_file, bitrate, command, timeout,
		status, progress, error, logs, created_at, started_at, completed_at
	FROM tasks
	ORDER BY created_at DESC
	`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("查询任务失败: %w", err)
	}
	defer rows.Close()

	var tasks []*models.Task

	for rows.Next() {
		var task models.Task
		var startedAt, completedAt sql.NullTime

		err := rows.Scan(
			&task.ID,
			&task.TaskType,
			&task.InputFile,
			&task.OutputFile,
			&task.Bitrate,
			&task.Command,
			&task.Timeout,
			&task.Status,
			&task.Progress,
			&task.Error,
			&task.Logs,
			&task.CreatedAt,
			&startedAt,
			&completedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描任务失败: %w", err)
		}

		if startedAt.Valid {
			task.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			task.CompletedAt = &completedAt.Time
		}

		tasks = append(tasks, &task)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历任务失败: %w", err)
	}

	return tasks, nil
}

// GetTasksWithPagination 分页获取任务
func (d *Database) GetTasksWithPagination(page, limit int) ([]*models.Task, error) {
	offset := (page - 1) * limit

	query := `
	SELECT id, task_type, input_file, output_file, bitrate, command, timeout,
		status, progress, error, logs, created_at, started_at, completed_at
	FROM tasks
	ORDER BY created_at DESC
	LIMIT ? OFFSET ?
	`

	rows, err := d.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("查询分页任务失败: %w", err)
	}
	defer rows.Close()

	var tasks []*models.Task

	for rows.Next() {
		var task models.Task
		var startedAt, completedAt sql.NullTime

		err := rows.Scan(
			&task.ID,
			&task.TaskType,
			&task.InputFile,
			&task.OutputFile,
			&task.Bitrate,
			&task.Command,
			&task.Timeout,
			&task.Status,
			&task.Progress,
			&task.Error,
			&task.Logs,
			&task.CreatedAt,
			&startedAt,
			&completedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描任务失败: %w", err)
		}

		if startedAt.Valid {
			task.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			task.CompletedAt = &completedAt.Time
		}

		tasks = append(tasks, &task)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历任务失败: %w", err)
	}

	return tasks, nil
}

// GetTasksByStatus 根据状态获取任务
func (d *Database) GetTasksByStatus(status models.TaskStatus) ([]*models.Task, error) {
	query := `
	SELECT id, task_type, input_file, output_file, bitrate, command, timeout,
		status, progress, error, logs, created_at, started_at, completed_at
	FROM tasks
	WHERE status = ?
	ORDER BY created_at DESC
	`

	rows, err := d.db.Query(query, status)
	if err != nil {
		return nil, fmt.Errorf("查询任务失败: %w", err)
	}
	defer rows.Close()

	var tasks []*models.Task

	for rows.Next() {
		var task models.Task
		var startedAt, completedAt sql.NullTime

		err := rows.Scan(
			&task.ID,
			&task.TaskType,
			&task.InputFile,
			&task.OutputFile,
			&task.Bitrate,
			&task.Command,
			&task.Timeout,
			&task.Status,
			&task.Progress,
			&task.Error,
			&task.Logs,
			&task.CreatedAt,
			&startedAt,
			&completedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描任务失败: %w", err)
		}

		if startedAt.Valid {
			task.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			task.CompletedAt = &completedAt.Time
		}

		tasks = append(tasks, &task)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历任务失败: %w", err)
	}

	return tasks, nil
}

// UpdateTask 更新任务状态
func (d *Database) UpdateTask(task *models.Task) error {
	query := `
	UPDATE tasks SET
		status = ?,
		progress = ?,
		error = ?,
		logs = ?,
		started_at = ?,
		completed_at = ?
	WHERE id = ?
	`

	_, err := d.db.Exec(
		query,
		task.Status,
		task.Progress,
		task.Error,
		task.Logs,
		task.StartedAt,
		task.CompletedAt,
		task.ID,
	)
	if err != nil {
		return fmt.Errorf("更新任务失败: %w", err)
	}

	return nil
}

// DeleteTask 删除任务
func (d *Database) DeleteTask(id string) error {
	query := `DELETE FROM tasks WHERE id = ?`

	_, err := d.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("删除任务失败: %w", err)
	}

	return nil
}

// GetStats 获取统计信息
func (d *Database) GetStats() (map[string]interface{}, error) {
	query := `
	SELECT
		COUNT(*) as total,
		SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending,
		SUM(CASE WHEN status = 'running' THEN 1 ELSE 0 END) as running,
		SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) as completed,
		SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed
	FROM tasks
	`

	var total, pending, running, completed, failed int
	err := d.db.QueryRow(query).Scan(&total, &pending, &running, &completed, &failed)
	if err != nil {
		return nil, fmt.Errorf("获取统计失败: %w", err)
	}

	return map[string]interface{}{
		"total_tasks": total,
		"status_count": map[models.TaskStatus]int{
			models.StatusPending:   pending,
			models.StatusRunning:   running,
			models.StatusCompleted: completed,
			models.StatusFailed:    failed,
		},
	}, nil
}

// Close 关闭数据库连接
func (d *Database) Close() error {
	return d.db.Close()
}
