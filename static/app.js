// API 基础地址
const API_BASE = '/api';

// 存储正在运行任务的定时器
const runningTaskTimers = {};

// 存储选中的任务ID，用于批量删除
const selectedTasks = new Set();

// 存储当前显示的任务列表
let currentTasks = [];

// 格式化运行时间（从指定时间到现在的时长）
function formatRunningTime(startTime) {
    if (!startTime) return '-';
    const start = new Date(startTime).getTime();
    const now = Date.now();
    const diff = Math.floor((now - start) / 1000); // 秒

    if (diff < 60) {
        return `${diff}秒`;
    } else if (diff < 3600) {
        const minutes = Math.floor(diff / 60);
        const seconds = diff % 60;
        return `${minutes}分${seconds}秒`;
    } else {
        const hours = Math.floor(diff / 3600);
        const minutes = Math.floor((diff % 3600) / 60);
        return `${hours}小时${minutes}分钟`;
    }
}

// 更新指定任务的运行时间显示
function updateTaskRunningTime(taskId, startedAt) {
    const tbody = document.getElementById('task-list');
    const rows = tbody.querySelectorAll('tr');

    for (let row of rows) {
        const taskIdCell = row.querySelector('td');
        if (taskIdCell && taskIdCell.textContent.includes(taskId.substring(0, 8))) {
            const durationCell = row.querySelectorAll('td')[4]; // 用时列
            if (durationCell) {
                durationCell.textContent = formatRunningTime(startedAt) + ' (进行中)';
            }
            break;
        }
    }
}

// 格式化时间
function formatTime(timeString) {
    if (!timeString) return '-';
    const date = new Date(timeString);
    return date.toLocaleString('zh-CN');
}

// 计算用时
function formatDuration(startTime, endTime) {
    if (!startTime) return '-';
    if (!endTime) return '进行中...';

    const start = new Date(startTime).getTime();
    const end = new Date(endTime).getTime();
    const diff = Math.floor((end - start) / 1000); // 秒

    if (diff < 60) {
        return `${diff}秒`;
    } else if (diff < 3600) {
        const minutes = Math.floor(diff / 60);
        const seconds = diff % 60;
        return `${minutes}分${seconds}秒`;
    } else {
        const hours = Math.floor(diff / 3600);
        const minutes = Math.floor((diff % 3600) / 60);
        return `${hours}小时${minutes}分钟`;
    }
}

// 获取状态标签样式
function getStatusClass(status) {
    return `status status-${status}`;
}

// 获取任务类型显示文本
function getTaskTypeLabel(type) {
    switch (type) {
        case 'transcode':
            return '<span style="display: inline-flex; align-items: center; gap: 6px; padding: 6px 12px; background: rgba(239, 68, 68, 0.05); color: #dc2626; border-radius: 20px; font-size: 12px; font-weight: 600;">🎬 转码</span>';
        case 'command':
            return '<span style="display: inline-flex; align-items: center; gap: 6px; padding: 6px 12px; background: rgba(59, 130, 246, 0.05); color: #2563eb; border-radius: 20px; font-size: 12px; font-weight: 600;">⚡ 命令</span>';
        default:
            return type;
    }
}

// 格式化码率
function formatBitrate(bitrate) {
    if (!bitrate) return '默认';
    return bitrate;
}

// 格式化命令（过长时截断）
function formatCommand(command) {
    if (!command) return '-';
    if (command.length > 30) {
        return command.substring(0, 30) + '...';
    }
    return command;
}

// 显示消息
function showMessage(message, type = 'error') {
    const messageEl = document.getElementById('message');
    messageEl.innerHTML = `<div class="message ${type} show">${message}</div>`;
    setTimeout(() => {
        messageEl.innerHTML = '';
    }, 5000);
}

// 切换任务类型（Tab）
function switchTaskType(type) {
    // 更新Tab激活状态
    document.querySelectorAll('.tab').forEach(tab => {
        tab.classList.remove('active');
    });
    document.querySelector(`[data-tab="${type}"]`).classList.add('active');

    // 隐藏所有tab内容
    document.querySelectorAll('.tab-content').forEach(content => {
        content.classList.remove('active');
    });

    // 显示相应tab内容
    document.getElementById(`tab-${type}`).classList.add('active');

    // 更新表单标题
    const title = document.getElementById('form-title');
    if (type === 'transcode') {
        title.textContent = '🎬 创建视频转码任务';
    } else {
        title.textContent = '⚡ 创建命令执行任务';
    }

    // 更新提交按钮
    const submitBtn = document.getElementById('submit-btn');
    submitBtn.textContent = type === 'transcode' ? '添加转码任务' : '执行命令';

    // 更新必填字段
    const inputFile = document.getElementById('input-file');
    if (type === 'transcode') {
        inputFile.setAttribute('required', 'required');
    } else {
        inputFile.removeAttribute('required');
    }

    const form = document.getElementById('task-form');
    form.scrollIntoView({ behavior: 'smooth', block: 'start' });
}

