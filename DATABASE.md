# 数据库持久化说明

系统现在使用 SQLite 数据库保存所有任务，确保程序重启后任务不会丢失。

## 功能特性

- **自动保存**：所有任务自动持久化到 SQLite 数据库
- **重启恢复**：程序重启后自动恢复未完成任务
- **历史记录**：可以查询所有历史任务（包括已完成和失败的任务）
- **事务安全**：数据库操作保证数据一致性

## 数据库文件

**文件位置**：`./tasks.db`（相对于程序运行目录）

**操作数据库**：
```bash
# 查看数据库内容
sqlite3 tasks.db

# 查看所有任务
SELECT * FROM tasks ORDER BY created_at DESC;

# 查看失败的任务
SELECT * FROM tasks WHERE status = 'failed';

# 导出所有任务到CSV
sqlite3 -csv tasks.db "SELECT * FROM tasks" > tasks_backup.csv

# 备份数据库
cp tasks.db tasks_backup.db

# 清空数据库（会删除所有任务记录）
rm tasks.db
```

## 数据库结构

```sql
CREATE TABLE tasks (
    id TEXT PRIMARY KEY,           -- 任务ID（UUID）
    input_file TEXT NOT NULL,      -- 输入文件名
    output_file TEXT NOT NULL,     -- 输出文件名
    bitrate TEXT,                  -- 码率
    status TEXT NOT NULL,          -- 状态（pending/running/completed/failed）
    progress TEXT,                 -- 进度信息
    error TEXT,                    -- 错误信息
    created_at TIMESTAMP NOT NULL, -- 创建时间
    started_at TIMESTAMP,          -- 开始时间
    completed_at TIMESTAMP         -- 完成时间
);

CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_created_at ON tasks(created_at);
```

## 重启恢复机制

### 工作原理

1. **程序启动时**：
   - 自动创建数据库（如果不存在）
   - 加载 `pending` 状态的任务到队列
   - 加载 `running` 状态的任务并标记为 `failed`（因为程序中断）

2. **任务提交时**：
   - 立即保存到数据库
   - 同时加入内存队列和缓存

3. **任务状态更新时**：
   - 状态变化自动保存到数据库
   - 确保数据不丢失

### 恢复流程

```
程序启动
    ↓
创建/连接数据库
    ↓
查询 pending 任务 → 重新加入队列
    ↓
查询 running 任务 → 标记为 failed（并记录错误原因）
    ↓
启动工作池
    ↓
开始处理队列中的任务
```

## 使用示例

### 场景 1：程序意外崩溃后重启

**发生前**：
- 有 5 个 pending 任务
- 有 2 个 running 任务

**重启后**：
- 5 个 pending 任务自动重新加入队列
- 2 个 running 任务标记为 failed（错误原因："程序重启导致任务中断"）
- 队列继续处理 5 个 pending 任务

### 场景 2：正常重启
```bash
# Ctrl+C 停止程序
# 重新启动
./video-transcoder

# 输出：
# 正在关闭任务服务...
# 任务服务已关闭
# 加载配置:
#   ...
# 从数据库加载未完成任务...
# 任务 xxx 已重新加入队列
# 任务 yyy 已重新加入队列
# 成功加载 5 个未完成任务
# 视频转码管理系统启动成功!
```

### 场景 3：查看历史任务

```bash
# 通过 API 查询所有任务（包括已完成的）
curl http://localhost:8080/api/tasks

# 响应包含所有历史任务：
# - pending（等待中）
# - running（运行中）
# - completed（已完成）
# - failed（失败）
```

## 性能考虑

- **数据库**：SQLite（文件型数据库，无需独立服务）
- **并发**：支持多goroutine同时读写
- **缓存**：内存中保留任务map，减少数据库查询
- **写入**：每次状态变化都会立即写入，确保数据安全

## 数据保留

**默认行为**：
- ✅ 已完成的任务保留在数据库中
- ✅ 失败的任务保留在数据库中
- ✅ 所有任务的输入/输出路径、错误信息都保留

**清理数据**：
```bash
# 删除旧的已完成任务（需要手动执行SQL）
sqlite3 tasks.db "DELETE FROM tasks WHERE status = 'completed' AND created_at < datetime('now', '-30 days');"

# 删除所有任务记录
rm tasks.db
```

## 备份和恢复

### 备份数据库
```bash
# 方法1：复制文件
cp tasks.db tasks_$(date +%Y%m%d).db

# 方法2：导出SQL
sqlite3 tasks.db ".dump" > tasks_backup.sql

# 方法3：导出CSV
sqlite3 -csv tasks.db "SELECT * FROM tasks" > tasks_backup.csv
```

### 恢复数据库
```bash
# 方法1：复制文件
cp tasks_backup.db tasks.db

# 方法2：导入SQL
sqlite3 tasks.db < tasks_backup.sql
```

## API 行为变化

### 创建任务
- ✅ 立即保存到数据库
- ✅ 返回任务ID

### 查询任务
- ✅ `GET /api/tasks` - 返回所有任务（包括历史）
- ✅ `GET /api/tasks/{id}` - 返回单个任务

### 统计信息
- ✅ `GET /api/stats` - 统计所有任务状态

## 开发和调试

### 查看数据库
```bash
# 安装 SQLite 客户端（Ubuntu/Debian）
sudo apt-get install sqlite3

# 安装 SQLite 客户端（macOS）
brew install sqlite

# 查看数据库
sqlite3 tasks.db

# 常用命令
.tables           # 查看表
.schema tasks     # 查看表结构
SELECT * FROM tasks;  # 查询所有任务
SELECT COUNT(*) FROM tasks WHERE status = 'pending';  # 统计pending任务
```

### 数据库优化

如果任务数量很多（数万条），可以定期清理：

```bash
# 删除30天前的已完成任务
sqlite3 tasks.db "DELETE FROM tasks WHERE status = 'completed' AND created_at < date('now', '-30 days');"

# VACUUM 压缩数据库
sqlite3 tasks.db "VACUUM;"
```

## 故障排查

### 问题：无法写入数据库

**症状**：
```
unable to open database file
```

**原因**：
- 程序没有写入权限
- 磁盘已满

**解决**：
```bash
# 检查权限
ls -la tasks.db

# 修改权限
chmod 644 tasks.db

# 检查磁盘空间
df -h
```

### 问题：数据库损坏

**症状**：
```
database disk image is malformed
```

**解决**：
```bash
# 尝试修复（备份后）
sqlite3 tasks.db ".dump" > dump.sql
sqlite3 tasks_new.db < dump.sql
mv tasks_new.db tasks.db
```

### 问题：数据库被锁定

**症状**：
```
database is locked
```

**原因**：
- 多个进程同时写入
- 前一个进程没有正确关闭

**解决**：
```bash
# 等待自动解锁（有超时机制）
# 或重启程序
```

## 技术实现

### 使用的库

- **github.com/mattn/go-sqlite3** - SQLite 驱动

### 数据库操作

- **创建**：INSERT INTO tasks VALUES (...)
- **查询**：SELECT * FROM tasks WHERE id = ?
- **更新**：UPDATE tasks SET status = ?, ... WHERE id = ?
- **删除**：DELETE FROM tasks WHERE id = ?

### 事务支持

数据库操作默认自动提交，确保每个操作都是原子性的。

## 更新日志

- **2025-11-24**: 添加 SQLite 数据库持久化支持
- **2025-11-24**: 实现任务重启恢复功能
- **2025-11-24**: 添加数据库操作接口（Create, Read, Update, Delete）
