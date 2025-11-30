package service

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
	"video-transcoder/config"
	"video-transcoder/models"
)

type DockerExecutor struct {
	cfg *config.Config
}

func NewDockerExecutor(cfg *config.Config) *DockerExecutor {
	return &DockerExecutor{
		cfg: cfg,
	}
}

func (de *DockerExecutor) getDockerCmd(args ...string) *exec.Cmd {
	cmdArgs := []string{}
	if de.cfg.DockerHost != "" {
		cmdArgs = append(cmdArgs, "-H", de.cfg.DockerHost)
	}
	cmdArgs = append(cmdArgs, args...)
	return exec.Command("docker", cmdArgs...)
}

func (de *DockerExecutor) RunTranscodeTask(ctx context.Context, task *models.Task) error {
	// 准备docker run命令
	// 挂载宿主机的输入和输出目录到ffmpeg容器
	args := []string{
		"run", "--rm",
		"-v", fmt.Sprintf("%s:/input", de.cfg.InputDir),   // 挂载宿主机输入目录
		"-v", fmt.Sprintf("%s:/output", de.cfg.OutputDir), // 挂载宿主机输出目录
		"--name", de.getContainerName(task.ID),
	}

	// 获取视频编码器配置
	bitrate := task.Bitrate
	if bitrate == "" {
		bitrate = "1M" // 默认1Mbps
	}
	videoEncoder := GetVideoEncoder(de.cfg, bitrate)

	// 获取音频编码器配置
	audioEncoder := GetAudioEncoder(de.cfg)

	// 构建FFmpeg参数
	inputFile := fmt.Sprintf("/input/%s", task.InputFile)
	outputFile := fmt.Sprintf("/output/%s", task.OutputFile)
	ffmpegArgs := BuildFFmpegArgs(inputFile, outputFile, videoEncoder, audioEncoder)

	// 添加Docker运行参数
	args = append(args, ffmpegArgs...)

	cmd := de.getDockerCmd(args...)

	// 创建 buffer 来捕获日志
	var outputBuf, errorBuf strings.Builder
	cmd.Stdout = &outputBuf
	cmd.Stderr = &errorBuf

	task.Status = models.StatusRunning
	now := time.Now()
	task.StartedAt = &now

	fmt.Printf("开始转码任务: %s\n", task.ID)
	fmt.Printf("命令: docker %s\n", strings.Join(args, " "))

	// 打印到控制台（同时显示）
	fmt.Printf("=== 任务 %s 输出 ===\n", task.ID)

	err := cmd.Run()

	// 保存日志（stdout + stderr）
	task.Logs = outputBuf.String() + "\n" + errorBuf.String()

	if err != nil {
		task.Status = models.StatusFailed
		task.Error = err.Error()
		completedTime := time.Now()
		task.CompletedAt = &completedTime
		fmt.Printf("=== 任务 %s 失败 ===\n", task.ID)
		return fmt.Errorf("转码失败: %w", err)
	}

	task.Status = models.StatusCompleted
	completedTime := time.Now()
	task.CompletedAt = &completedTime

	fmt.Printf("=== 任务 %s 完成 ===\n", task.ID)
	return nil
}

