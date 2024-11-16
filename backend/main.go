package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend/services/admin_service/src/config"
	"backend/services/admin_service/src/momo"
	"backend/services/admin_service/src/routes"
	"backend/services/admin_service/src/utils"

	"github.com/joho/godotenv"
)

func main() {
	// Nạp file .env vào môi trường
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Kết nối MongoDB với retry
	maxRetries := 3
	retryDelay := 5 * time.Second
	if err := config.ConnectDatabaseWithRetry(maxRetries, retryDelay); err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	// Đảm bảo ngắt kết nối khi server dừng
	defer config.DisconnectDatabase()

	// Bắt đầu routine dọn dẹp token hết hạn
	utils.StartCleanupRoutine()
	r := routes.SetupRouter()

	// Gọi hàm init trong package momo để khởi tạo các giá trị cần thiết
	momo.Init()
	//r.GET("/blacklisted-tokens", utils.ListBlacklistedTokens)

	// Bắt tín hiệu tắt server để thực hiện cleanup
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Chạy server trong một goroutine riêng
	go func() {
		log.Println("Server is running...")
		if err := r.Run(":8082"); err != nil {
			log.Printf("Server exited: %v", err)
		}
	}()

	// Chờ tín hiệu tắt từ hệ thống
	<-quit
	log.Println("Shutting down server...")

	log.Println("Server gracefully stopped.")

	// go func() {
	//     <-quit
	//     log.Println("Shutting down server...")
	//     os.Exit(0)
	// }()

	// // Chạy server tại cổng 8082
	// if err := r.Run(":8082"); err != nil {
	//     log.Fatalf("Server encountered an error: %v", err)
	// }

}
