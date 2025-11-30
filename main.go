package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"video-transcoder/config"
	"video-transcoder/handler"
	"video-transcoder/service"

	"github.com/gorilla/mux"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 打印配置信息
	fmt.Printf("加载配置:\n")
	fmt.Printf("  输入目录: %s\n", cfg.InputDir)
	fmt.Printf("  输出目录: %s\n", cfg.OutputDir)
	fmt.Printf("  并发数: %d\n", cfg.MaxConcurrency)
	fmt.Printf("  Docker Host: %s\n", cfg.DockerHost)
	fmt.Println()

	// 检查输入输出目录是否存在
	if _, err := os.Stat(cfg.InputDir); os.IsNotExist(err) {
		fmt.Printf("警告: 输入目录不存在: %s\n", cfg.InputDir)
	}
	if _, err := os.Stat(cfg.OutputDir); os.IsNotExist(err) {
		fmt.Printf("警告: 输出目录不存在: %s\n", cfg.OutputDir)
	}

	// 初始化任务服务
	taskService, err := service.NewTaskService(cfg)
	if err != nil {
		log.Fatalf("初始化任务服务失败: %v", err)
	}
	defer taskService.Shutdown()

	// 初始化HTTP处理器
	taskHandler := handler.NewTaskHandler(taskService)

	// 设置路由
	r := mux.NewRouter()

	// API路由
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/tasks", taskHandler.CreateTask).Methods("POST")
	api.HandleFunc("/tasks", taskHandler.ListTasks).Methods("GET")
	api.HandleFunc("/tasks/{id}", taskHandler.GetTask).Methods("GET")
	api.HandleFunc("/tasks/{id}", taskHandler.DeleteTask).Methods("DELETE")
	api.HandleFunc("/tasks/{id}/cancel", taskHandler.CancelTask).Methods("POST")
	api.HandleFunc("/stats", taskHandler.GetStats).Methods("GET")
	api.HandleFunc("/concurrency", taskHandler.UpdateConcurrency).Methods("POST")
	api.HandleFunc("/concurrency", taskHandler.GetStats).Methods("GET")

	// 静态文件和首页
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	r.HandleFunc("/", taskHandler.ServeStatic)

	// 启动HTTP服务器
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// 优雅关闭
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP服务器启动失败: %v", err)
		}
	}()

	fmt.Println("视频转码管理系统启动成功!")
	fmt.Println("访问地址: http://localhost:8080")
	fmt.Printf("当前并发数: %d\n", cfg.MaxConcurrency)

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\n正在关闭服务...")

	// 优雅关闭HTTP服务器
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("HTTP服务器关闭失败: %v", err)
	}

	fmt.Println("服务已停止")
}