// 加载任务列表
async function loadTasks(page = 1) {
    const tbody = document.getElementById('task-list');

    try {
        // 获取任务列表（带分页参数）
        const limit = 20; // 每页20个
        const response = await fetch(`${API_BASE}/tasks?page=${page}&limit=${limit}`);
        if (!response.ok) throw new Error('获取任务列表失败');

        const tasks = await response.json();

        // 存储当前任务列表（用于全选功能）
        currentTasks = tasks;

        // 获取统计信息
        const statsResponse = await fetch(`${API_BASE}/stats`);
        if (!statsResponse.ok) throw new Error('获取统计信息失败');

        const stats = await statsResponse.json();

        // 更新统计
        document.getElementById('total-tasks').textContent = stats.total_tasks || 0;
        document.getElementById('pending-tasks').textContent = stats.status_count?.pending || 0;
        document.getElementById('running-tasks').textContent = stats.status_count?.running || 0;
        document.getElementById('completed-tasks').textContent = stats.status_count?.completed || 0;

        // 更新并发数输入框
        if (stats.max_concurrency) {
            document.getElementById('concurrency').value = stats.max_concurrency;
        }

        // 更新任务列表
        if (tasks.length === 0) {
            tbody.innerHTML = '<tr><td colspan="7" style="text-align: center; color: #999;">暂无任务</td></tr>';
            // 清空所有定时器
            Object.keys(runningTaskTimers).forEach(taskId => {
                clearInterval(runningTaskTimers[taskId]);
                delete runningTaskTimers[taskId];
            });
            // 清空选中的任务并更新UI
            selectedTasks.clear();
            updateBatchDeleteUI();
            return;
        }

        // 清理已不存在或已完成的任务的定时器
        const currentRunningTaskIds = tasks.filter(t => t.status === 'running').map(t => t.id);
        Object.keys(runningTaskTimers).forEach(taskId => {
            if (!currentRunningTaskIds.includes(taskId)) {
                clearInterval(runningTaskTimers[taskId]);
                delete runningTaskTimers[taskId];
            }
        });

        tbody.innerHTML = tasks.map(task => {
            let durationText = '';
            if (task.status === 'running') {
                // 对于运行中的任务，显示实时计时
                durationText = formatRunningTime(task.started_at) + ' (进行中)';

                // 如果这个任务还没有定时器，创建一个
                if (!runningTaskTimers[task.id]) {
                    runningTaskTimers[task.id] = setInterval(() => {
                        updateTaskRunningTime(task.id, task.started_at);
                    }, 1000);
                }
            } else {
                // 对于非运行中的任务，显示总用时
                durationText = formatDuration(task.started_at, task.completed_at);
            }

            // 根据任务状态生成操作按钮
            let actionButtons = '';
            // 查看按钮始终显示
            actionButtons += `<button onclick="showTaskDetails('${task.id}')" class="btn btn-sm btn-secondary">👁️ 查看</button>`;

            // 取消按钮：仅对 pending 和 running 状态显示
            if (task.status === 'pending' || task.status === 'running') {
                actionButtons += `<button onclick="cancelTask('${task.id}')" class="btn btn-sm btn-warning">⏹️ 取消</button>`;
            }

            // 删除按钮：仅对非 running 状态显示
            if (task.status !== 'running') {
                actionButtons += `<button onclick="deleteTask('${task.id}')" class="btn btn-sm btn-danger">🗑️ 删除</button>`;
            }

            // 复选框：仅对非 running 状态显示
            const checkboxHtml = task.status !== 'running'
                ? `<input type="checkbox" onchange="toggleTaskSelection('${task.id}')" ${selectedTasks.has(task.id) ? 'checked' : ''}>`
                : '';

            return `
                <tr>
                    <td style="text-align: center;">${checkboxHtml}</td>
                    <td>${task.id.substring(0, 8)}...</td>
                    <td>${getTaskTypeLabel(task.task_type)}</td>
                    <td><span class="${getStatusClass(task.status)}">${task.status}</span></td>
                    <td>${formatTime(task.started_at)}</td>
                    <td>${durationText}</td>
                    <td>${actionButtons}</td>
                </tr>
            `;
        }).join('');

        // 更新批量删除UI
        updateBatchDeleteUI();

    } catch (error) {
        console.error('加载任务失败:', error);
        tbody.innerHTML = '<tr><td colspan="7" style="text-align: center; color: #f00;">加载失败，请刷新重试</td></tr>';
        // 清空所有定时器
        Object.keys(runningTaskTimers).forEach(taskId => {
            clearInterval(runningTaskTimers[taskId]);
            delete runningTaskTimers[taskId];
        });
        // 清空选中的任务并更新UI
        selectedTasks.clear();
        updateBatchDeleteUI();
    }
}