func (de *DockerExecutor) RunCommandTask(ctx context.Context, task *models.Task) error {
	// 准备docker run命令
	// 挂载宿主机的输入和输出目录到容器
	args := []string{
		"run", "--rm",
		"-v", fmt.Sprintf("%s:/input", de.cfg.InputDir),   // 挂载宿主机输入目录
		"-v", fmt.Sprintf("%s:/output", de.cfg.OutputDir), // 挂载宿主机输出目录
		"--name", de.getContainerName(task.ID),
		"--workdir", "/workspace",
	}

	// 使用Alpine Linux作为基础镜像，体积小且包含常用命令
	// 用户需要在命令中指定完整的工作目录路径
	args = append(args, "alpine:latest", "sh", "-c", task.Command)

	cmd := de.getDockerCmd(args...)

	// 创建 buffer 来捕获日志
	var outputBuf, errorBuf strings.Builder
	cmd.Stdout = &outputBuf
	cmd.Stderr = &errorBuf

	task.Status = models.StatusRunning
	now := time.Now()
	task.StartedAt = &now

	// 设置超时时间（默认300秒）
	timeoutSeconds := task.Timeout
	if timeoutSeconds == 0 {
		timeoutSeconds = 300
	}

	// 使用带超时的context
	ctxWithTimeout, cancel := context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	fmt.Printf("开始命令任务: %s (超时: %d秒)\n", task.ID, timeoutSeconds)
	fmt.Printf("命令: docker %s\n", strings.Join(args, " "))

	// 打印到控制台
	fmt.Printf("=== 任务 %s 输出 ===\n", task.ID)

	// 启动命令
	if err := cmd.Start(); err != nil {
		task.Status = models.StatusFailed
		task.Error = fmt.Sprintf("启动命令失败: %v", err)
		fmt.Printf("=== 任务 %s 启动失败 ===\n", task.ID)
		return err
	}

	// 等待命令完成或超时
	ch := make(chan error, 1)
	go func() {
		ch <- cmd.Wait()
	}()

	select {
	case err := <-ch:
		// 保存日志
		task.Logs = outputBuf.String() + "\n" + errorBuf.String()

		if err != nil {
			task.Status = models.StatusFailed
			task.Error = err.Error()
			completedTime := time.Now()
			task.CompletedAt = &completedTime
			fmt.Printf("=== 任务 %s 失败 ===\n", task.ID)
			return fmt.Errorf("命令执行失败: %w", err)
		}

		task.Status = models.StatusCompleted
		completedTime := time.Now()
		task.CompletedAt = &completedTime

		fmt.Printf("=== 任务 %s 完成 ===\n", task.ID)
		return nil

	case <-ctxWithTimeout.Done():
		// 超时
		fmt.Printf("=== 任务 %s 超时（%d秒）===\n", task.ID, timeoutSeconds)

		// 停止容器
		stopErr := de.StopContainer(de.getContainerName(task.ID))
		if stopErr != nil {
			fmt.Printf("停止超时容器失败: %v\n", stopErr)
		}

		// 等待容器完全停止
		time.Sleep(2 * time.Second)

		// 保存日志
		task.Logs = outputBuf.String() + "\n" + errorBuf.String() + "\n\n=== 任务超时 ==="

		task.Status = models.StatusFailed
		task.Error = fmt.Sprintf("命令执行超时（%d秒）", timeoutSeconds)
		completedTime := time.Now()
		task.CompletedAt = &completedTime

		return fmt.Errorf("命令执行超时（%d秒）", timeoutSeconds)
	}
}

func (de *DockerExecutor) getContainerName(taskID string) string {
	return fmt.Sprintf("transcode-%s", taskID)
}

func (de *DockerExecutor) StopContainer(containerName string) error {
	fmt.Printf("停止容器: %s\n", containerName)

	// 首先尝试停止容器
	stopCmd := de.getDockerCmd("stop", containerName)
	stopCmd.Stdout = os.Stdout
	stopCmd.Stderr = os.Stderr

	if err := stopCmd.Run(); err != nil {
		// 容器可能已经停止或不存在，这不一定是错误
		fmt.Printf("停止容器 %s 失败（可能已停止）: %v\n", containerName, err)
	}

	// 等待一小段时间确保容器停止
	time.Sleep(500 * time.Millisecond)

	// 然后尝试删除容器（如果还存在）
	rmCmd := de.getDockerCmd("rm", "-f", containerName)
	rmCmd.Stdout = os.Stdout
	rmCmd.Stderr = os.Stderr

	if err := rmCmd.Run(); err != nil {
		fmt.Printf("删除容器 %s 失败: %v\n", containerName, err)
		return err
	}

	fmt.Printf("容器 %s 已成功停止并删除\n", containerName)
	return nil
}