// 提交表单
document.getElementById('task-form').addEventListener('submit', async (e) => {
    e.preventDefault();

    // 获取当前激活的tab
    const activeTab = document.querySelector('.tab.active');
    const taskType = activeTab.getAttribute('data-tab');

    try {
        let requestBody = {
            task_type: taskType
        };

        if (taskType === 'transcode') {
            const inputFile = document.getElementById('input-file').value.trim();
            const bitrate = document.getElementById('bitrate').value.trim();

            if (!inputFile) {
                showMessage('请输入输入文件名', 'error');
                return;
            }

            // 验证文件名不包含路径分隔符
            if (inputFile.includes('/') || inputFile.includes('\\') || inputFile.includes('..')) {
                showMessage('文件名不能包含路径，请只提供文件名（例如: a.mp4）', 'error');
                return;
            }

            // 如果输入了码率，添加 'k' 单位（例如 1000 -> 1000k）
            requestBody.input_file = inputFile;
            requestBody.bitrate = bitrate ? `${bitrate}k` : undefined;

        } else if (taskType === 'command') {
            const command = document.getElementById('command').value.trim();
            const timeout = document.getElementById('timeout').value;

            if (!command) {
                showMessage('请输入要执行的命令', 'error');
                return;
            }

            requestBody.command = command;
            requestBody.timeout = timeout ? parseInt(timeout) : 300; // 默认300秒
        }

        const response = await fetch(`${API_BASE}/tasks`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(requestBody),
        });

        if (!response.ok) {
            const error = await response.text();
            throw new Error(error);
        }

        const task = await response.json();
        const taskTypeLabel = taskType === 'transcode' ? '转码任务' : '命令任务';
        showMessage(`${taskTypeLabel}添加成功! ID: ${task.id.substring(0, 8)}`, 'success');

        // 清空表单
        document.getElementById('input-file').value = '';
        document.getElementById('bitrate').value = '';
        document.getElementById('command').value = '';
        document.getElementById('timeout').value = '300';

        // 刷新任务列表
        loadTasks();

    } catch (error) {
        console.error('添加任务失败:', error);
        showMessage(`添加任务失败: ${error.message}`, 'error');
    }
});

// 显示任务详情
async function showTaskDetails(taskId) {
    const modal = document.getElementById('task-modal');
    const title = document.getElementById('modal-title');
    const content = document.getElementById('modal-content');
    const deleteBtn = document.getElementById('modal-delete-btn');

    try {
        const response = await fetch(`${API_BASE}/tasks/${taskId}`);
        if (!response.ok) {
            throw new Error('获取任务详情失败');
        }

        const task = await response.json();

        title.textContent = `任务详情: ${task.id.substring(0, 8)}...`;

        let details = '';
        details += `任务ID: ${task.id}\n`;
        details += `任务类型: ${getTaskTypeLabel(task.task_type)}\n`;
        details += `状态: ${task.status}\n`;

        if (task.task_type === 'transcode') {
            details += `\n输入文件: ${task.input_file || '-'}\n`;
            details += `输出文件: ${task.output_file || '-'}\n`;
            details += `码率: ${task.bitrate || '默认'}\n`;
        } else {
            details += `\n命令: ${task.command || '-'}\n`;
            details += `超时时间: ${task.timeout}秒\n`;
        }

        details += `\n创建时间: ${formatTime(task.created_at)}\n`;
        details += `开始时间: ${formatTime(task.started_at)}\n`;
        details += `完成时间: ${formatTime(task.completed_at)}\n`;
        details += `用时: ${formatDuration(task.started_at, task.completed_at)}\n`;

        if (task.error) {
            details += `\n错误信息:\n${task.error}\n`;
        }

        if (task.logs) {
            details += `\n日志输出:\n${task.logs}\n`;
        }

        content.textContent = details;

        // 根据任务状态显示/隐藏删除按钮
        if (task.status !== 'running') {
            deleteBtn.style.display = 'inline-block';
            deleteBtn.onclick = () => {
                closeModal();
                deleteTask(taskId);
            };
        } else {
            deleteBtn.style.display = 'none';
        }

        modal.classList.add('visible');
    } catch (error) {
        console.error('获取任务详情失败:', error);
        showMessage(error.message, 'error');
    }
}

// 关闭模态框
function closeModal() {
    document.getElementById('task-modal').classList.remove('visible');
}

// 取消任务
async function cancelTask(taskId) {
    if (!confirm(`确定要取消任务 ${taskId.substring(0, 8)}... 吗？`)) {
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/tasks/${taskId}/cancel`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
        });

        if (!response.ok) {
            const error = await response.text();
            throw new Error(error);
        }

        const result = await response.json();
        showMessage(result.message, 'success');

        // 刷新任务列表
        loadTasks();

    } catch (error) {
        console.error('取消任务失败:', error);
        showMessage(`取消任务失败: ${error.message}`, 'error');
    }
}

// 删除任务
async function deleteTask(taskId) {
    if (!confirm(`确定要删除任务 ${taskId.substring(0, 8)}... 吗？此操作不可恢复。`)) {
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/tasks/${taskId}`, {
            method: 'DELETE',
            headers: {
                'Content-Type': 'application/json',
            },
        });

        if (!response.ok) {
            const error = await response.text();
            throw new Error(error);
        }

        const result = await response.json();
        showMessage(result.message, 'success');

        // 刷新任务列表
        loadTasks();

    } catch (error) {
        console.error('删除任务失败:', error);
        showMessage(`删除任务失败: ${error.message}`, 'error');
    }
}

// 打开系统设置模态框
function openSettingsModal() {
    // 从主页面同步当前值
    const concurrencyInput = document.getElementById('concurrency');
    const concurrencyModal = document.getElementById('concurrency-modal');
    if (concurrencyInput && concurrencyModal) {
        concurrencyModal.value = concurrencyInput.value;
    }
    document.getElementById('settings-modal').classList.add('visible');
}

// 关闭系统设置模态框
function closeSettingsModal() {
    document.getElementById('settings-modal').classList.remove('visible');
}

// 保存并发数
async function saveConcurrency() {
    const concurrency = document.getElementById('concurrency-modal').value;

    if (!concurrency || concurrency < 1 || concurrency > 1000) {
        showMessage('请输入有效的并发数（1-1000）', 'error');
        return;
    }

    // 同时更新主页面和模态框中的值
    const mainConcurrencyInput = document.getElementById('concurrency');
    if (mainConcurrencyInput) {
        mainConcurrencyInput.value = concurrency;
    }

    try {
        const response = await fetch(`${API_BASE}/concurrency`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ max_concurrency: parseInt(concurrency) }),
        });

        if (!response.ok) {
            const error = await response.text();
            throw new Error(error);
        }

        const result = await response.json();
        showMessage(result.message, 'success');
        closeSettingsModal();

    } catch (error) {
        console.error('更新并发数失败:', error);
        showMessage(`更新并发数失败: ${error.message}`, 'error');
    }
}

// 切换全选/取消全选
function toggleSelectAll() {
    const selectAllCheckbox = document.getElementById('select-all');

    if (selectAllCheckbox.checked) {
        // 全选：将所有非running状态的任务添加到selectedTasks
        currentTasks.forEach(task => {
            if (task.status !== 'running') {
                selectedTasks.add(task.id);
            }
        });
    } else {
        // 取消全选：从selectedTasks中移除当前页面的所有任务
        currentTasks.forEach(task => {
            selectedTasks.delete(task.id);
        });
    }

    // 立即更新复选框状态，而不是重新加载整个列表
    const tbody = document.getElementById('task-list');
    const checkboxes = tbody.querySelectorAll('input[type="checkbox"]');
    checkboxes.forEach((checkbox, index) => {
        if (index > 0) { // 跳过全选复选框本身
            const row = checkbox.closest('tr');
            const statusCell = row.querySelectorAll('td')[3];
            if (statusCell) {
                const status = statusCell.querySelector('span');
                if (status && status.textContent !== 'running') {
                    checkbox.checked = selectAllCheckbox.checked;
                }
            }
        }
    });

    updateBatchDeleteUI();
}

// 切换单个任务的选择状态
function toggleTaskSelection(taskId) {
    if (selectedTasks.has(taskId)) {
        selectedTasks.delete(taskId);
    } else {
        selectedTasks.add(taskId);
    }

    updateBatchDeleteUI();
}

// 更新批量删除UI（显示已选择数量和批量删除按钮）
function updateBatchDeleteUI() {
    const selectedCount = document.getElementById('selected-count');
    const batchDeleteBtn = document.getElementById('batch-delete-btn');
    const selectAllCheckbox = document.getElementById('select-all');

    const count = selectedTasks.size;

    if (count > 0) {
        selectedCount.style.display = 'inline';
        selectedCount.textContent = `已选择 ${count} 项`;
        batchDeleteBtn.style.display = 'inline-block';

        // 检查是否全选：检查currentTasks中所有非running任务是否都被选中
        let allCheckableSelected = true;
        let hasCheckableTask = false;

        for (const task of currentTasks) {
            if (task.status !== 'running') {
                hasCheckableTask = true;
                if (!selectedTasks.has(task.id)) {
                    allCheckableSelected = false;
                    break;
                }
            }
        }

        // 只有当有可选任务且所有可选任务都被选中时，才勾选全选框
        selectAllCheckbox.checked = hasCheckableTask && allCheckableSelected;
    } else {
        selectedCount.style.display = 'none';
        selectedCount.textContent = '已选择 0 项';
        batchDeleteBtn.style.display = 'none';
        selectAllCheckbox.checked = false;
    }
}

// 批量删除选中的任务
async function batchDeleteTasks() {
    if (selectedTasks.size === 0) {
        showMessage('请先选择要删除的任务', 'error');
        return;
    }

    const taskIds = Array.from(selectedTasks);
    const shortIds = taskIds.map(id => id.substring(0, 8)).join(', ');

    if (!confirm(`确定要删除选中的 ${taskIds.length} 个任务吗？此操作不可恢复。\n\n任务: ${shortIds}`)) {
        return;
    }

    const batchDeleteBtn = document.getElementById('batch-delete-btn');
    batchDeleteBtn.disabled = true;
    batchDeleteBtn.textContent = '删除中...';

    let successCount = 0;
    let failCount = 0;

    for (const taskId of taskIds) {
        try {
            const response = await fetch(`${API_BASE}/tasks/${taskId}`, {
                method: 'DELETE',
                headers: {
                    'Content-Type': 'application/json',
                },
            });

            if (!response.ok) {
                const error = await response.text();
                console.error(`删除任务 ${taskId} 失败:`, error);
                failCount++;
            } else {
                successCount++;
                selectedTasks.delete(taskId); // 删除成功后从选中列表中移除
            }
        } catch (error) {
            console.error(`删除任务 ${taskId} 失败:`, error);
            failCount++;
        }
    }

    // 恢复按钮状态
    batchDeleteBtn.disabled = false;
    batchDeleteBtn.textContent = '🗑️ 批量删除';

    // 显示结果
    if (failCount === 0) {
        showMessage(`成功删除 ${successCount} 个任务`, 'success');
    } else {
        showMessage(`删除结果: 成功 ${successCount} 个, 失败 ${failCount} 个`, 'error');
    }

    // 刷新任务列表和UI
    loadTasks();
    updateBatchDeleteUI();
}


// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', () => {
    // 初始化任务类型（默认转码）
    switchTaskType('transcode');

    // 监听Tab点击事件
    document.querySelectorAll('.tab').forEach(tab => {
        tab.addEventListener('click', () => {
            const tabType = tab.getAttribute('data-tab');
            switchTaskType(tabType);
        });
    });

    // 点击模态框外部关闭
    document.getElementById('task-modal').addEventListener('click', (e) => {
        if (e.target.id === 'task-modal') {
            closeModal();
        }
    });

    // 点击系统设置模态框外部关闭
    document.getElementById('settings-modal').addEventListener('click', (e) => {
        if (e.target.id === 'settings-modal') {
            closeSettingsModal();
        }
    });

    // 加载任务列表
    loadTasks();

    // 每5秒自动刷新任务列表
    setInterval(loadTasks, 5000);
});
